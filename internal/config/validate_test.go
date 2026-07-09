package config

import "testing"

// mustParse parses a doc for a validation test; a parse failure is a test-setup
// error, not the behavior under test.
func mustParse(t *testing.T, doc string) Config {
	t.Helper()
	cfg, err := Parse([]byte(doc))
	if err != nil {
		t.Fatalf("setup Parse failed: %v", err)
	}
	return cfg
}

func TestValidate_HTTPSDisabled_Errors(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": false, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "all", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate accepted HTTPS-disabled config; HTTPS is a hard precondition (intake failure #5)")
	}
}

func TestValidate_HTTPSEnabled_MissingCertKey_Errors(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "", "key_path": ""},
  "bind": {"mode": "all", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate accepted HTTPS-enabled config with empty cert/key (intake failure #5)")
	}
}

func TestValidate_HTTPSEnabled_WithCertKey_OK(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "all", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if cfg.HTTPS.CertPath != "/certs/server.crt" || cfg.HTTPS.KeyPath != "/certs/server.key" {
		t.Fatalf("cert/key not parsed: cert=%q key=%q", cfg.HTTPS.CertPath, cfg.HTTPS.KeyPath)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate rejected HTTPS-on config with cert+key: %v", err)
	}
}

func TestValidate_TailscaleOnly_AllInterfacesTarget_Errors(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "tailscale-only", "target": "0.0.0.0"},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate silently accepted tailscale-only with all-interfaces target 0.0.0.0 (intake failure #6)")
	}
}

func TestValidate_TailscaleOnly_EmptyTarget_Errors(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "tailscale-only", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate accepted tailscale-only with empty/unresolvable target (intake failure #6)")
	}
}

func TestValidate_TailscaleOnly_WithTailscaleTarget_OK(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "tailscale-only", "target": "100.64.0.5"},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if cfg.Bind.Mode != BindTailscaleOnly {
		t.Fatalf("Bind.Mode = %q, want %q", cfg.Bind.Mode, BindTailscaleOnly)
	}
	if cfg.Bind.Target != "100.64.0.5" {
		t.Fatalf("Bind.Target = %q, want tailscale-shaped 100.64.0.5", cfg.Bind.Target)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate rejected tailscale-only with a tailscale-shaped target: %v", err)
	}
}

func TestValidate_BindAll_OK(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "all", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if cfg.Bind.Mode != BindAll {
		t.Fatalf("Bind.Mode = %q, want %q", cfg.Bind.Mode, BindAll)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate rejected bind=all config: %v", err)
	}
}

func TestValidate_PINRequired_IndependentOfBind(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "tailscale-only", "target": "100.64.0.5"},
  "pin": {"required": true, "attempt_limit": 5, "source": "env:GADGETCAST_PIN"}
}`)
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate rejected a valid tailscale-only + PIN-required config: %v", err)
	}
	if !cfg.PIN.Required {
		t.Errorf("PIN.Required = false after tailscale-only bind; tailscale must NEVER auto-disable PIN (decision #6)")
	}
}

func TestValidate_PINRequired_MissingPINMaterial_Errors(t *testing.T) {
	cfg := mustParse(t, `{
  "https": {"enabled": true, "cert_path": "/certs/server.crt", "key_path": "/certs/server.key"},
  "bind": {"mode": "all", "target": ""},
  "pin": {"required": true, "attempt_limit": 5, "source": ""}
}`)
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate accepted PIN-required config with no PIN material/source (README §4)")
	}
}
