// agentic:shim — minimal WRONG-body compile shim so RED tests fail by assertion.
// Replaced by real production code during SCAFFOLD/GREEN. Stdlib encoding/json only.
package config

// Bind-interface modes.
const (
	BindAll           = "all"
	BindTailscaleOnly = "tailscale-only"
)

// Documented defaults (README §3.1/§3.2/§4, decision #3).
const (
	DefaultVideoPrimaryFormat  = "MJPEG"
	DefaultVideoFallbackFormat = "YUYV"
	DefaultAudioCodec          = "Opus"
	DefaultAudioSampleRate     = 48000
	DefaultPINAttemptLimit     = 5
	DefaultPINRequired         = true
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

// Parse is a shim with a WRONG body: it ignores its input, applies no defaults,
// and never reports a parse error — so every RED test fails by assertion.
func Parse(_ []byte) (Config, error) {
	return Config{}, nil
}

// Validate is a shim with a WRONG body: it always reports success, so every
// negative-path RED test fails by assertion.
func (c Config) Validate() error {
	return nil
}
