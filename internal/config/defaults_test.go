package config

import "testing"

func TestDefaults_Video_MJPEGPrimary_YUYVFallback(t *testing.T) {
	cfg, err := Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse({}) failed: %v", err)
	}
	if cfg.Video.PrimaryFormat != DefaultVideoPrimaryFormat {
		t.Errorf("Video.PrimaryFormat = %q, want %q (MJPEG primary, §3.1)", cfg.Video.PrimaryFormat, DefaultVideoPrimaryFormat)
	}
	if cfg.Video.FallbackFormat != DefaultVideoFallbackFormat {
		t.Errorf("Video.FallbackFormat = %q, want %q (YUYV fallback, §3.1)", cfg.Video.FallbackFormat, DefaultVideoFallbackFormat)
	}
}

func TestDefaults_Video_Resolutions_1080_720_480_at30(t *testing.T) {
	cfg, err := Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse({}) failed: %v", err)
	}
	for _, height := range []int{1080, 720, 480} {
		if !containsResolution(cfg.Video.Resolutions, height, 30) {
			t.Errorf("default resolutions missing %dp30 (§3.1); got %+v", height, cfg.Video.Resolutions)
		}
	}
}

func TestDefaults_Audio_Opus48k(t *testing.T) {
	cfg, err := Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse({}) failed: %v", err)
	}
	if cfg.Audio.Codec != DefaultAudioCodec {
		t.Errorf("Audio.Codec = %q, want %q (§3.2/§4)", cfg.Audio.Codec, DefaultAudioCodec)
	}
	if cfg.Audio.SampleRate != DefaultAudioSampleRate {
		t.Errorf("Audio.SampleRate = %d, want %d Hz (§3.2/§4)", cfg.Audio.SampleRate, DefaultAudioSampleRate)
	}
}

func TestDefaults_AttemptLimit_Is5(t *testing.T) {
	cfg, err := Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse({}) failed: %v", err)
	}
	if cfg.PIN.AttemptLimit != DefaultPINAttemptLimit {
		t.Errorf("PIN.AttemptLimit = %d, want %d (decision #3)", cfg.PIN.AttemptLimit, DefaultPINAttemptLimit)
	}
}

func TestParse_OmittedFields_UseDefaults(t *testing.T) {
	cfg, err := Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse({}) failed: %v", err)
	}
	if cfg.Video.PrimaryFormat != DefaultVideoPrimaryFormat {
		t.Errorf("Video.PrimaryFormat = %q, want %q", cfg.Video.PrimaryFormat, DefaultVideoPrimaryFormat)
	}
	if cfg.Video.FallbackFormat != DefaultVideoFallbackFormat {
		t.Errorf("Video.FallbackFormat = %q, want %q", cfg.Video.FallbackFormat, DefaultVideoFallbackFormat)
	}
	if cfg.Audio.Codec != DefaultAudioCodec {
		t.Errorf("Audio.Codec = %q, want %q", cfg.Audio.Codec, DefaultAudioCodec)
	}
	if cfg.Audio.SampleRate != DefaultAudioSampleRate {
		t.Errorf("Audio.SampleRate = %d, want %d", cfg.Audio.SampleRate, DefaultAudioSampleRate)
	}
	if cfg.PIN.AttemptLimit != DefaultPINAttemptLimit {
		t.Errorf("PIN.AttemptLimit = %d, want %d", cfg.PIN.AttemptLimit, DefaultPINAttemptLimit)
	}
	if cfg.PIN.Required != DefaultPINRequired {
		t.Errorf("PIN.Required = %v, want %v (default true)", cfg.PIN.Required, DefaultPINRequired)
	}
}
