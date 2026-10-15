package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Store errors.
var (
	ErrPipelineNotFound = errors.New("pipeline not found")
	ErrInvalidPipeline  = errors.New("invalid pipeline")
)

// reservedIDs cannot be used as pipeline ids because they collide with fixed routes.
var reservedIDs = map[string]bool{"stop": true}

// PipelineStore persists pipelines as one JSON file per id in a directory.
type PipelineStore struct {
	mu  sync.Mutex
	dir string
	now func() time.Time
}

// OpenPipelineStore opens (creating if needed) the store at dir and seeds the
// default pipeline when the directory holds no pipelines.
func OpenPipelineStore(dir string) (*PipelineStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("pipelines: create dir: %w", err)
	}
	s := &PipelineStore{dir: dir, now: time.Now}
	list, err := s.List()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		if _, err := s.Put(defaultPipeline()); err != nil {
			return nil, fmt.Errorf("pipelines: seed default: %w", err)
		}
	}
	return s, nil
}

func defaultPipeline() Pipeline {
	return Pipeline{
		ID:          "default",
		Name:        "Browser mic → USB mic",
		Description: "Route a browser microphone to the host as a USB (UAC2) microphone.",
		IsDefault:   true,
		Nodes: []PipelineNode{
			{ID: "n1", Type: "browser_mic", X: 80, Y: 120},
			{ID: "n2", Type: "uac2_out", X: 420, Y: 120},
		},
		Edges: []PipelineEdge{
			{ID: "e1", From: "n1", FromPort: 0, To: "n2", ToPort: 0, Kind: PortAudio},
		},
	}
}

func (s *PipelineStore) path(id string) string { return filepath.Join(s.dir, id+".json") }

// List returns all pipelines sorted by id.
func (s *PipelineStore) List() ([]Pipeline, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("pipelines: read dir: %w", err)
	}
	out := []Pipeline{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if !validID(id) {
			continue
		}
		p, err := s.readLocked(id)
		if err != nil {
			continue // skip unreadable / corrupt files rather than failing the list
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Get returns one pipeline or ErrPipelineNotFound.
func (s *PipelineStore) Get(id string) (Pipeline, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked(id)
}

func (s *PipelineStore) readLocked(id string) (Pipeline, error) {
	if !validID(id) {
		return Pipeline{}, ErrPipelineNotFound
	}
	data, err := os.ReadFile(s.path(id))
	if errors.Is(err, os.ErrNotExist) {
		return Pipeline{}, ErrPipelineNotFound
	}
	if err != nil {
		return Pipeline{}, fmt.Errorf("pipelines: read %s: %w", id, err)
	}
	var p Pipeline
	if err := json.Unmarshal(data, &p); err != nil {
		return Pipeline{}, fmt.Errorf("pipelines: decode %s: %w", id, err)
	}
	normalize(&p)
	return p, nil
}

// Put validates, stamps updatedAt, and atomically writes p (temp file + rename).
func (s *PipelineStore) Put(p Pipeline) (Pipeline, error) {
	normalize(&p)
	if err := validatePipeline(p); err != nil {
		return Pipeline{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return Pipeline{}, fmt.Errorf("pipelines: encode: %w", err)
	}
	tmp, err := os.CreateTemp(s.dir, "."+p.ID+".*.tmp")
	if err != nil {
		return Pipeline{}, fmt.Errorf("pipelines: temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		cleanup()
		return Pipeline{}, fmt.Errorf("pipelines: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		cleanup()
		return Pipeline{}, fmt.Errorf("pipelines: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return Pipeline{}, fmt.Errorf("pipelines: close: %w", err)
	}
	if err := os.Rename(tmpName, s.path(p.ID)); err != nil {
		cleanup()
		return Pipeline{}, fmt.Errorf("pipelines: rename: %w", err)
	}
	return p, nil
}

// Delete removes a pipeline, returning ErrPipelineNotFound if absent.
func (s *PipelineStore) Delete(id string) error {
	if !validID(id) {
		return ErrPipelineNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path(id))
	if errors.Is(err, os.ErrNotExist) {
		return ErrPipelineNotFound
	}
	if err != nil {
		return fmt.Errorf("pipelines: delete %s: %w", id, err)
	}
	return nil
}

func normalize(p *Pipeline) {
	if p.Nodes == nil {
		p.Nodes = []PipelineNode{}
	}
	if p.Edges == nil {
		p.Edges = []PipelineEdge{}
	}
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidPipeline, fmt.Sprintf(format, args...))
}

// validatePipeline checks the structural invariants of a pipeline graph.
func validatePipeline(p Pipeline) error {
	if !validID(p.ID) {
		return invalid("id %q must match ^[a-zA-Z0-9_-]{1,64}$", p.ID)
	}
	if reservedIDs[p.ID] {
		return invalid("id %q is reserved", p.ID)
	}
	if strings.TrimSpace(p.Name) == "" {
		return invalid("name is required")
	}
	nodes := make(map[string]bool, len(p.Nodes))
	for i, n := range p.Nodes {
		if n.ID == "" {
			return invalid("nodes[%d]: id is required", i)
		}
		if nodes[n.ID] {
			return invalid("nodes[%d]: duplicate id %q", i, n.ID)
		}
		if n.Type == "" {
			return invalid("nodes[%d]: type is required", i)
		}
		nodes[n.ID] = true
	}
	edges := make(map[string]bool, len(p.Edges))
	for i, e := range p.Edges {
		if e.ID == "" {
			return invalid("edges[%d]: id is required", i)
		}
		if edges[e.ID] {
			return invalid("edges[%d]: duplicate id %q", i, e.ID)
		}
		edges[e.ID] = true
		if !nodes[e.From] {
			return invalid("edges[%d]: from %q is not a node", i, e.From)
		}
		if !nodes[e.To] {
			return invalid("edges[%d]: to %q is not a node", i, e.To)
		}
		if e.FromPort < 0 || e.ToPort < 0 {
			return invalid("edges[%d]: port index must be >= 0", i)
		}
		if !validPortKind(e.Kind) {
			return invalid("edges[%d]: kind %q must be video|audio|ai|data", i, e.Kind)
		}
	}
	return nil
}
