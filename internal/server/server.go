// Package server is the Pi-side HTTP / WebSocket server: the REST API and
// event push described in docs/api.md, the /ws WebRTC signaling transport over
// session.Hub, and the embedded Studio UI.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/adeelahmad/gadgetcast/internal/core"
	"github.com/adeelahmad/gadgetcast/internal/pairing"
	"github.com/adeelahmad/gadgetcast/internal/session"
	"github.com/adeelahmad/gadgetcast/internal/switcher"
)

const (
	writeTimeout   = 5 * time.Second
	maxBodyBytes   = 1 << 20
	maxFrameBytes  = 256 << 10
	sendQueueDepth = 64
)

// Options configures a Server. Zero values select sensible defaults.
type Options struct {
	// DataDir is where pipelines are persisted (DataDir/pipelines/<id>.json). Required.
	DataDir string
	// PIN is the expected pairing PIN; PINAttemptLimit the lockout threshold (default 5).
	PIN             string
	PINAttemptLimit int
	// MetricsInterval is the /api/events metrics period (default 2s).
	MetricsInterval time.Duration
	// ProcRoot, SysRoot, DevRoot override /proc, /sys, /dev (tests).
	ProcRoot, SysRoot, DevRoot string
	// Static is the built UI (web/dist contents). Nil serves the "no UI" notice.
	Static fs.FS
	// OriginPatterns are extra cross-origin hosts allowed on WebSockets
	// (e.g. "localhost:*" for a UI dev server). Same-origin is always allowed.
	OriginPatterns []string
	// Logger receives server log lines; nil discards them.
	Logger *log.Logger
}

// Server holds all server state. Create with New; serve Handler().
type Server struct {
	opts     Options
	logger   *log.Logger
	store    *PipelineStore
	metrics  *MetricsCollector
	devices  *DeviceScanner
	events   *broker
	mux      *http.ServeMux
	done     chan struct{}
	stopOnce sync.Once

	// hubMu guards the session hub, its switcher, and the conn routing table.
	hubMu sync.Mutex
	hub   *session.Hub
	sw    *switcher.Switcher
	conns map[core.ClientID]*sigConn
	order []core.ClientID

	// mu guards the running pipeline and the last published status key.
	mu        sync.Mutex
	running   *string
	statusKey string
}

// New builds a Server, opening (and seeding) the pipeline store.
func New(opts Options) (*Server, error) {
	if opts.DataDir == "" {
		return nil, errors.New("server: DataDir is required")
	}
	if opts.PINAttemptLimit <= 0 {
		opts.PINAttemptLimit = 5
	}
	if opts.MetricsInterval <= 0 {
		opts.MetricsInterval = 2 * time.Second
	}
	if opts.ProcRoot == "" {
		opts.ProcRoot = "/proc"
	}
	if opts.SysRoot == "" {
		opts.SysRoot = "/sys"
	}
	if opts.DevRoot == "" {
		opts.DevRoot = "/dev"
	}
	store, err := OpenPipelineStore(filepath.Join(opts.DataDir, "pipelines"))
	if err != nil {
		return nil, err
	}
	pm := pairing.New(opts.PIN, opts.PINAttemptLimit)
	sw := switcher.New(pm)
	s := &Server{
		opts:    opts,
		logger:  opts.Logger,
		store:   store,
		metrics: NewMetricsCollector(opts.ProcRoot, opts.SysRoot),
		devices: NewDeviceScanner(opts.DevRoot, opts.ProcRoot, opts.SysRoot),
		events:  newBroker(),
		done:    make(chan struct{}),
		hub:     session.New(pm, sw),
		sw:      sw,
		conns:   make(map[core.ClientID]*sigConn),
	}
	s.routes()
	return s, nil
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler { return s.mux }

// Close tells all open WebSocket handlers to close. http.Server.Shutdown does
// not track hijacked connections, so call this on shutdown.
func (s *Server) Close() {
	s.stopOnce.Do(func() { close(s.done) })
}

func (s *Server) routes() {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/health", s.handleHealth)
	m.HandleFunc("GET /api/nodes", s.handleNodes)
	m.HandleFunc("GET /api/pipelines", s.handleListPipelines)
	m.HandleFunc("GET /api/pipelines/{id}", s.handleGetPipeline)
	m.HandleFunc("PUT /api/pipelines/{id}", s.handlePutPipeline)
	m.HandleFunc("DELETE /api/pipelines/{id}", s.handleDeletePipeline)
	m.HandleFunc("POST /api/pipelines/{id}/run", s.handleRunPipeline)
	m.HandleFunc("POST /api/pipelines/stop", s.handleStopPipeline)
	m.HandleFunc("GET /api/status", s.handleStatus)
	m.HandleFunc("GET /api/devices", s.handleDevices)
	m.HandleFunc("GET /api/events", s.handleEvents)
	m.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "no such endpoint: "+r.Method+" "+r.URL.Path)
	})
	m.HandleFunc("GET /ws", s.handleSignaling)
	m.Handle("/", s.staticHandler())
	s.mux = m
}

func (s *Server) acceptOptions() *websocket.AcceptOptions {
	return &websocket.AcceptOptions{OriginPatterns: s.opts.OriginPatterns}
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// pathID extracts and validates the {id} path segment, writing a 400 if invalid.
func pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !validID(id) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid id %q: must match ^[a-zA-Z0-9_-]{1,64}$", id))
		return "", false
	}
	return id, true
}

// ---- status ----

// Status returns the current run state with a fresh metrics sample.
func (s *Server) Status() Status {
	st := s.statusNoMetrics()
	st.Metrics = s.metrics.Sample()
	return st
}

func (s *Server) statusNoMetrics() Status {
	s.hubMu.Lock()
	clients := len(s.order)
	var onAir *string
	if id, live := s.sw.OnAir(); live {
		v := id.String()
		onAir = &v
	}
	s.hubMu.Unlock()

	s.mu.Lock()
	var running *string
	if s.running != nil {
		v := *s.running
		running = &v
	}
	s.mu.Unlock()
	return Status{Running: running, OnAir: onAir, Clients: clients, Metrics: Metrics{LoadAvg: []float64{}}}
}

// publishStatusIfChanged pushes a "status" event when running / onAir / clients changed.
func (s *Server) publishStatusIfChanged() {
	st := s.statusNoMetrics()
	key, _ := json.Marshal(st)
	s.mu.Lock()
	changed := string(key) != s.statusKey
	s.statusKey = string(key)
	s.mu.Unlock()
	if changed {
		st.Metrics = s.metrics.Sample()
		s.events.publish(newEvent("status", st))
	}
}

func (s *Server) clients() []Client {
	s.hubMu.Lock()
	defer s.hubMu.Unlock()
	onAir, live := s.sw.OnAir()
	out := make([]Client, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, Client{ID: id.String(), OnAir: live && id == onAir})
	}
	return out
}

// ---- REST handlers ----

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": Version})
}

func (s *Server) handleNodes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, nodeCatalogue)
}

func (s *Server) handleListPipelines(w http.ResponseWriter, _ *http.Request) {
	list, err := s.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetPipeline(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	p, err := s.store.Get(id)
	if errors.Is(err, ErrPipelineNotFound) {
		writeError(w, http.StatusNotFound, "pipeline "+id+" not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handlePutPipeline(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var p Pipeline
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid pipeline JSON: "+err.Error())
		return
	}
	if p.ID != id {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("body id %q does not match path id %q", p.ID, id))
		return
	}
	saved, err := s.store.Put(p)
	if errors.Is(err, ErrInvalidPipeline) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.logEvent("info", "pipeline", "saved pipeline "+id)
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) handleDeletePipeline(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	err := s.store.Delete(id)
	if errors.Is(err, ErrPipelineNotFound) {
		writeError(w, http.StatusNotFound, "pipeline "+id+" not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.mu.Lock()
	if s.running != nil && *s.running == id {
		s.running = nil
	}
	s.mu.Unlock()
	s.logEvent("info", "pipeline", "deleted pipeline "+id)
	s.publishStatusIfChanged()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRunPipeline(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.Get(id); err != nil {
		if errors.Is(err, ErrPipelineNotFound) {
			writeError(w, http.StatusNotFound, "pipeline "+id+" not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.mu.Lock()
	s.running = &id
	s.mu.Unlock()
	s.logEvent("info", "pipeline", "running pipeline "+id+" (no media engine yet)")
	s.publishStatusIfChanged()
	writeJSON(w, http.StatusOK, s.Status())
}

func (s *Server) handleStopPipeline(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	was := s.running
	s.running = nil
	s.mu.Unlock()
	if was != nil {
		s.logEvent("info", "pipeline", "stopped pipeline "+*was)
	}
	s.publishStatusIfChanged()
	writeJSON(w, http.StatusOK, s.Status())
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Status())
}

func (s *Server) handleDevices(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, devicesResponse{Devices: s.devices.Scan(), Clients: s.clients()})
}
