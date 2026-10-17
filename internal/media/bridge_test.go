package media

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/rtcp"
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

// videoBrowser is browser() with an H.264 video track as well as Opus.
func videoBrowser(t *testing.T) (*webrtc.PeerConnection, *webrtc.TrackLocalStaticSample, string, chan struct{}) {
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
	vt, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000,
		SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f"}, "video", "cam")
	if err != nil {
		t.Fatal(err)
	}
	sender, err := pc.AddTrack(vt)
	if err != nil {
		t.Fatal(err)
	}
	plis := make(chan struct{}, 16)
	go func() {
		for {
			pkts, _, err := sender.ReadRTCP()
			if err != nil {
				return
			}
			for _, p := range pkts {
				if _, ok := p.(*rtcp.PictureLossIndication); ok {
					select {
					case plis <- struct{}{}:
					default:
					}
				}
			}
		}
	}()
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	g := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}
	<-g
	return pc, vt, pc.LocalDescription().SDP, plis
}

type ptSink struct {
	mu  sync.Mutex
	n   int
	pts map[byte]bool
}

func (s *ptSink) WriteRTP(p []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pts == nil {
		s.pts = map[byte]bool{}
	}
	s.n++
	s.pts[p[1]&0x7f] = true
	return nil
}

func (s *ptSink) snapshot() (int, map[byte]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := map[byte]bool{}
	for k := range s.pts {
		cp[k] = true
	}
	return s.n, cp
}

func TestVideoReachesVideoSinkWithFixedPTAndKeyframeRequest(t *testing.T) {
	video := &ptSink{}
	b, err := New(&countSink{}, Options{DisableMDNS: true, Video: video})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	id := cid("cam")
	pc, vt, offer, plis := videoBrowser(t)
	answer, err := b.Offer(id, offer)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "H264") {
		t.Fatalf("answer lacks H264:\n%s", answer)
	}
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: answer}); err != nil {
		t.Fatal(err)
	}
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		nal := []byte{0, 0, 0, 1, 0x65, 0x88, 0x84, 0x00, 0x33}
		for {
			select {
			case <-stop:
				return
			case <-time.After(33 * time.Millisecond):
				_ = vt.WriteSample(media.Sample{Data: nal, Duration: 33 * time.Millisecond})
			}
		}
	}()
	time.Sleep(300 * time.Millisecond)
	if n, _ := video.snapshot(); n != 0 {
		t.Fatalf("video forwarded while off air: %d", n)
	}
	b.SetOnAir(id, true)
	waitFor(t, "video packets", func() bool { n, _ := video.snapshot(); return n >= 10 })
	if _, pts := video.snapshot(); len(pts) != 1 || !pts[VideoPT] {
		t.Errorf("payload types = %v, want only %d", pts, VideoPT)
	}
	select {
	case <-plis:
	case <-time.After(5 * time.Second):
		t.Fatal("no keyframe request after going on air")
	}
	if b.ForwardedVideo() == 0 {
		t.Error("ForwardedVideo not counted")
	}
}

func TestFindUVCDevice(t *testing.T) {
	root := t.TempDir()
	write := func(dev, name string) {
		d := filepath.Join(root, "class/video4linux", dev)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "name"), []byte(name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := FindUVCDevice(root); got != "" {
		t.Fatalf("empty sysfs: %q", got)
	}
	write("video0", "bcm2835-codec-decode")
	write("video3", "fe980000.usb")
	if got := FindUVCDevice(root); got != "/dev/video3" {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(strings.Join(GstVideoPipeline(5000, "/dev/video3", VideoPT), " "), "uvcsink v4l2sink::device=/dev/video3") {
		t.Error("video pipeline does not target the device")
	}
}

func TestGstVideoSinkWaitsForGadget(t *testing.T) {
	var logs strings.Builder
	var mu sync.Mutex
	g, err := NewGstVideoSink(AutoUVCDevice, log.New(lockedWriter{&mu, &logs}, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	g.sysRoot = t.TempDir()
	g.cmd = "true"
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	g.Run(ctx)
	mu.Lock()
	defer mu.Unlock()
	if !strings.Contains(logs.String(), "no USB webcam gadget") {
		t.Errorf("log: %s", logs.String())
	}
}

func TestGstVideoSinkRestartsOnCodecChange(t *testing.T) {
	if args := strings.Join(GstVideoPipeline(5006, "/dev/video9", VP8PT), " "); !strings.Contains(args, "vp8dec") || !strings.Contains(args, "payload=97") {
		t.Errorf("VP8 pipeline: %s", args)
	}
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := filepath.Join(dir, "gst")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"$@\" >> "+argsFile+"\nexec sleep 60\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	var logs strings.Builder
	var mu sync.Mutex
	g, err := NewGstVideoSink("/dev/video9", log.New(lockedWriter{&mu, &logs}, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	g.cmd = script
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { g.Run(ctx); close(done) }()
	defer func() { cancel(); <-done }()

	started := func(n int) bool { b, _ := os.ReadFile(argsFile); return strings.Count(string(b), "\n") >= n }
	waitFor(t, "first pipeline", func() bool { return started(1) })
	_ = g.WriteRTP([]byte{0x80, VideoPT}) // same codec: no restart
	_ = g.WriteRTP([]byte{0x80, 0x80 | VP8PT})
	waitFor(t, "VP8 pipeline", func() bool { return started(2) })
	b, _ := os.ReadFile(argsFile)
	runs := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(runs) != 2 || !strings.Contains(runs[0], "h264parse") || !strings.Contains(runs[1], "vp8dec") {
		t.Errorf("runs:\n%s", b)
	}
}
