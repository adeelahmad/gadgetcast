---
type: validate
story: S0-02
---

# Validate — S0-02 config-model

PASS/FAIL rubric mirroring the task headings. Literal commands, literal expected output.
Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] S0-00 merged: `go build ./internal/core/` exits 0.
- [ ] `go test ./internal/config/` compiles.
- [ ] `internal/config` uses stdlib `encoding/json` only:
      `go list -deps ./internal/config/ | grep -v '^internal/\|^github.com/adeelahmad/gadgetcast'`
      lists only standard-library paths (decision #7).

## T1 — parse

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| well-formed parses | `go test ./internal/config/ -run TestParse_WellFormed_NoErrors -v` | `--- PASS` | intake success #5 |
| malformed errors | `go test ./internal/config/ -run TestParse_MalformedJSON_Errors -v` | `--- PASS` | standards §Errors (no panic; typed error) |

## T2 — HTTPS + bind validation

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| HTTPS disabled rejected | `go test ./internal/config/ -run TestValidate_HTTPSDisabled_Errors -v` | `--- PASS` | intake failure #5 (HTTPS hard precondition) |
| HTTPS on, no cert/key rejected | `go test ./internal/config/ -run TestValidate_HTTPSEnabled_MissingCertKey_Errors -v` | `--- PASS` | intake failure #5 |
| HTTPS on with cert/key OK | `go test ./internal/config/ -run TestValidate_HTTPSEnabled_WithCertKey_OK -v` | `--- PASS` | intake success #5 |
| tailscale-only + all-interfaces rejected | `go test ./internal/config/ -run TestValidate_TailscaleOnly_AllInterfacesTarget_Errors -v` | `--- PASS` | intake failure #6 (no silent bind-all) |
| tailscale-only + empty target rejected | `go test ./internal/config/ -run TestValidate_TailscaleOnly_EmptyTarget_Errors -v` | `--- PASS` | intake failure #6 |
| tailscale-only + tailscale target OK | `go test ./internal/config/ -run TestValidate_TailscaleOnly_WithTailscaleTarget_OK -v` | `--- PASS` | intake success #5 |
| bind=all OK | `go test ./internal/config/ -run TestValidate_BindAll_OK -v` | `--- PASS` | intake success #5 |
| PIN independent of bind | `go test ./internal/config/ -run TestValidate_PINRequired_IndependentOfBind -v` | `--- PASS` | decision #6 (tailscale never auto-disables PIN) |
| PIN required needs material | `go test ./internal/config/ -run TestValidate_PINRequired_MissingPINMaterial_Errors -v` | `--- PASS` | README §4 (PIN enforceable) |

FAIL if a tailscale-only config with an all-interfaces target validates clean (silent
fall-back — intake failure #6), or if selecting tailscale-only flips PIN-required off
(decision #6).

## T3 — defaults

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| video defaults | `go test ./internal/config/ -run 'TestDefaults_Video' -v` | both `TestDefaults_Video_*` PASS | README §3.1 (MJPEG/YUYV, 1080/720/480p30) |
| audio default | `go test ./internal/config/ -run TestDefaults_Audio_Opus48k -v` | `--- PASS` | README §3.2/§4 (Opus 48 kHz) |
| attempt-limit default 5 | `go test ./internal/config/ -run TestDefaults_AttemptLimit_Is5 -v` | `--- PASS` | decision #3 (default 5) |
| omitted fields defaulted | `go test ./internal/config/ -run TestParse_OmittedFields_UseDefaults -v` | `--- PASS` | intake success #5 (defaults source) |

## Final sign-off

- [ ] Every T1–T3 row PASS; `go test ./internal/config/` exits 0.
- [ ] `gofmt -l internal/config` empty; `go vet ./internal/config/` clean.
- [ ] `go test ./internal/config/ -cover` ≥ 80%.
- [ ] `internal/config` imports only stdlib + `internal/core` (decision #7).
