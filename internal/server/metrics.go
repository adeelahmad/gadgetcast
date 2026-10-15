package server

import (
	"bufio"
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetricsCollector samples host metrics from procfs/sysfs. Every source is
// best-effort: a missing or unparsable file yields zero (or null tempC), so the
// collector degrades gracefully off Linux.
type MetricsCollector struct {
	procRoot string
	sysRoot  string

	mu       sync.Mutex
	prevIdle uint64
	prevAll  uint64
	last     Metrics
	lastAt   time.Time
}

// NewMetricsCollector builds a collector over the given proc and sys roots
// (normally "/proc" and "/sys").
func NewMetricsCollector(procRoot, sysRoot string) *MetricsCollector {
	return &MetricsCollector{procRoot: procRoot, sysRoot: sysRoot}
}

// minSampleGap bounds how often CPU deltas are recomputed so several concurrent
// callers share one sample instead of producing tiny noisy deltas.
const minSampleGap = 500 * time.Millisecond

// Sample returns a metrics snapshot. CPU% is the busy share since the previous
// sample (or since boot on the first one).
func (m *MetricsCollector) Sample() Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.lastAt.IsZero() && time.Since(m.lastAt) < minSampleGap {
		return cloneMetrics(m.last)
	}
	out := Metrics{LoadAvg: []float64{0, 0, 0}}
	if idle, all, ok := m.readCPU(); ok {
		dIdle, dAll := idle-m.prevIdle, all-m.prevAll
		if all >= m.prevAll && idle >= m.prevIdle && dAll > 0 {
			out.CPUPct = round1(100 * float64(dAll-dIdle) / float64(dAll))
		}
		m.prevIdle, m.prevAll = idle, all
	}
	out.RAMTotalMB, out.RAMUsedMB = m.readMem()
	out.TempC = m.readTemp()
	out.UptimeS = m.readUptime()
	if la := m.readLoad(); la != nil {
		out.LoadAvg = la
	}
	m.last, m.lastAt = out, time.Now()
	return cloneMetrics(out)
}

func cloneMetrics(in Metrics) Metrics {
	out := in
	out.LoadAvg = append([]float64(nil), in.LoadAvg...)
	if in.TempC != nil {
		t := *in.TempC
		out.TempC = &t
	}
	return out
}

func round1(f float64) float64 { return math.Round(f*10) / 10 }

// readCPU parses the aggregate "cpu" line of /proc/stat into idle and total jiffies.
func (m *MetricsCollector) readCPU() (idle, all uint64, ok bool) {
	b, err := os.ReadFile(filepath.Join(m.procRoot, "stat"))
	if err != nil {
		return 0, 0, false
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 5 || f[0] != "cpu" {
			continue
		}
		for i, s := range f[1:] {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			// guest / guest_nice (fields 9, 10) are already counted in user / nice.
			if i >= 8 {
				break
			}
			all += v
			if i == 3 || i == 4 { // idle, iowait
				idle += v
			}
		}
		return idle, all, true
	}
	return 0, 0, false
}

// readMem returns total and used MB from /proc/meminfo (used = total - available).
func (m *MetricsCollector) readMem() (totalMB, usedMB float64) {
	b, err := os.ReadFile(filepath.Join(m.procRoot, "meminfo"))
	if err != nil {
		return 0, 0
	}
	vals := map[string]float64{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 2 {
			continue
		}
		v, err := strconv.ParseFloat(f[1], 64)
		if err != nil {
			continue
		}
		vals[strings.TrimSuffix(f[0], ":")] = v // kB
	}
	total := vals["MemTotal"]
	avail, ok := vals["MemAvailable"]
	if !ok {
		avail = vals["MemFree"] + vals["Buffers"] + vals["Cached"]
	}
	used := total - avail
	if used < 0 {
		used = 0
	}
	return round1(total / 1024), round1(used / 1024)
}

// readTemp returns the SoC temperature in °C, or nil if unavailable.
func (m *MetricsCollector) readTemp() *float64 {
	b, err := os.ReadFile(filepath.Join(m.sysRoot, "class", "thermal", "thermal_zone0", "temp"))
	if err != nil {
		return nil
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	if err != nil {
		return nil
	}
	t := round1(v / 1000)
	return &t
}

func (m *MetricsCollector) readUptime() float64 {
	b, err := os.ReadFile(filepath.Join(m.procRoot, "uptime"))
	if err != nil {
		return 0
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return 0
	}
	v, err := strconv.ParseFloat(f[0], 64)
	if err != nil {
		return 0
	}
	return math.Floor(v)
}

func (m *MetricsCollector) readLoad() []float64 {
	b, err := os.ReadFile(filepath.Join(m.procRoot, "loadavg"))
	if err != nil {
		return nil
	}
	f := strings.Fields(string(b))
	if len(f) < 3 {
		return nil
	}
	out := make([]float64, 3)
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseFloat(f[i], 64)
		if err != nil {
			return nil
		}
		out[i] = v
	}
	return out
}
