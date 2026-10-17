// Package media terminates browser WebRTC sessions on the Pi (README §3.2)
// and forwards the on-air client's audio and video to sinks that play them
// into the UAC2 and UVC gadget functions (README §3.1). Only one client is on
// air at a time; the others stay connected so a switch is instant (README §3.3).
package media

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// Sink receives raw RTP packets from the on-air client. Packets arrive with
// their payload type rewritten to the sink's fixed one (AudioPT, VideoPT),
// whatever the browser negotiated.
type Sink interface {
	WriteRTP(pkt []byte) error
}

// AudioSink receives Opus RTP.
type AudioSink = Sink

// Fixed RTP payload types the sinks' pipelines expect. The video sink
// switches decoder by payload type.
const (
	AudioPT = 111 // Opus
	VideoPT = 96  // H.264
	VP8PT   = 97  // VP8, for browsers that cannot send H.264
)

// pliInterval is how often the on-air video sender is asked for a keyframe,
// so a restarted decoder recovers without waiting for a switch.
const pliInterval = 10 * time.Second

// Options configures a Bridge.
type Options struct {
	// DisableMDNS turns off resolving browsers' obfuscated .local ICE
	// candidates. Tests use it; on a Pi it must stay off.
	DisableMDNS bool
	// Video receives the on-air client's H.264 (VideoPT) or VP8 (VP8PT) RTP.
	// Nil ignores video.
	Video Sink
}

// Bridge holds one PeerConnection per client.
type Bridge struct {
	api   *webrtc.API
	sink  AudioSink
	video Sink

	mu    sync.Mutex
	peers map[core.ClientID]*webrtc.PeerConnection
	vssrc map[core.ClientID]uint32 // each client's video SSRC, for keyframe requests
	onAir core.ClientID
	live  bool

	forwarded      atomic.Uint64
	forwardedVideo atomic.Uint64
}

// New builds a Bridge that accepts Opus audio and writes the on-air
// client's packets to sink.
func New(sink AudioSink, opts Options) (*Bridge, error) {
	me := &webrtc.MediaEngine{}
	if err := me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1"},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}
	// H.264 constrained baseline, packetization-mode 1: what every browser
	// sends and the Pi 4 decodes in hardware (README §4).
	for i, profile := range []string{"42e01f", "42001f"} {
		if err := me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=" + profile,
				RTCPFeedback: []webrtc.RTCPFeedback{{Type: "nack"}, {Type: "nack", Parameter: "pli"}, {Type: "ccm", Parameter: "fir"}}},
			PayloadType: webrtc.PayloadType(VideoPT + 2*i),
		}, webrtc.RTPCodecTypeVideo); err != nil {
			return nil, err
		}
	}
	// VP8 last: every browser can send it, but the Pi decodes it in software.
	if err := me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000,
			RTCPFeedback: []webrtc.RTCPFeedback{{Type: "nack"}, {Type: "nack", Parameter: "pli"}, {Type: "ccm", Parameter: "fir"}}},
		PayloadType: VP8PT + 3,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}
	var se webrtc.SettingEngine
	if opts.DisableMDNS {
		se.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	} else {
		se.SetICEMulticastDNSMode(ice.MulticastDNSModeQueryOnly)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithSettingEngine(se))
	return &Bridge{api: api, sink: sink, video: opts.Video, peers: map[core.ClientID]*webrtc.PeerConnection{}, vssrc: map[core.ClientID]uint32{}}, nil
}

// Offer answers a client's SDP offer. ICE is not trickled: the answer carries
// every Pi candidate, and the offer must carry the browser's.
func (b *Bridge) Offer(id core.ClientID, sdp string) (string, error) {
	if id.IsZero() {
		return "", errors.New("media: offer without a client")
	}
	pc, err := b.api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}
	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		switch {
		case t.Kind() == webrtc.RTPCodecTypeAudio:
			b.pump(id, t, b.sink, AudioPT, &b.forwarded)
		case t.Kind() == webrtc.RTPCodecTypeVideo && b.video != nil:
			b.mu.Lock()
			b.vssrc[id] = uint32(t.SSRC())
			b.mu.Unlock()
			b.requestKeyframe(id)
			stop := make(chan struct{})
			go b.keyframeTicker(id, stop)
			pt := byte(VideoPT)
			if strings.EqualFold(t.Codec().MimeType, webrtc.MimeTypeVP8) {
				pt = VP8PT
			}
			b.pump(id, t, b.video, pt, &b.forwardedVideo)
			close(stop)
		}
	})
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp}); err != nil {
		pc.Close()
		return "", fmt.Errorf("media: bad offer: %w", err)
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		pc.Close()
		return "", fmt.Errorf("media: answer: %w", err)
	}
	gathered := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(answer); err != nil {
		pc.Close()
		return "", err
	}
	<-gathered

	b.mu.Lock()
	old := b.peers[id]
	b.peers[id] = pc
	b.mu.Unlock()
	if old != nil {
		old.Close()
	}
	return pc.LocalDescription().SDP, nil
}

// pump forwards a track's packets to sink while its client is on air.
func (b *Bridge) pump(id core.ClientID, t *webrtc.TrackRemote, sink Sink, pt byte, count *atomic.Uint64) {
	buf := make([]byte, 1500)
	for {
		n, _, err := t.Read(buf)
		if err != nil {
			return
		}
		if !b.isOnAir(id) || n < 12 {
			continue
		}
		buf[1] = buf[1]&0x80 | pt // keep the marker bit
		if sink.WriteRTP(buf[:n]) == nil {
			count.Add(1)
		}
	}
}

func (b *Bridge) isOnAir(id core.ClientID) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.live && b.onAir == id
}

// requestKeyframe asks id's video sender for a keyframe (RTCP PLI) if it is
// on air, so the decoder can start on a clean picture.
func (b *Bridge) requestKeyframe(id core.ClientID) {
	b.mu.Lock()
	pc, ssrc, on := b.peers[id], b.vssrc[id], b.live && b.onAir == id
	b.mu.Unlock()
	if pc == nil || ssrc == 0 || !on {
		return
	}
	_ = pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: ssrc}})
}

func (b *Bridge) keyframeTicker(id core.ClientID, stop <-chan struct{}) {
	t := time.NewTicker(pliInterval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			b.requestKeyframe(id)
		}
	}
}

// SetOnAir selects whose media reaches the sinks; live=false mutes all.
func (b *Bridge) SetOnAir(id core.ClientID, live bool) {
	b.mu.Lock()
	changed := b.onAir != id || b.live != live
	b.onAir, b.live = id, live
	b.mu.Unlock()
	if changed && live {
		b.requestKeyframe(id)
	}
}

// Forwarded counts audio packets handed to the sink.
func (b *Bridge) Forwarded() uint64 { return b.forwarded.Load() }

// ForwardedVideo counts video packets handed to the video sink.
func (b *Bridge) ForwardedVideo() uint64 { return b.forwardedVideo.Load() }

// Drop closes a client's connection, if any.
func (b *Bridge) Drop(id core.ClientID) {
	b.mu.Lock()
	pc := b.peers[id]
	delete(b.peers, id)
	delete(b.vssrc, id)
	b.mu.Unlock()
	if pc != nil {
		pc.Close()
	}
}

// Close closes every connection.
func (b *Bridge) Close() {
	b.mu.Lock()
	peers := b.peers
	b.peers = map[core.ClientID]*webrtc.PeerConnection{}
	b.mu.Unlock()
	for _, pc := range peers {
		pc.Close()
	}
}
