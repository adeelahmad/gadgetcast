---
type: tasks
story: S0-02
---

# Tasks — S0-02 config-model

Wave 2 (parallel). Package `internal/config`. Stdlib `encoding/json` only (decision #7).
Parses/validates the README §4/§3.1/§3.2 knobs. Stands up no listener, resolves no real
interface — it only represents and validates.

## T1 — Config struct + JSON parse

Define the config struct and `Parse([]byte) (Config, error)` in
`internal/config/config.go` covering: HTTPS on/off + cert/key paths, bind-interface
(all vs. tailscale-only) + target, PIN-required (bool), PIN attempt limit (int), and
video/audio format fields. Malformed JSON returns a typed parse error (never panics).

- Atomic: struct + `Parse`.
- Testable: `config_test.go` parses a well-formed doc and rejects malformed JSON.

## T2 — HTTPS + bind-interface validation

Add `internal/config/validate.go` with `Validate()`:
- HTTPS disabled → error; HTTPS enabled with no cert/key → error (README §3.2/§4 —
  HTTPS is a hard precondition; intake failure #5).
- `bind=tailscale-only` combined with an all-interfaces / unresolvable target → error,
  never a silent fall-back to bind-all (intake failure #6).
- `bind=all` and `bind=tailscale-only` with a tailscale-shaped target → OK.
- PIN-required is validated independently of bind interface — tailscale-only NEVER
  auto-disables PIN (decision #6).

- Atomic: the validation rules.
- Testable: `validate_test.go` covers each accept/reject branch.

## T3 — Format + attempt-limit defaults

Add `internal/config/defaults.go` (or default-fill in `Parse`) so omitted fields take
documented defaults: video MJPEG primary / YUYV fallback at 1080p30, 720p30, 480p30
(§3.1); audio Opus at 48 kHz (§3.2/§4); PIN attempt limit default 5 (decision #3);
PIN-required default true.

- Atomic: defaulting logic + the default constants.
- Testable: `defaults_test.go` asserts each default and that an omitted-fields doc fills
  them in.
