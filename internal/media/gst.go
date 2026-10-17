package media

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DefaultALSADevice is the UAC2 gadget's playback device: what the host
// hears as the GadgetCast microphone.
const DefaultALSADevice = "plughw:UAC2Gadget"

// GstPipeline is the gst-launch-1.0 pipeline that turns Opus RTP on a local
// UDP port into 48 kHz PCM on the UAC2 gadget. The jitter buffer absorbs
// network jitter (README §4 latency budget: 30-50 ms); alsasink slaves to
// the USB host's clock and resamples to absorb clock drift.
func GstPipeline(port int, device string) []string {
	p := fmt.Sprintf("udpsrc address=127.0.0.1 port=%d caps=application/x-rtp,media=audio,encoding-name=OPUS,clock-rate=48000,payload=111 "+
		"! rtpjitterbuffer latency=40 drop-on-latency=true ! rtpopusdepay ! opusdec plc=true "+
		"! audioconvert ! audioresample ! audio/x-raw,format=S16LE,rate=48000,channels=2 "+
		"! alsasink device=%s buffer-time=60000 latency-time=10000 slave-method=resample", port, device)
	return append([]string{"-q"}, strings.Fields(p)...)
}

// GstSink forwards RTP to a supervised gst-launch-1.0 process.
type GstSink struct {
	conn   net.Conn
	port   int
	device string
	cmd    string
	logger *log.Logger
}

// NewGstSink reserves a local UDP port for the pipeline. Call Run to start
// and supervise the process.
func NewGstSink(device string, logger *log.Logger) (*GstSink, error) {
	if device == "" {
		device = DefaultALSADevice
	}
	l, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := l.LocalAddr().(*net.UDPAddr).Port
	l.Close() // gst-launch binds it; a race here only costs a restart
	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.New(nopWriter{}, "", 0)
	}
	return &GstSink{conn: conn, port: port, device: device, cmd: "gst-launch-1.0", logger: logger}, nil
}

// WriteRTP sends one packet to the pipeline. Packets sent while the process
// is restarting are dropped, which the jitter buffer conceals.
func (g *GstSink) WriteRTP(pkt []byte) error {
	_, err := g.conn.Write(pkt)
	return err
}

// Run keeps the pipeline running until ctx ends, restarting it with backoff.
func (g *GstSink) Run(ctx context.Context) {
	defer g.conn.Close()
	if _, err := exec.LookPath(g.cmd); err != nil {
		g.logger.Printf("audio output disabled: %s not installed (apt install gstreamer1.0-tools gstreamer1.0-plugins-good gstreamer1.0-plugins-base gstreamer1.0-alsa)", g.cmd)
		<-ctx.Done()
		return
	}
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		c := exec.CommandContext(ctx, g.cmd, GstPipeline(g.port, g.device)...)
		out, err := c.CombinedOutput()
		if ctx.Err() != nil {
			return
		}
		g.logger.Printf("audio pipeline exited (%v): %s", err, strings.TrimSpace(string(out)))
		if time.Since(start) > time.Minute {
			backoff = time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }
