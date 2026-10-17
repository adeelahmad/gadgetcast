// Package media terminates browser WebRTC sessions on the Pi (README §3.2)
// and forwards the on-air client's audio to an AudioSink, which plays it into
// the UAC2 gadget (README §3.1). Only one client is on air at a time; the
// others stay connected so a switch is instant (README §3.3).
package media

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// AudioSink receives raw Opus RTP packets from the on-air client.
type AudioSink interface {
	WriteRTP(pkt []byte) error
}

// Options configures a Bridge.
type Options struct {
	// DisableMDNS turns off resolving browsers' obfuscated .local ICE
	// candidates. Tests use it; on a Pi it must stay off.
	DisableMDNS bool
}

// Bridge holds one PeerConnection per client.
type Bridge struct {
	api  *webrtc.API
	sink AudioSink

	mu    sync.Mutex
	peers map[core.ClientID]*webrtc.PeerConnection
	onAir core.ClientID
	live  bool

	forwarded atomic.Uint64
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
	var se webrtc.SettingEngine
	if opts.DisableMDNS {
		se.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	} else {
		se.SetICEMulticastDNSMode(ice.MulticastDNSModeQueryOnly)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithSettingEngine(se))
	return &Bridge{api: api, sink: sink, peers: map[core.ClientID]*webrtc.PeerConnection{}}, nil
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
		if t.Kind() != webrtc.RTPCodecTypeAudio {
			return
		}
		b.pump(id, t)
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

// pump forwards a track's packets to the sink while its client is on air.
func (b *Bridge) pump(id core.ClientID, t *webrtc.TrackRemote) {
	buf := make([]byte, 1500)
	for {
		n, _, err := t.Read(buf)
		if err != nil {
			return
		}
		b.mu.Lock()
		on := b.live && b.onAir == id
		b.mu.Unlock()
		if !on {
			continue
		}
		if b.sink.WriteRTP(buf[:n]) == nil {
			b.forwarded.Add(1)
		}
	}
}

// SetOnAir selects whose audio reaches the sink; live=false mutes all.
func (b *Bridge) SetOnAir(id core.ClientID, live bool) {
	b.mu.Lock()
	b.onAir, b.live = id, live
	b.mu.Unlock()
}

// Forwarded counts packets handed to the sink.
func (b *Bridge) Forwarded() uint64 { return b.forwarded.Load() }

// Drop closes a client's connection, if any.
func (b *Bridge) Drop(id core.ClientID) {
	b.mu.Lock()
	pc := b.peers[id]
	delete(b.peers, id)
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
