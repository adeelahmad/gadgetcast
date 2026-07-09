package config

import (
	"strings"
	"testing"
)

// validConfigJSON is a well-formed doc: HTTPS on with cert/key, bind=all, PIN
// required with a source, video MJPEG/YUYV at 1080p30/720p30/480p30, audio Opus/48k.
// Shared across the T1–T3 test files (same package).
func validConfigJSON() []byte {
	return []byte(`{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "all", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"},
  "video": {
    "primary_format": "MJPEG",
    "fallback_format": "YUYV",
    "resolutions": [
      {"width": 1920, "height": 1080, "fps": 30},
      {"width": 1280, "height": 720, "fps": 30},
      {"width": 854, "height": 480, "fps": 30}
    ]
  },
  "audio": {"codec": "Opus", "sample_rate": 48000}
}`)
}

// containsResolution reports whether rs has an entry at the given height and fps.
func containsResolution(rs []Resolution, height, fps int) bool {
	for _, r := range rs {
		if r.Height == height && r.FPS == fps {
			return true
		}
	}
	return false
}

func TestParse_WellFormed_NoErrors(t *testing.T) {
	cfg, err := Parse(validConfigJSON())
	if err != nil {
		t.Fatalf("Parse returned error on well-formed doc: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error on well-formed doc: %v", err)
	}
	if !cfg.HTTPS.Enabled {
		t.Errorf("HTTPS.Enabled = false, want true")
	}
	if cfg.Bind.Mode != BindAll {
		t.Errorf("Bind.Mode = %q, want %q", cfg.Bind.Mode, BindAll)
	}
	if cfg.Video.PrimaryFormat != DefaultVideoPrimaryFormat {
		t.Errorf("Video.PrimaryFormat = %q, want %q", cfg.Video.PrimaryFormat, DefaultVideoPrimaryFormat)
	}
	if cfg.Audio.SampleRate != DefaultAudioSampleRate {
		t.Errorf("Audio.SampleRate = %d, want %d", cfg.Audio.SampleRate, DefaultAudioSampleRate)
	}
}

func TestParse_MalformedJSON_Errors(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Parse panicked on malformed JSON: %v", r)
		}
	}()

	_, err := Parse([]byte(`{"https": `))
	if err == nil {
		t.Fatalf("Parse returned nil error on truncated/malformed JSON, want typed parse error")
	}
	if strings.TrimSpace(err.Error()) == "" {
		t.Errorf("Parse error message is empty, want a descriptive parse error")
	}
}
