// Package config parses and validates the GadgetCast configuration document.
// Stdlib encoding/json only.
package config

// Bind-interface modes.
const (
	BindAll           = "all"
	BindTailscaleOnly = "tailscale-only"
)

// HTTPSConfig models the TLS precondition (README §3.2/§4).
type HTTPSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}

// BindConfig models the listener bind interface + target.
type BindConfig struct {
	Mode   string `json:"mode"`
	Target string `json:"target"`
}

// PINConfig models the PIN pairing knobs (README §4, decision #3/#6).
type PINConfig struct {
	Required     bool   `json:"required"`
	AttemptLimit int    `json:"attempt_limit"`
	Source       string `json:"source"`
}

// Resolution is one video resolution/frame-rate entry.
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	FPS    int `json:"fps"`
}

// VideoConfig models the video format ladder (README §3.1).
type VideoConfig struct {
	PrimaryFormat  string       `json:"primary_format"`
	FallbackFormat string       `json:"fallback_format"`
	Resolutions    []Resolution `json:"resolutions"`
}

// AudioConfig models the audio codec/sample-rate (README §3.2/§4).
type AudioConfig struct {
	Codec      string `json:"codec"`
	SampleRate int    `json:"sample_rate"`
}

// Config is the parsed+defaulted GadgetCast configuration.
type Config struct {
	HTTPS HTTPSConfig `json:"https"`
	Bind  BindConfig  `json:"bind"`
	PIN   PINConfig   `json:"pin"`
	Video VideoConfig `json:"video"`
	Audio AudioConfig `json:"audio"`
}

// Parse unmarshals a JSON configuration document, fills documented defaults for
// omitted fields, and returns a typed parse error on malformed input.
func Parse(_ []byte) (Config, error) {
	panic("SUB-AGENT-TODO: json.Unmarshal the doc, return a typed parse error on malformed JSON, then apply documented defaults (see defaults.go)")
}
