package server

import (
	"encoding/json"
	"regexp"
)

// Version is the server version reported by /api/health.
const Version = "0.1.0"

// idPattern is the contract's id rule for path segments (docs/api.md).
var idPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// validID reports whether s is a valid pipeline / path id.
func validID(s string) bool { return idPattern.MatchString(s) }

// PortKind is the media kind carried by a node port or edge.
type PortKind string

// Port kinds (docs/api.md).
const (
	PortVideo PortKind = "video"
	PortAudio PortKind = "audio"
	PortAI    PortKind = "ai"
	PortData  PortKind = "data"
)

func validPortKind(k PortKind) bool {
	switch k {
	case PortVideo, PortAudio, PortAI, PortData:
		return true
	}
	return false
}

// Port is a [kind, format label] pair, serialised as a 2-element JSON array.
type Port [2]string

// NodeType is one entry of the node catalogue.
type NodeType struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	Cls      string `json:"cls"`
	Category string `json:"category"`
	Media    string `json:"media,omitempty"`
	Gst      string `json:"gst,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	URL      string `json:"url,omitempty"`
	Dev      string `json:"dev,omitempty"`
	In       []Port `json:"in,omitempty"`
	Out      []Port `json:"out,omitempty"`
}

// NodeGroup is a titled group of node types.
type NodeGroup struct {
	Group string     `json:"group"`
	Items []NodeType `json:"items"`
}

// PipelineNode is a placed node in a pipeline graph.
type PipelineNode struct {
	ID      string                     `json:"id"`
	Type    string                     `json:"type"`
	X       float64                    `json:"x"`
	Y       float64                    `json:"y"`
	Title   string                     `json:"title,omitempty"`
	Enabled *bool                      `json:"enabled,omitempty"`
	Config  map[string]json.RawMessage `json:"config,omitempty"`
}

// PipelineEdge connects an output port of one node to an input port of another.
type PipelineEdge struct {
	ID       string   `json:"id"`
	From     string   `json:"from"`
	FromPort int      `json:"fromPort"`
	To       string   `json:"to"`
	ToPort   int      `json:"toPort"`
	Kind     PortKind `json:"kind"`
}

// Pipeline is a saved node graph.
type Pipeline struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	IsDefault   bool           `json:"isDefault,omitempty"`
	Nodes       []PipelineNode `json:"nodes"`
	Edges       []PipelineEdge `json:"edges"`
	UpdatedAt   string         `json:"updatedAt,omitempty"`
}

// Metrics is a host resource snapshot.
type Metrics struct {
	CPUPct     float64   `json:"cpuPct"`
	RAMUsedMB  float64   `json:"ramUsedMB"`
	RAMTotalMB float64   `json:"ramTotalMB"`
	TempC      *float64  `json:"tempC"`
	UptimeS    float64   `json:"uptimeS"`
	LoadAvg    []float64 `json:"loadAvg"`
}

// Status is the server's overall run state.
type Status struct {
	Running *string `json:"running"`
	OnAir   *string `json:"onAir"`
	Clients int     `json:"clients"`
	Metrics Metrics `json:"metrics"`
}

// Device is a local capture/playback device.
type Device struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// Client is a connected signaling client.
type Client struct {
	ID    string `json:"id"`
	OnAir bool   `json:"onAir"`
}

// devicesResponse is the /api/devices body.
type devicesResponse struct {
	Devices []Device `json:"devices"`
	Clients []Client `json:"clients"`
}

// Event is one /api/events frame.
type Event struct {
	Type string `json:"type"`
	TS   string `json:"ts"`
	Data any    `json:"data"`
}

// LogData is the payload of a "log" event.
type LogData struct {
	Level     string `json:"level"`
	Subsystem string `json:"subsystem"`
	Msg       string `json:"msg"`
}
