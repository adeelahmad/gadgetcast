package server

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// DeviceScanner enumerates local V4L2 and ALSA devices best-effort.
type DeviceScanner struct {
	devRoot  string
	procRoot string
	sysRoot  string
}

// NewDeviceScanner builds a scanner over the given roots (normally "/dev",
// "/proc", "/sys").
func NewDeviceScanner(devRoot, procRoot, sysRoot string) *DeviceScanner {
	return &DeviceScanner{devRoot: devRoot, procRoot: procRoot, sysRoot: sysRoot}
}

var (
	videoNode = regexp.MustCompile(`^video(\d+)$`)
	// " 0 [Headphones     ]: bcm2835_headphonbcm2835 Headphones - bcm2835 Headphones"
	asoundCard = regexp.MustCompile(`^\s*(\d+)\s+\[([^\]]*)\]:\s*(.*)$`)
)

// Scan returns video devices (sorted by index) followed by audio cards.
func (d *DeviceScanner) Scan() []Device {
	out := []Device{}
	out = append(out, d.scanVideo()...)
	out = append(out, d.scanAudio()...)
	return out
}

func (d *DeviceScanner) scanVideo() []Device {
	entries, err := os.ReadDir(d.devRoot)
	if err != nil {
		return nil
	}
	type idx struct {
		n    int
		name string
	}
	var found []idx
	for _, e := range entries {
		m := videoNode.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		found = append(found, idx{n, e.Name()})
	}
	sort.Slice(found, func(i, j int) bool { return found[i].n < found[j].n })
	out := make([]Device, 0, len(found))
	for _, f := range found {
		name := f.name
		if b, err := os.ReadFile(filepath.Join(d.sysRoot, "class", "video4linux", f.name, "name")); err == nil {
			if s := strings.TrimSpace(string(b)); s != "" {
				name = s
			}
		}
		out = append(out, Device{ID: f.name, Kind: "video", Name: name, Path: filepath.Join(d.devRoot, f.name)})
	}
	return out
}

func (d *DeviceScanner) scanAudio() []Device {
	b, err := os.ReadFile(filepath.Join(d.procRoot, "asound", "cards"))
	if err != nil {
		return nil
	}
	var out []Device
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		m := asoundCard.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		name := strings.TrimSpace(m[3])
		if i := strings.LastIndex(name, " - "); i >= 0 {
			name = strings.TrimSpace(name[i+3:])
		}
		if name == "" {
			name = strings.TrimSpace(m[2])
		}
		out = append(out, Device{ID: "card" + m[1], Kind: "audio", Name: name, Path: "hw:" + m[1]})
	}
	return out
}
