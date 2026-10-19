package gadget

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// fakeConfigFS models the configfs rules that matter for ordering: parents
// must exist, symlink targets must exist, and a directory can only be removed
// once no child directories or links remain (attribute files don't count).
type fakeConfigFS struct {
	dirs  map[string]bool
	files map[string]string
	links map[string]string
	bound string
}

func newFake() *fakeConfigFS {
	return &fakeConfigFS{dirs: map[string]bool{"/": true}, files: map[string]string{}, links: map[string]string{}}
}

func (f *fakeConfigFS) parentOK(p string) bool {
	// configfs creates function class dirs itself (control/class/fs, …).
	if strings.Contains(p, "/class/") {
		return true
	}
	return f.dirs[path.Dir(p)]
}

func (f *fakeConfigFS) Mkdir(p string) error {
	for d := p; d != "/"; d = path.Dir(d) {
		f.dirs[d] = true
	}
	return nil
}

func (f *fakeConfigFS) WriteFile(p, data string) error {
	if !f.dirs[path.Dir(p)] {
		return os.ErrNotExist
	}
	if path.Base(p) == "UDC" {
		if data == "" && f.bound == "" {
			return syscall.ENODEV
		}
		f.bound = data
	}
	f.files[p] = data
	return nil
}

func (f *fakeConfigFS) Symlink(target, p string) error {
	if !f.parentOK(p) {
		return os.ErrNotExist
	}
	if !f.dirs[path.Join(path.Dir(p), target)] {
		return errors.New("dangling link " + p)
	}
	f.links[p] = target
	return nil
}

func (f *fakeConfigFS) Remove(p string) error {
	if _, ok := f.links[p]; ok {
		delete(f.links, p)
		return nil
	}
	if !f.dirs[p] {
		return os.ErrNotExist
	}
	if f.bound != "" && path.Dir(p) == "/" {
		return errors.New("rmdir of bound gadget")
	}
	for d := range f.dirs {
		if strings.HasPrefix(d, p+"/") && !strings.Contains(d, "/class/") {
			if !kernelMade(d) {
				return errors.New("directory not empty: " + p + " has " + d)
			}
		}
	}
	for l := range f.links {
		if strings.HasPrefix(l, p+"/") {
			return errors.New("directory not empty: " + p + " has link " + l)
		}
	}
	for d := range f.dirs {
		if d == p || strings.HasPrefix(d, p+"/") {
			delete(f.dirs, d)
		}
	}
	return nil
}

// kernelMade reports directories configfs creates on its own inside a
// function, which userspace never removes.
func kernelMade(d string) bool {
	for _, s := range []string{"/control", "/control/header", "/streaming", "/streaming/header", "/streaming/mjpeg", "/configs/c.1/strings", "/strings", "/configs", "/functions"} {
		if strings.HasSuffix(d, s) {
			return true
		}
	}
	return false
}

func TestPlanDefaultHasAudioVideoAndBind(t *testing.T) {
	ops, err := Plan(DefaultSpec(), "fe980000.usb")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"gadgetcast/idVendor":                                                  "0x1d6b",
		"gadgetcast/bDeviceClass":                                              "0xef",
		"gadgetcast/functions/uac2.usb0/p_srate":                               "48000",
		"gadgetcast/functions/uac2.usb0/c_chmask":                              "3",
		"gadgetcast/functions/uvc.usb0/streaming/mjpeg/m/720p/wWidth":          "1280",
		"gadgetcast/functions/uvc.usb0/streaming/mjpeg/m/1080p/wHeight":        "1080",
		"gadgetcast/functions/uvc.usb0/streaming/mjpeg/m/720p/dwFrameInterval": "333333",
	}
	got := map[string]string{}
	for _, op := range ops {
		if op.Kind == Write {
			got[op.Path] = op.Data
		}
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	last := ops[len(ops)-1]
	if last.Kind != Bind || last.Path != "gadgetcast/UDC" || last.Data != "fe980000.usb" {
		t.Errorf("last op = %+v, want bind to UDC", last)
	}
	links := 0
	for _, op := range ops {
		if op.Kind == Symlink && strings.HasPrefix(op.Path, "gadgetcast/configs/c.1/") {
			links++
		}
	}
	if links != 2 {
		t.Errorf("config links = %d, want 2", links)
	}
}

func TestPlanAudioOnly(t *testing.T) {
	s := DefaultSpec()
	s.Video = false
	ops, err := Plan(s, "udc")
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range ops {
		if strings.Contains(op.Path, "uvc") {
			t.Fatalf("audio-only plan touches %s", op.Path)
		}
	}
}

func TestValidate(t *testing.T) {
	bad := []func(*Spec){
		func(s *Spec) { s.Name = "" },
		func(s *Spec) { s.Name = "a/b" },
		func(s *Spec) { s.Name = ".." },
		func(s *Spec) { s.Audio, s.Video = false, false },
		func(s *Spec) { s.Modes = nil },
		func(s *Spec) { s.Modes = []Mode{{0, 720, 30}} },
		func(s *Spec) { s.Modes = []Mode{{1280, 720, 0}} },
	}
	for i, mut := range bad {
		s := DefaultSpec()
		mut(&s)
		if _, err := Plan(s, "udc"); err == nil {
			t.Errorf("case %d: Plan accepted invalid spec", i)
		}
	}
	if _, err := Plan(DefaultSpec(), ""); err == nil {
		t.Error("Plan accepted empty UDC")
	}
}

func TestUpDownOnFakeConfigFS(t *testing.T) {
	fs := newFake()
	s := DefaultSpec()
	if err := Up(fs, "/", s, "udc0"); err != nil {
		t.Fatal(err)
	}
	if fs.bound != "udc0" {
		t.Fatalf("bound = %q", fs.bound)
	}
	// Repeating Up rebuilds cleanly.
	if err := Up(fs, "/", s, "udc0"); err != nil {
		t.Fatalf("second Up: %v", err)
	}
	if err := Down(fs, "/", s); err != nil {
		t.Fatal(err)
	}
	if fs.bound != "" {
		t.Errorf("still bound to %q", fs.bound)
	}
	if fs.dirs["/gadgetcast"] || len(fs.links) != 0 {
		t.Errorf("leftovers: dirs=%v links=%v", fs.dirs, fs.links)
	}
	// Down on an absent gadget is a no-op.
	if err := Down(fs, "/", s); err != nil {
		t.Errorf("Down on absent gadget: %v", err)
	}
}

func TestDownReportsRealErrors(t *testing.T) {
	if err := Down(failingExec{}, "/", DefaultSpec()); err == nil {
		t.Fatal("Down hid an error")
	}
	if err := Up(failingExec{}, "/", DefaultSpec(), "udc"); err == nil {
		t.Fatal("Up hid an error")
	}
}

type failingExec struct{}

func (failingExec) Mkdir(string) error             { return errors.New("boom") }
func (failingExec) WriteFile(string, string) error { return errors.New("boom") }
func (failingExec) Symlink(string, string) error   { return errors.New("boom") }
func (failingExec) Remove(string) error            { return errors.New("boom") }

type mkdirFails struct{ fakeConfigFS }

func (m *mkdirFails) Mkdir(string) error { return errors.New("no configfs") }

func TestUpReportsStepErrors(t *testing.T) {
	m := &mkdirFails{*newFake()}
	err := Up(m, "/", DefaultSpec(), "udc")
	if err == nil || !strings.Contains(err.Error(), "gadgetcast") {
		t.Fatalf("err = %v", err)
	}
}

func TestUpOnRealFilesystemResolvesLinks(t *testing.T) {
	root := t.TempDir()
	if err := Up(OS{}, root, DefaultSpec(), "udc0"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "gadgetcast/UDC"))
	if err != nil || string(b) != "udc0\n" {
		t.Fatalf("UDC = %q, %v", b, err)
	}
	for _, l := range []string{
		"gadgetcast/configs/c.1/uvc.usb0",
		"gadgetcast/configs/c.1/uac2.usb0",
		"gadgetcast/functions/uvc.usb0/streaming/class/hs/h",
		"gadgetcast/functions/uvc.usb0/streaming/header/h/m",
	} {
		if _, err := os.Stat(filepath.Join(root, l)); err != nil {
			t.Errorf("%s does not resolve: %v", l, err)
		}
	}
}

// configfs resolves a symlink target against the caller's working directory,
// not the link's directory, so OS must write relative targets as absolute.
func TestOSSymlinkWritesAbsoluteTarget(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "header/h"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "class/fs/h")
	if err := (OS{}).Symlink("../../header/h", link); err != nil {
		t.Fatal(err)
	}
	got, err := os.Readlink(link)
	if want := filepath.Join(root, "header/h"); err != nil || got != want {
		t.Fatalf("link target = %q, %v; want %q", got, err, want)
	}
}

func TestFindUDC(t *testing.T) {
	dir := t.TempDir()
	// The overlay alone is not enough: Pi 1 B, 2 B, 3 B and 3 B+ have no
	// device-capable port, so the message must name the boards that do.
	if _, err := FindUDC(dir); err == nil || !strings.Contains(err.Error(), "dr_mode=peripheral") || !strings.Contains(err.Error(), "Zero") {
		t.Fatalf("empty dir: err = %v", err)
	}
	if _, err := FindUDC(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing dir: want error")
	}
	for _, n := range []string{"b.usb", "a.usb"} {
		if err := os.WriteFile(filepath.Join(dir, n), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := FindUDC(dir)
	if err != nil || got != "a.usb" {
		t.Fatalf("FindUDC = %q, %v", got, err)
	}
	f := filepath.Join(dir, "a.usb")
	if _, err := FindUDC(f); err == nil {
		t.Fatal("file instead of dir: want error")
	}
}
