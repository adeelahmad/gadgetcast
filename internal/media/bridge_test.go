package media

import (
	"context"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

type countSink struct {
	mu sync.Mutex
	n  int
}

func (c *countSink) WriteRTP([]byte) error { c.mu.Lock(); c.n++; c.mu.Unlock(); return nil }
func (c *countSink) count() int            { c.mu.Lock(); defer c.mu.Unlock(); return c.n }

// browser builds a sending peer like a browser would: an Opus track and a
// non-trickle offer carrying all of its candidates.
func browser(t *testing.T) (*webrtc.PeerConnection, *webrtc.TrackLocalStaticSample, string) {
	t.Helper()
	var se webrtc.SettingEngine
	se.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	me := &webrtc.MediaEngine{}
	if err := me.RegisterDefaultCodecs(); err != nil {
		t.Fatal(err)
	}
	pc, err := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithSettingEngine(se)).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2}, "audio", "mic")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pc.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}); err != nil {
		t.Fatal(err)
	}
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	g := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}
	<-g
	return pc, track, pc.LocalDescription().SDP
}

func connect(t *testing.T, b *Bridge, id core.ClientID) *webrtc.TrackLocalStaticSample {
	t.Helper()
	pc, track, offer := browser(t)
	answer, err := b.Offer(id, offer)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "recvonly") || !strings.Contains(answer, "opus") {
		t.Fatalf("answer is not recvonly opus:\n%s", answer)
	}
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: answer}); err != nil {
		t.Fatal(err)
	}
	return track
}

// send writes 20 ms Opus frames until stop closes.
func send(track *webrtc.TrackLocalStaticSample, stop <-chan struct{}) {
	frame := []byte{0xf8, 0xff, 0xfe} // an Opus silence frame
	for {
		select {
		case <-stop:
			return
		case <-time.After(20 * time.Millisecond):
			_ = track.WriteSample(media.Sample{Data: frame, Duration: 20 * time.Millisecond})
		}
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestOnAirClientReachesSinkOthersDoNot(t *testing.T) {
	sink := &countSink{}
	b, err := New(sink, Options{DisableMDNS: true})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	alice, bob := cid("alice"), cid("bob")
	stop := make(chan struct{})
	defer close(stop)
	go send(connect(t, b, alice), stop)
	go send(connect(t, b, bob), stop)

	time.Sleep(300 * time.Millisecond)
	if n := sink.count(); n != 0 {
		t.Fatalf("sink got %d packets with nobody on air", n)
	}

	b.SetOnAir(alice, true)
	waitFor(t, "alice's audio", func() bool { return sink.count() >= 10 })
	if b.Forwarded() == 0 {
		t.Error("Forwarded not counted")
	}

	b.SetOnAir(core.ClientID{}, false)
	time.Sleep(100 * time.Millisecond)
	before := sink.count()
	time.Sleep(300 * time.Millisecond)
	if after := sink.count(); after != before {
		t.Fatalf("sink kept receiving after going off air: %d -> %d", before, after)
	}

	b.Drop(alice)
	b.SetOnAir(bob, true)
	waitFor(t, "bob's audio", func() bool { return sink.count() >= before+10 })
}

func TestOfferErrors(t *testing.T) {
	b, err := New(&countSink{}, Options{DisableMDNS: true})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if _, err := b.Offer(core.ClientID{}, "v=0"); err == nil {
		t.Error("offer without client accepted")
	}
	if _, err := b.Offer(cid("x"), "not sdp"); err == nil {
		t.Error("garbage offer accepted")
	}
	b.Drop(cid("nobody")) // no-op
}

func TestReofferReplacesConnection(t *testing.T) {
	sink := &countSink{}
	b, err := New(sink, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	id := cid("alice")
	connect(t, b, id)
	stop := make(chan struct{})
	defer close(stop)
	go send(connect(t, b, id), stop)
	b.SetOnAir(id, true)
	waitFor(t, "audio after re-offer", func() bool { return sink.count() >= 5 })
	b.mu.Lock()
	n := len(b.peers)
	b.mu.Unlock()
	if n != 1 {
		t.Fatalf("peers = %d, want 1", n)
	}
}

func TestGstPipeline(t *testing.T) {
	args := strings.Join(GstPipeline(5004, "hw:X"), " ")
	for _, want := range []string{"port=5004", "device=hw:X", "opusdec", "rtpjitterbuffer", "rate=48000"} {
		if !strings.Contains(args, want) {
			t.Errorf("pipeline lacks %q: %s", want, args)
		}
	}
}

func cid(s string) core.ClientID {
	id, err := core.NewClientID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func TestGstSinkSupervises(t *testing.T) {
	var logs strings.Builder
	var mu sync.Mutex
	g, err := NewGstSink("", log.New(lockedWriter{&mu, &logs}, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	if g.device != DefaultALSADevice {
		t.Errorf("device = %q", g.device)
	}
	if err := g.WriteRTP([]byte{0x80}); err != nil {
		t.Errorf("WriteRTP: %v", err)
	}
	g.cmd = "false" // exits at once: Run must log and restart until cancelled
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { g.Run(ctx); close(done) }()
	time.Sleep(1500 * time.Millisecond)
	cancel()
	<-done
	mu.Lock()
	out := logs.String()
	mu.Unlock()
	if strings.Count(out, "audio pipeline exited") < 2 {
		t.Errorf("want restarts, log:\n%s", out)
	}

	g2, _ := NewGstSink("hw:X", nil)
	g2.cmd = "gadgetcast-no-such-binary"
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()
	g2.Run(ctx2) // returns once cancelled even without gst installed
}

type lockedWriter struct {
	mu *sync.Mutex
	b  *strings.Builder
}

func (w lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.WriteString(string(p))
}
