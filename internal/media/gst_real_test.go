package media

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These run the exact pipelines through real GStreamer, swapping only the
// hardware element (alsasink, uvcsink, alsasrc) for a file or test source,
// so a typo'd element or property fails here instead of on a Pi. CI's gst job
// installs GStreamer and sets GADGETCAST_GST=1; other runs skip them.
func needGst(t *testing.T) {
	t.Helper()
	if os.Getenv("GADGETCAST_GST") == "" {
		t.Skip("set GADGETCAST_GST=1 with GStreamer installed")
	}
}

// swapLast replaces the pipeline's final element with sink.
func swapLast(args []string, sink ...string) []string {
	last := 0
	for i, a := range args {
		if a == "!" {
			last = i
		}
	}
	return append(append([]string{}, args[:last+1]...), sink...)
}

func freePort(t *testing.T) int {
	t.Helper()
	c, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).Port
}

// receive runs the receiving pipeline while a gst-launch sender streams RTP
// into it, and returns the receiver's output files.
func receive(t *testing.T, recv []string, send string, glob string) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r := exec.CommandContext(ctx, "gst-launch-1.0", recv...)
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Process.Kill(); _ = r.Wait() }()
	time.Sleep(time.Second)
	if out, err := exec.CommandContext(ctx, "gst-launch-1.0", strings.Fields("-q "+send)...).CombinedOutput(); err != nil {
		t.Fatalf("sender: %v\n%s", err, out)
	}
	time.Sleep(time.Second)
	files, _ := filepath.Glob(glob)
	return files
}

func TestRealGstAudioPipeline(t *testing.T) {
	needGst(t)
	dir, port := t.TempDir(), freePort(t)
	out := filepath.Join(dir, "pcm.raw")
	recv := swapLast(GstPipeline(port, "unused"), "filesink", "location="+out)
	receive(t, recv, "audiotestsrc is-live=true num-buffers=100 ! audio/x-raw,rate=48000,channels=2 ! opusenc frame-size=20 ! rtpopuspay pt=111 ! udpsink host=127.0.0.1 port="+strconv.Itoa(port), out)
	st, err := os.Stat(out)
	if err != nil || st.Size() < 48000*2*2/2 { // at least half a second of 48 kHz stereo S16
		t.Fatalf("decoded PCM: %v, %d bytes", err, sizeOf(st))
	}
}

func TestRealGstVideoPipelines(t *testing.T) {
	needGst(t)
	for _, c := range []struct {
		name string
		pt   int
		enc  string
	}{
		{"H264", VideoPT, "openh264enc ! h264parse config-interval=-1 ! rtph264pay config-interval=-1 pt=96"},
		{"VP8", VP8PT, "vp8enc deadline=1 ! rtpvp8pay pt=97"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir, port := t.TempDir(), freePort(t)
			recv := swapLast(GstVideoPipeline(port, "unused", c.pt), "multifilesink", "location="+filepath.Join(dir, "f%03d.jpg"))
			files := receive(t, recv, "videotestsrc is-live=true num-buffers=60 ! video/x-raw,width=320,height=240,framerate=30/1 ! "+c.enc+" ! udpsink host=127.0.0.1 port="+strconv.Itoa(port), filepath.Join(dir, "*.jpg"))
			if len(files) < 10 {
				t.Fatalf("got %d JPEG frames, want at least 10", len(files))
			}
			b, _ := os.ReadFile(files[len(files)/2])
			if len(b) < 2 || b[0] != 0xFF || b[1] != 0xD8 {
				t.Fatalf("%s is not a JPEG", files[len(files)/2])
			}
		})
	}
}

func TestRealGstSpeakerPipeline(t *testing.T) {
	needGst(t)
	ln, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	args := GstSpeakerPipeline(ln.LocalAddr().(*net.UDPAddr).Port, "unused", 0)
	// Swap the capture element (alsasrc and its properties) for a test tone.
	src := []string{"-q", "audiotestsrc", "is-live=true", "num-buffers=50"}
	i := 1
	for args[i] != "!" {
		i++
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gst-launch-1.0", append(src, args[i:]...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("speaker pipeline: %v\n%s", err, out)
	}
	n := 0
	buf := make([]byte, 1500)
	for {
		_ = ln.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		m, _, err := ln.ReadFrom(buf)
		if err != nil {
			break
		}
		if m > 12 && buf[1]&0x7f == AudioPT {
			n++
		}
	}
	if n < 10 {
		t.Fatalf("got %d Opus RTP packets with PT %d, want at least 10", n, AudioPT)
	}
}

func sizeOf(st os.FileInfo) int64 {
	if st == nil {
		return 0
	}
	return st.Size()
}
