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
func applyDefaults(c Config) Config {
	if c.Video.PrimaryFormat == "" {
		c.Video.PrimaryFormat = DefaultVideoPrimaryFormat
	}
	if c.Video.FallbackFormat == "" {
		c.Video.FallbackFormat = DefaultVideoFallbackFormat
	}
	if len(c.Video.Resolutions) == 0 {
		c.Video.Resolutions = []Resolution{
			{Width: 1920, Height: 1080, FPS: 30},
			{Width: 1280, Height: 720, FPS: 30},
			{Width: 854, Height: 480, FPS: 30},
		}
	}
	if c.Audio.Codec == "" {
		c.Audio.Codec = DefaultAudioCodec
	}
	if c.Audio.SampleRate == 0 {
		c.Audio.SampleRate = DefaultAudioSampleRate
	}
	if c.PIN.AttemptLimit == 0 {
		c.PIN.AttemptLimit = DefaultPINAttemptLimit
	}
	if !c.PIN.Required {
		c.PIN.Required = DefaultPINRequired
	}
	return c
}
