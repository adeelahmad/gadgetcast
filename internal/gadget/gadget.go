// Package gadget builds the configfs USB composite gadget that makes the Pi
// appear to the host as a UVC webcam plus a UAC2 microphone and speaker
// (README §3.1). Setup is expressed as an ordered list of filesystem steps so
// it can be tested without a kernel; teardown is the same list in reverse.
package gadget

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// DefaultConfigFS is where configfs exposes USB gadgets.
const DefaultConfigFS = "/sys/kernel/config/usb_gadget"

// DefaultUDCDir lists the USB device controllers a gadget can bind to.
const DefaultUDCDir = "/sys/class/udc"

// Mode is one MJPEG frame size offered to the host.
type Mode struct {
	Width, Height int
	FPS           int
}

// Spec describes the composite device.
type Spec struct {
	Name         string // directory under the configfs gadget root
	VendorID     uint16
	ProductID    uint16
	Manufacturer string
	Product      string
	Serial       string
	Audio        bool // UAC2 mic (to host) + speaker (from host)
	Video        bool // UVC webcam, MJPEG
	Modes        []Mode
}

// DefaultSpec is the gadget gadgetcast presents: Linux Foundation multifunction
// composite IDs, 48 kHz stereo audio both ways, and MJPEG at 720p and 1080p.
func DefaultSpec() Spec {
	return Spec{
		Name:         "gadgetcast",
		VendorID:     0x1d6b,
		ProductID:    0x0104,
		Manufacturer: "gadgetcast",
		Product:      "GadgetCast",
		Serial:       "0001",
		Audio:        true,
		Video:        true,
		Modes:        []Mode{{1280, 720, 30}, {1920, 1080, 30}},
	}
}

// OpKind is one kind of filesystem step.
type OpKind int

const (
	Mkdir   OpKind = iota // create a directory (rmdir on teardown)
	Write                 // write an attribute file (skipped on teardown)
	Symlink               // link Path -> Target (removed on teardown)
	Bind                  // write the UDC name to Path ("" on teardown)
)

// Op is one filesystem step. Paths are relative to the configfs gadget root.
type Op struct {
	Kind   OpKind
	Path   string
	Data   string // Write, Bind
	Target string // Symlink: relative to the link's directory
}

// Validate reports a spec that cannot produce a working device.
func (s Spec) Validate() error {
	if s.Name == "" || strings.ContainsAny(s.Name, `/\`) || s.Name == "." || s.Name == ".." {
		return fmt.Errorf("gadget: invalid name %q", s.Name)
	}
	if !s.Audio && !s.Video {
		return errors.New("gadget: enable at least one of audio and video")
	}
	if s.Video {
		if len(s.Modes) == 0 {
			return errors.New("gadget: video needs at least one mode")
		}
		for _, m := range s.Modes {
			if m.Width <= 0 || m.Height <= 0 || m.FPS <= 0 {
				return fmt.Errorf("gadget: invalid mode %dx%d@%d", m.Width, m.Height, m.FPS)
			}
		}
	}
	return nil
}

// Plan returns the ordered steps that create the gadget and bind it to udc.
func Plan(s Spec, udc string) ([]Op, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if udc == "" {
		return nil, errors.New("gadget: no USB device controller")
	}
	var ops []Op
	g := s.Name
	mk := func(p ...string) { ops = append(ops, Op{Kind: Mkdir, Path: path.Join(p...)}) }
	wr := func(p, data string) { ops = append(ops, Op{Kind: Write, Path: p, Data: data}) }
	ln := func(p, target string) { ops = append(ops, Op{Kind: Symlink, Path: p, Target: target}) }

	mk(g)
	wr(g+"/idVendor", fmt.Sprintf("0x%04x", s.VendorID))
	wr(g+"/idProduct", fmt.Sprintf("0x%04x", s.ProductID))
	wr(g+"/bcdDevice", "0x0100")
	wr(g+"/bcdUSB", "0x0200")
	// Miscellaneous class + Interface Association Descriptors: required for
	// Windows to bind UVC and UAC2 functions in one composite device.
	wr(g+"/bDeviceClass", "0xef")
	wr(g+"/bDeviceSubClass", "0x02")
	wr(g+"/bDeviceProtocol", "0x01")
	mk(g, "strings/0x409")
	wr(g+"/strings/0x409/manufacturer", s.Manufacturer)
	wr(g+"/strings/0x409/product", s.Product)
	wr(g+"/strings/0x409/serialnumber", s.Serial)
	mk(g, "configs/c.1")
	mk(g, "configs/c.1/strings/0x409")
	wr(g+"/configs/c.1/strings/0x409/configuration", s.Product)
	wr(g+"/configs/c.1/MaxPower", "500")
	wr(g+"/configs/c.1/bmAttributes", "0x80")

	var funcs []string
	if s.Video {
		f := g + "/functions/uvc.usb0"
		funcs = append(funcs, "uvc.usb0")
		mk(f)
		wr(f+"/streaming_maxpacket", "2048")
		mk(f, "control/header/h")
		ln(f+"/control/class/fs/h", "../../header/h")
		ln(f+"/control/class/ss/h", "../../header/h")
		mk(f, "streaming/mjpeg/m")
		for _, m := range s.Modes {
			fr := fmt.Sprintf("%s/streaming/mjpeg/m/%dp", f, m.Height)
			mk(fr)
			interval := 10_000_000 / m.FPS // 100 ns units
			bits := m.Width * m.Height * 16 * m.FPS
			wr(fr+"/wWidth", fmt.Sprint(m.Width))
			wr(fr+"/wHeight", fmt.Sprint(m.Height))
			wr(fr+"/dwMinBitRate", fmt.Sprint(bits))
			wr(fr+"/dwMaxBitRate", fmt.Sprint(bits))
			wr(fr+"/dwMaxVideoFrameBufferSize", fmt.Sprint(m.Width*m.Height*2))
			wr(fr+"/dwDefaultFrameInterval", fmt.Sprint(interval))
			wr(fr+"/dwFrameInterval", fmt.Sprint(interval))
		}
		mk(f, "streaming/header/h")
		ln(f+"/streaming/header/h/m", "../../mjpeg/m")
		ln(f+"/streaming/class/fs/h", "../../header/h")
		ln(f+"/streaming/class/hs/h", "../../header/h")
		ln(f+"/streaming/class/ss/h", "../../header/h")
	}
	if s.Audio {
		f := g + "/functions/uac2.usb0"
		funcs = append(funcs, "uac2.usb0")
		mk(f)
		// p_* is the gadget's playback (the host's microphone), c_* its
		// capture (the host's speaker).
		for _, dir := range []string{"p", "c"} {
			wr(f+"/"+dir+"_chmask", "3")
			wr(f+"/"+dir+"_srate", "48000")
			wr(f+"/"+dir+"_ssize", "2")
		}
	}
	for _, fn := range funcs {
		ln(g+"/configs/c.1/"+fn, "../../functions/"+fn)
	}
	ops = append(ops, Op{Kind: Bind, Path: g + "/UDC", Data: udc})
	return ops, nil
}

// Teardown returns the steps that undo plan, in configfs-safe order.
func Teardown(plan []Op) []Op {
	out := make([]Op, 0, len(plan))
	for i := len(plan) - 1; i >= 0; i-- {
		op := plan[i]
		switch op.Kind {
		case Write:
			continue // configfs attributes go away with their directory
		case Bind:
			op.Data = ""
		}
		out = append(out, op)
	}
	return out
}

// Executor performs filesystem steps. OS is the real one.
type Executor interface {
	Mkdir(path string) error
	WriteFile(path, data string) error
	Symlink(target, path string) error
	Remove(path string) error
}

// OS executes steps on the real filesystem.
type OS struct{}

func (OS) Mkdir(p string) error { return os.MkdirAll(p, 0o755) }
func (OS) WriteFile(p, data string) error {
	return os.WriteFile(p, []byte(data+"\n"), 0o644)
}

// Symlink makes a relative target absolute because configfs resolves link
// targets against the working directory, not the link's own directory.
func (OS) Symlink(target, p string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(p), target)
	}
	return os.Symlink(target, p)
}
func (OS) Remove(p string) error { return os.Remove(p) }

// Up creates the gadget under root and binds it. A gadget left from a
// previous run is torn down first so Up is safe to repeat.
func Up(x Executor, root string, s Spec, udc string) error {
	plan, err := Plan(s, udc)
	if err != nil {
		return err
	}
	if err := Down(x, root, s); err != nil {
		return err
	}
	for _, op := range plan {
		p := filepath.Join(root, op.Path)
		switch op.Kind {
		case Mkdir:
			err = x.Mkdir(p)
		case Write, Bind:
			err = x.WriteFile(p, op.Data)
		case Symlink:
			err = x.Symlink(op.Target, p)
		}
		if err != nil {
			return fmt.Errorf("gadget: %s: %w", op.Path, err)
		}
	}
	return nil
}

// Down unbinds and removes the gadget. Missing pieces are ignored, so Down
// is safe on a half-built or absent gadget.
func Down(x Executor, root string, s Spec) error {
	plan, err := Plan(s, "teardown")
	if err != nil {
		return err
	}
	for _, op := range Teardown(plan) {
		p := filepath.Join(root, op.Path)
		switch op.Kind {
		case Bind:
			err = x.WriteFile(p, op.Data)
		default:
			err = x.Remove(p)
		}
		// ENODEV: unbinding a gadget that is not bound.
		if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, syscall.ENODEV) {
			return fmt.Errorf("gadget: %s: %w", op.Path, err)
		}
	}
	return nil
}

// FindUDC returns the first USB device controller in dir (normally
// /sys/class/udc). It is empty when the Pi's USB port is not in peripheral
// mode, which means dtoverlay=dwc2,dr_mode=peripheral is missing.
func FindUDC(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "", errors.New("gadget: no USB device controller found; enable dtoverlay=dwc2,dr_mode=peripheral and reboot")
	}
	return names[0], nil
}
