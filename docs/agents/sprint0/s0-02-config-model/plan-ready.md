---
type: plan-ready
story: S0-02
from_red_at: 2026-07-09T22:42:38Z
---

# plan-ready — S0-02 config-model

RED verified on sprint0 (gate-red-verify green). Boxes tick ONLY on GREEN pass.

- [x] `internal/config/config_test.go::TestParse_WellFormed_NoErrors` — input: JSON with HTTPS on + cert/key, bind=all, PIN required, video MJPEG/YUYV at 1080p30/720p30/480p30, audio Opus/48k; act: `Parse` then `Validate`; assert: no error (intake success #5).
- [x] `internal/config/config_test.go::TestParse_MalformedJSON_Errors` — input: `{"https": ` (truncated); act: `Parse`; assert: typed parse error, no panic.
- [x] `internal/config/validate_test.go::TestValidate_HTTPSDisabled_Errors` — input: config with HTTPS disabled; act: `Validate`; assert: error (intake failure #5 — HTTPS is a hard precondition).
- [x] `internal/config/validate_test.go::TestValidate_HTTPSEnabled_MissingCertKey_Errors` — input: HTTPS on, empty cert/key paths; act: `Validate`; assert: error (intake failure #5).
- [x] `internal/config/validate_test.go::TestValidate_HTTPSEnabled_WithCertKey_OK` — input: HTTPS on with cert+key set; act: `Validate`; assert: no error on the HTTPS dimension.
- [x] `internal/config/validate_test.go::TestValidate_TailscaleOnly_AllInterfacesTarget_Errors` — input: bind=tailscale-only with an all-interfaces target (e.g. `0.0.0.0`); act: `Validate`; assert: error, not a silent bind-all (intake failure #6).
- [x] `internal/config/validate_test.go::TestValidate_TailscaleOnly_EmptyTarget_Errors` — input: bind=tailscale-only with an empty/unresolvable target; act: `Validate`; assert: error (intake failure #6).
- [x] `internal/config/validate_test.go::TestValidate_TailscaleOnly_WithTailscaleTarget_OK` — input: bind=tailscale-only with a tailscale-shaped target (e.g. `100.64.x.x`); act: `Validate`; assert: no error on the bind dimension (intake success #5).
- [x] `internal/config/validate_test.go::TestValidate_BindAll_OK` — input: bind=all; act: `Validate`; assert: no error on the bind dimension.
- [x] `internal/config/validate_test.go::TestValidate_PINRequired_IndependentOfBind` — input: bind=tailscale-only with PIN required=true; act: `Validate`; assert: PIN-required stays true and is not auto-disabled (decision #6).
- [x] `internal/config/validate_test.go::TestValidate_PINRequired_MissingPINMaterial_Errors` — input: PIN required=true but no PIN source configured; act: `Validate`; assert: error (README §4 — PIN pairing must be enforceable).
- [x] `internal/config/defaults_test.go::TestDefaults_Video_MJPEGPrimary_YUYVFallback` — input: config omitting video formats; act: `Parse`/default-fill; assert: primary MJPEG, fallback YUYV (§3.1).
- [x] `internal/config/defaults_test.go::TestDefaults_Video_Resolutions_1080_720_480_at30` — input: config omitting resolutions; act: default-fill; assert: 1080p30, 720p30, 480p30 present (§3.1).
- [x] `internal/config/defaults_test.go::TestDefaults_Audio_Opus48k` — input: config omitting audio; act: default-fill; assert: Opus codec at 48 kHz (§3.2/§4).
- [x] `internal/config/defaults_test.go::TestDefaults_AttemptLimit_Is5` — input: config omitting the PIN attempt limit; act: default-fill; assert: attempt limit == 5 (decision #3).
- [x] `internal/config/defaults_test.go::TestParse_OmittedFields_UseDefaults` — input: a minimal JSON doc `{}` (or HTTPS-only); act: `Parse`; assert: all defaulted fields (formats, attempt limit, PIN-required) match the documented defaults.
