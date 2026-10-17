package media

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
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

// GstVideoPipeline turns H.264 RTP on a local UDP port into MJPEG on the UVC
// gadget's video node (README §4: MJPEG because USB 2.0 can't carry raw
// 1080p). decodebin picks the Pi 4's hardware decoder when present; uvcsink
// answers the host's UVC requests and scales to the size the host picks.
// pt selects the decoder: VideoPT (H.264) or VP8PT.
func GstVideoPipeline(port int, device string, pt int) []string {
	enc, dec := "H264", "rtph264depay ! h264parse ! decodebin"
	if pt == VP8PT {
		enc, dec = "VP8", "rtpvp8depay ! vp8dec"
	}
	p := fmt.Sprintf("udpsrc address=127.0.0.1 port=%d caps=application/x-rtp,media=video,encoding-name=%s,clock-rate=90000,payload=%d "+
		"! rtpjitterbuffer latency=50 drop-on-latency=true ! %s "+
		"! videoconvert ! videoscale ! videorate ! jpegenc quality=85 "+
		"! uvcsink v4l2sink::device=%s", port, enc, pt, dec, device)
	return append([]string{"-q"}, strings.Fields(p)...)
}

// FindUVCDevice returns the UVC gadget's V4L2 output node, found by its
// sysfs name (the UDC, e.g. "fe980000.usb"), or "" if the gadget has no
// webcam function yet. sysRoot is normally "/sys".
func FindUVCDevice(sysRoot string) string {
	matches, _ := filepath.Glob(filepath.Join(sysRoot, "class/video4linux/video*/name"))
	sort.Strings(matches)
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(b))
		if strings.HasSuffix(name, ".usb") || strings.Contains(strings.ToLower(name), "gadget") {
			return "/dev/" + filepath.Base(filepath.Dir(m))
		}
	}
	return ""
}

// GstSink forwards RTP to a supervised gst-launch-1.0 process.
type GstSink struct {
	conn     net.Conn
	port     int
	device   string
	kind     string // "audio" or "video", for log lines
	pipeline func(port int, device string, pt int) []string
	cmd      string
	switchPT bool         // restart the pipeline when the payload type changes (video)
	pt       atomic.Int32 // payload type the running pipeline decodes
	restart  chan struct{}
	sysRoot  string // for AutoUVCDevice; "/sys" outside tests
	logger   *log.Logger
}

// NewGstSink reserves a local UDP port for the audio pipeline. Call Run to
// start and supervise the process.
func NewGstSink(device string, logger *log.Logger) (*GstSink, error) {
	if device == "" {
		device = DefaultALSADevice
	}
	return newGstSink("audio", device, func(port int, device string, _ int) []string { return GstPipeline(port, device) }, logger)
}

// AutoUVCDevice makes the video sink look up the UVC gadget's node with
// FindUVCDevice each time it starts the pipeline.
const AutoUVCDevice = "auto"

// NewGstVideoSink is NewGstSink for the webcam: device is the UVC gadget's
// video node, or AutoUVCDevice.
func NewGstVideoSink(device string, logger *log.Logger) (*GstSink, error) {
	g, err := newGstSink("video", device, GstVideoPipeline, logger)
	if err == nil {
		g.switchPT = true
		g.pt.Store(VideoPT)
	}
	return g, err
}

func newGstSink(kind, device string, pipeline func(int, string, int) []string, logger *log.Logger) (*GstSink, error) {
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
	return &GstSink{conn: conn, port: port, device: device, kind: kind, pipeline: pipeline, cmd: "gst-launch-1.0", sysRoot: "/sys", logger: logger, restart: make(chan struct{}, 1)}, nil
}

// WriteRTP sends one packet to the pipeline. Packets sent while the process
// is restarting are dropped, which the jitter buffer conceals.
func (g *GstSink) WriteRTP(pkt []byte) error {
	if g.switchPT && len(pkt) > 1 {
		if pt := int32(pkt[1] & 0x7f); g.pt.Swap(pt) != pt {
			select { // a different codec went on air: restart with its decoder
			case g.restart <- struct{}{}:
			default:
			}
		}
	}
	_, err := g.conn.Write(pkt)
	return err
}

// Run keeps the pipeline running until ctx ends, restarting it with backoff.
func (g *GstSink) Run(ctx context.Context) {
	defer g.conn.Close()
	if _, err := exec.LookPath(g.cmd); err != nil {
		g.logger.Printf("%s output disabled: %s not installed (apt install gstreamer1.0-tools gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-libav gstreamer1.0-alsa)", g.kind, g.cmd)
		<-ctx.Done()
		return
	}
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		device := g.device
		if device == AutoUVCDevice {
			device = FindUVCDevice(g.sysRoot)
		}
		if device == "" {
			g.logger.Printf("%s output waiting: no USB webcam gadget (run: gadgetcast gadget up)", g.kind)
		} else {
			select { // the pipeline below starts with the current payload type
			case <-g.restart:
			default:
			}
			runCtx, cancel := context.WithCancel(ctx)
			done := make(chan struct{})
			go func() {
				select {
				case <-g.restart:
					cancel()
				case <-done:
				}
			}()
			c := exec.CommandContext(runCtx, g.cmd, g.pipeline(g.port, device, int(g.pt.Load()))...)
			out, err := c.CombinedOutput()
			close(done)
			restarted := runCtx.Err() != nil
			cancel()
			if ctx.Err() != nil {
				return
			}
			if restarted {
				g.logger.Printf("%s pipeline restarting for payload type %d", g.kind, g.pt.Load())
				continue
			}
			g.logger.Printf("%s pipeline exited (%v): %s", g.kind, err, strings.TrimSpace(string(out)))
		}
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

// DefaultSpeakerDevice is the UAC2 gadget's capture device: what the host
// plays to the GadgetCast speaker.
const DefaultSpeakerDevice = "plughw:UAC2Gadget"

// GstSpeakerPipeline captures the host's audio from the UAC2 gadget and
// sends it as Opus RTP (AudioPT, 20 ms frames) to a local UDP port.
func GstSpeakerPipeline(port int, device string, _ int) []string {
	p := fmt.Sprintf("alsasrc device=%s buffer-time=40000 latency-time=10000 "+
		"! audioconvert ! audioresample ! audio/x-raw,rate=48000,channels=2 "+
		"! opusenc bitrate=96000 frame-size=20 ! rtpopuspay pt=%d "+
		"! udpsink host=127.0.0.1 port=%d sync=false async=false", device, AudioPT, port)
	return append([]string{"-q"}, strings.Fields(p)...)
}

// GstSource reads RTP from a supervised gst-launch-1.0 process.
type GstSource struct {
	g *GstSink // supervision; its conn is the UDP socket the pipeline sends to
}

// NewGstSource binds a local UDP port for the speaker pipeline. Call Run to
// start it.
func NewGstSource(device string, logger *log.Logger) (*GstSource, error) {
	if device == "" {
		device = DefaultSpeakerDevice
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.New(nopWriter{}, "", 0)
	}
	g := &GstSink{conn: conn, port: conn.LocalAddr().(*net.UDPAddr).Port, device: device, kind: "speaker", pipeline: GstSpeakerPipeline,
		cmd: "gst-launch-1.0", sysRoot: "/sys", logger: logger, restart: make(chan struct{}, 1)}
	return &GstSource{g: g}, nil
}

// Run keeps the pipeline running until ctx ends and hands each packet to
// write.
func (s *GstSource) Run(ctx context.Context, write func(pkt []byte) error) {
	go func() {
		buf := make([]byte, 1500)
		for {
			n, err := s.g.conn.Read(buf)
			if err != nil {
				return // closed by Run's supervisor on exit
			}
			_ = write(buf[:n])
		}
	}()
	s.g.Run(ctx)
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }
