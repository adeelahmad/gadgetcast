package config

// Documented defaults (README §3.1/§3.2/§4, decision #3).
const (
	DefaultVideoPrimaryFormat  = "MJPEG"
	DefaultVideoFallbackFormat = "YUYV"
	DefaultAudioCodec          = "Opus"
	DefaultAudioSampleRate     = 48000
	DefaultPINAttemptLimit     = 5
	DefaultPINRequired         = true
)

// applyDefaults fills documented defaults into any config fields left at their
// zero value after unmarshalling (video formats/resolutions, audio codec/rate,
// PIN attempt limit and required flag).
func applyDefaults(_ Config) Config {
	panic("SUB-AGENT-TODO: fill zero-value fields with the Default* constants (MJPEG/YUYV, 1080p30/720p30/480p30, Opus/48k, attempt-limit 5, PIN-required true)")
}
