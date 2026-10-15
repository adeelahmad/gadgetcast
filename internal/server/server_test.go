package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// fakeHost lays out a minimal /proc, /sys, /dev tree.
func fakeHost(t *testing.T) (procRoot, sysRoot, devRoot string) {
	t.Helper()
	root := t.TempDir()
	procRoot = filepath.Join(root, "proc")
	sysRoot = filepath.Join(root, "sys")
	devRoot = filepath.Join(root, "dev")
	files := map[string]string{
		"proc/stat":    "cpu  100 0 100 700 100 0 0 0 0 0\ncpu0 100 0 100 700 100 0 0 0 0 0\n",
		"proc/meminfo": "MemTotal:        4096000 kB\nMemFree:         1000000 kB\nMemAvailable:    3072000 kB\n",
		"proc/uptime":  "12345.67 40000.00\n",
		"proc/loadavg": "0.50 0.25 0.10 1/200 1234\n",
		"proc/asound/cards": " 0 [Headphones     ]: bcm2835_headphonbcm2835 Headphones - bcm2835 Headphones\n" +
			"                      bcm2835 Headphones\n" +
			" 1 [UAC2Gadget     ]: UAC2_Gadget - UAC2_Gadget\n                      UAC2_Gadget 0\n",
		"sys/class/thermal/thermal_zone0/temp": "48312\n",
		"sys/class/video4linux/video0/name":    "UVC Camera\n",
		"dev/video0":                           "",
		"dev/video10":                          "",
		"dev/null":                             "",
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return procRoot, sysRoot, devRoot
}

type testEnv struct {
	srv     *Server
	ts      *httptest.Server
	dataDir string
}

func newEnv(t *testing.T, static fstest.MapFS) *testEnv {
	t.Helper()
	proc, sys, dev := fakeHost(t)
	data := t.TempDir()
	opts := Options{
		DataDir:         data,
		PIN:             "1234",
		PINAttemptLimit: 3,
		MetricsInterval: 50 * time.Millisecond,
		ProcRoot:        proc,
		SysRoot:         sys,
		DevRoot:         dev,
	}
	if static != nil {
		opts.Static = static
	}
	srv, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		srv.Close()
		ts.Close()
	})
	return &testEnv{srv: srv, ts: ts, dataDir: data}
}

func (e *testEnv) do(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	t.Helper()
	var rdr io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rdr = strings.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, e.ts.URL+path, rdr)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, out
}

func decode[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("decode %T: %v\nbody: %s", v, err, b)
	}
	return v
}

func wantStatus(t *testing.T, resp *http.Response, body []byte, code int) {
	t.Helper()
	if resp.StatusCode != code {
		t.Fatalf("%s %s: status %d, want %d; body %s", resp.Request.Method, resp.Request.URL.Path, resp.StatusCode, code, body)
	}
}

func wantJSONError(t *testing.T, resp *http.Response, body []byte, code int) {
	t.Helper()
	wantStatus(t, resp, body, code)
	e := decode[map[string]string](t, body)
	if strings.TrimSpace(e["error"]) == "" {
		t.Fatalf("error body %s lacks an \"error\" message", body)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("error Content-Type = %q, want application/json", ct)
	}
}

func samplePipeline(id string) Pipeline {
	on := false
	return Pipeline{
		ID:   id,
		Name: "Screen → webcam",
		Nodes: []PipelineNode{
			{ID: "a", Type: "browser_screen", X: 10, Y: 20},
			{ID: "b", Type: "uvc_out", X: 300, Y: 20, Title: "Webcam", Enabled: &on,
				Config: map[string]json.RawMessage{"res": json.RawMessage(`"720p"`), "fps": json.RawMessage(`30`)}},
		},
		Edges: []PipelineEdge{{ID: "e1", From: "a", FromPort: 0, To: "b", ToPort: 0, Kind: PortVideo}},
	}
}

func TestHealth(t *testing.T) {
	e := newEnv(t, nil)
	resp, body := e.do(t, "GET", "/api/health", nil)
	wantStatus(t, resp, body, 200)
	got := decode[map[string]any](t, body)
	if got["ok"] != true || got["version"] != "0.1.0" {
		t.Fatalf("health = %s", body)
	}
}

func TestNodes_CatalogueCoversCategories(t *testing.T) {
	e := newEnv(t, nil)
	resp, body := e.do(t, "GET", "/api/nodes", nil)
	wantStatus(t, resp, body, 200)
	groups := decode[[]NodeGroup](t, body)
	want := []string{"Browser", "Network", "Hardware", "Bluetooth", "Mixers/FX", "AI", "Logic", "Storage", "Outputs"}
	if len(groups) != len(want) {
		t.Fatalf("got %d groups, want %d", len(groups), len(want))
	}
	types := map[string]bool{}
	for i, g := range groups {
		if g.Group != want[i] {
			t.Errorf("group[%d] = %q, want %q", i, g.Group, want[i])
		}
		if len(g.Items) == 0 {
			t.Errorf("group %q is empty", g.Group)
		}
		for _, it := range g.Items {
			if types[it.Type] {
				t.Errorf("duplicate node type %q", it.Type)
			}
			types[it.Type] = true
			if it.Category != g.Group || it.Title == "" || it.Icon == "" || it.Cls == "" {
				t.Errorf("node %q incomplete: %+v", it.Type, it)
			}
		}
	}
	for _, need := range []string{"browser_mic", "uvc_out", "uac2_out", "rtmp_out", "rtp_out", "rtsp_out", "recorder"} {
		if !types[need] {
			t.Errorf("catalogue missing %q", need)
		}
	}
	// Ports serialise as [kind, label] arrays.
	if !strings.Contains(string(body), `["audio","Opus"]`) {
		t.Errorf("ports not serialised as 2-element arrays: %.200s", body)
	}
}

func TestPipelines_SeededDefault(t *testing.T) {
	e := newEnv(t, nil)
	resp, body := e.do(t, "GET", "/api/pipelines", nil)
	wantStatus(t, resp, body, 200)
	list := decode[[]Pipeline](t, body)
	if len(list) != 1 || list[0].ID != "default" || !list[0].IsDefault {
		t.Fatalf("seed = %s", body)
	}
	if list[0].Nodes[0].Type != "browser_mic" || list[0].UpdatedAt == "" {
		t.Fatalf("seed contents = %+v", list[0])
	}
}

func TestPipelines_CRUDAndPersistence(t *testing.T) {
	e := newEnv(t, nil)
	p := samplePipeline("screen-cam_1")

	resp, body := e.do(t, "PUT", "/api/pipelines/screen-cam_1", p)
	wantStatus(t, resp, body, 200)
	saved := decode[Pipeline](t, body)
	if saved.UpdatedAt == "" {
		t.Fatal("server did not set updatedAt")
	}
	if _, err := time.Parse(time.RFC3339, saved.UpdatedAt); err != nil {
		t.Fatalf("updatedAt %q not RFC3339: %v", saved.UpdatedAt, err)
	}

	resp, body = e.do(t, "GET", "/api/pipelines/screen-cam_1", nil)
	wantStatus(t, resp, body, 200)
	got := decode[Pipeline](t, body)
	if got.Name != p.Name || len(got.Nodes) != 2 || len(got.Edges) != 1 {
		t.Fatalf("round-trip = %+v", got)
	}
	if got.Nodes[1].Enabled == nil || *got.Nodes[1].Enabled {
		t.Fatalf("enabled=false lost: %+v", got.Nodes[1])
	}
	if string(got.Nodes[1].Config["res"]) != `"720p"` {
		t.Fatalf("config lost: %s", got.Nodes[1].Config["res"])
	}

	// Persisted as a file; a fresh server over the same data dir sees it.
	if _, err := os.Stat(filepath.Join(e.dataDir, "pipelines", "screen-cam_1.json")); err != nil {
		t.Fatalf("pipeline file missing: %v", err)
	}
	matches, _ := filepath.Glob(filepath.Join(e.dataDir, "pipelines", "*.tmp"))
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
	srv2, err := New(Options{DataDir: e.dataDir, PIN: "1"})
	if err != nil {
		t.Fatal(err)
	}
	ts2 := httptest.NewServer(srv2.Handler())
	defer ts2.Close()
	r2, err := http.Get(ts2.URL + "/api/pipelines")
	if err != nil {
		t.Fatal(err)
	}
	b2, _ := io.ReadAll(r2.Body)
	r2.Body.Close()
	list := decode[[]Pipeline](t, b2)
	if len(list) != 2 || list[0].ID != "default" || list[1].ID != "screen-cam_1" {
		t.Fatalf("reopened list = %s", b2)
	}
	if list[1].UpdatedAt != saved.UpdatedAt {
		t.Fatalf("updatedAt changed across restart: %q vs %q", list[1].UpdatedAt, saved.UpdatedAt)
	}

	resp, body = e.do(t, "DELETE", "/api/pipelines/screen-cam_1", nil)
	wantStatus(t, resp, body, 204)
	resp, body = e.do(t, "GET", "/api/pipelines/screen-cam_1", nil)
	wantJSONError(t, resp, body, 404)
	resp, body = e.do(t, "DELETE", "/api/pipelines/screen-cam_1", nil)
	wantJSONError(t, resp, body, 404)
}

func TestPipelines_DeleteAllDoesNotReseedUntilRestart(t *testing.T) {
	e := newEnv(t, nil)
	resp, body := e.do(t, "DELETE", "/api/pipelines/default", nil)
	wantStatus(t, resp, body, 204)
	resp, body = e.do(t, "GET", "/api/pipelines", nil)
	wantStatus(t, resp, body, 200)
	if strings.TrimSpace(string(body)) != "[]" {
		t.Fatalf("list after delete = %s, want []", body)
	}
}

func TestPipelines_BadRequests(t *testing.T) {
	e := newEnv(t, nil)
	long := strings.Repeat("a", 65)

	cases := []struct {
		name, method, path string
		body               any
		code               int
	}{
		{"get invalid id", "GET", "/api/pipelines/bad.id", nil, 400},
		{"get too-long id", "GET", "/api/pipelines/" + long, nil, 400},
		{"get missing", "GET", "/api/pipelines/nope", nil, 404},
		{"delete invalid id", "DELETE", "/api/pipelines/bad%20id", nil, 400},
		{"run invalid id", "POST", "/api/pipelines/bad.id/run", nil, 400},
		{"run missing", "POST", "/api/pipelines/nope/run", nil, 404},
		{"put invalid path id", "PUT", "/api/pipelines/bad.id", samplePipeline("bad.id"), 400},
		{"put id mismatch", "PUT", "/api/pipelines/abc", samplePipeline("xyz"), 400},
		{"put empty body id", "PUT", "/api/pipelines/abc", samplePipeline(""), 400},
		{"put malformed json", "PUT", "/api/pipelines/abc", `{"id":"abc",`, 400},
		{"put reserved id", "PUT", "/api/pipelines/stop", samplePipeline("stop"), 400},
		{"put missing name", "PUT", "/api/pipelines/abc", func() Pipeline { p := samplePipeline("abc"); p.Name = ""; return p }(), 400},
		{"put dangling edge", "PUT", "/api/pipelines/abc", func() Pipeline { p := samplePipeline("abc"); p.Edges[0].To = "zz"; return p }(), 400},
		{"put bad edge kind", "PUT", "/api/pipelines/abc", func() Pipeline { p := samplePipeline("abc"); p.Edges[0].Kind = "smell"; return p }(), 400},
		{"put duplicate node", "PUT", "/api/pipelines/abc", func() Pipeline { p := samplePipeline("abc"); p.Nodes[1].ID = "a"; return p }(), 400},
		{"unknown api path", "GET", "/api/nope", nil, 404},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, body := e.do(t, c.method, c.path, c.body)
			wantJSONError(t, resp, body, c.code)
		})
	}
	// Nothing invalid was persisted.
	resp, body := e.do(t, "GET", "/api/pipelines", nil)
	wantStatus(t, resp, body, 200)
	if n := len(decode[[]Pipeline](t, body)); n != 1 {
		t.Fatalf("invalid PUTs persisted pipelines: %s", body)
	}
}

func TestRunStopAndStatus(t *testing.T) {
	e := newEnv(t, nil)
	resp, body := e.do(t, "GET", "/api/status", nil)
	wantStatus(t, resp, body, 200)
	if !strings.Contains(string(body), `"running":null`) || !strings.Contains(string(body), `"onAir":null`) {
		t.Fatalf("idle status = %s", body)
	}

	resp, body = e.do(t, "POST", "/api/pipelines/default/run", nil)
	wantStatus(t, resp, body, 200)
	st := decode[Status](t, body)
	if st.Running == nil || *st.Running != "default" {
		t.Fatalf("run status = %s", body)
	}
	resp, body = e.do(t, "GET", "/api/status", nil)
	wantStatus(t, resp, body, 200)
	if st := decode[Status](t, body); st.Running == nil || *st.Running != "default" {
		t.Fatalf("status after run = %s", body)
	}

	resp, body = e.do(t, "POST", "/api/pipelines/stop", nil)
	wantStatus(t, resp, body, 200)
	if st := decode[Status](t, body); st.Running != nil {
		t.Fatalf("stop status = %s", body)
	}

	// Deleting the running pipeline stops it.
	e.do(t, "POST", "/api/pipelines/default/run", nil)
	e.do(t, "DELETE", "/api/pipelines/default", nil)
	_, body = e.do(t, "GET", "/api/status", nil)
	if st := decode[Status](t, body); st.Running != nil {
		t.Fatalf("running survived delete: %s", body)
	}
}

func TestStatus_Metrics(t *testing.T) {
	e := newEnv(t, nil)
	_, body := e.do(t, "GET", "/api/status", nil)
	m := decode[Status](t, body).Metrics
	if m.RAMTotalMB != 4000 || m.RAMUsedMB != 1000 {
		t.Errorf("ram = %v/%v, want 1000/4000", m.RAMUsedMB, m.RAMTotalMB)
	}
	if m.TempC == nil || *m.TempC != 48.3 {
		t.Errorf("tempC = %v, want 48.3", m.TempC)
	}
	if m.UptimeS != 12345 {
		t.Errorf("uptime = %v", m.UptimeS)
	}
	if len(m.LoadAvg) != 3 || m.LoadAvg[0] != 0.5 {
		t.Errorf("loadAvg = %v", m.LoadAvg)
	}
	if m.CPUPct != 20 { // (100+100)/(100+100+700+100) busy since boot
		t.Errorf("cpuPct = %v, want 20", m.CPUPct)
	}
}

func TestMetrics_DegradeWithoutProcfs(t *testing.T) {
	empty := t.TempDir()
	m := NewMetricsCollector(filepath.Join(empty, "proc"), filepath.Join(empty, "sys")).Sample()
	if m.TempC != nil || m.CPUPct != 0 || m.RAMTotalMB != 0 || len(m.LoadAvg) != 3 {
		t.Fatalf("degraded metrics = %+v", m)
	}
	b, _ := json.Marshal(m)
	if !strings.Contains(string(b), `"tempC":null`) {
		t.Fatalf("tempC not null: %s", b)
	}
}

func TestMetrics_CPUDelta(t *testing.T) {
	proc, sys, _ := fakeHost(t)
	c := NewMetricsCollector(proc, sys)
	c.Sample()
	// +50 busy, +50 idle → 50%.
	os.WriteFile(filepath.Join(proc, "stat"), []byte("cpu  150 0 100 750 100 0 0 0 0 0\n"), 0o644)
	c.lastAt = time.Time{} // bypass the shared-sample window
	if got := c.Sample().CPUPct; got != 50 {
		t.Fatalf("cpu delta = %v, want 50", got)
	}
}

func TestDevices(t *testing.T) {
	e := newEnv(t, nil)
	resp, body := e.do(t, "GET", "/api/devices", nil)
	wantStatus(t, resp, body, 200)
	got := decode[devicesResponse](t, body)
	if got.Clients == nil || len(got.Clients) != 0 {
		t.Fatalf("clients = %s", body)
	}
	if len(got.Devices) != 4 {
		t.Fatalf("devices = %+v", got.Devices)
	}
	v0, v10, a0, a1 := got.Devices[0], got.Devices[1], got.Devices[2], got.Devices[3]
	if v0.ID != "video0" || v0.Kind != "video" || v0.Name != "UVC Camera" || !strings.HasSuffix(v0.Path, "/video0") {
		t.Errorf("video0 = %+v", v0)
	}
	if v10.ID != "video10" || v10.Name != "video10" {
		t.Errorf("video10 = %+v", v10)
	}
	if a0.ID != "card0" || a0.Kind != "audio" || a0.Name != "bcm2835 Headphones" || a0.Path != "hw:0" {
		t.Errorf("card0 = %+v", a0)
	}
	if a1.ID != "card1" || a1.Name != "UAC2_Gadget" {
		t.Errorf("card1 = %+v", a1)
	}
}

func TestStatic_SPAFallback(t *testing.T) {
	e := newEnv(t, fstest.MapFS{
		"index.html":    {Data: []byte("<!doctype html><title>Studio</title>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
		".gitkeep":      {Data: nil},
	})
	resp, body := e.do(t, "GET", "/", nil)
	wantStatus(t, resp, body, 200)
	if !strings.Contains(string(body), "Studio") {
		t.Fatalf("index = %s", body)
	}
	resp, body = e.do(t, "GET", "/assets/app.js", nil)
	wantStatus(t, resp, body, 200)
	if string(body) != "console.log(1)" {
		t.Fatalf("asset = %s", body)
	}
	resp, body = e.do(t, "GET", "/pipelines/editor/42", nil)
	wantStatus(t, resp, body, 200)
	if !strings.Contains(string(body), "Studio") || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		t.Fatalf("SPA fallback = %s", body)
	}
	resp, body = e.do(t, "GET", "/.gitkeep", nil)
	if !strings.Contains(string(body), "Studio") {
		t.Fatalf("dotfile served directly: %d %q", resp.StatusCode, body)
	}
}

func TestStatic_NoUINotice(t *testing.T) {
	e := newEnv(t, fstest.MapFS{".gitkeep": {}})
	resp, body := e.do(t, "GET", "/anything", nil)
	wantStatus(t, resp, body, 200)
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") || !strings.Contains(string(body), "index.html is missing") {
		t.Fatalf("notice = %q", body)
	}
}
