---
type: plan-ready
story: S1-01
from_red_at: 2026-07-10T00:22:20Z
---

# plan-ready — S1-01 session-hub

RED verified on sprint1 (gate-red-verify green). Boxes tick ONLY on GREEN pass.

- [ ] `internal/session/session_test.go::TestJoin_AdmitsClientAndRegistersPairing` — input: fresh hub; act: `Process(Inbound{From:A, Data:encode(join A)})`; assert: `sw.Has(A)` true, `pm.State(A)==StatePending`, no on-air change, effects empty (D1; success #1).
- [ ] `internal/session/session_test.go::TestJoin_DuplicateClient_ErrorEffect` — input: A already joined; act: `Process` a second join for A; assert: one `Effect{To:A, Msg:ErrorMessage}` (surfacing `switcher.ErrDuplicateClient`), switcher set unchanged (failure #3).
- [ ] `internal/session/session_test.go::TestJoin_UnpairedClient_VisibleButNotPaired` — input: fresh hub; act: join A only (no PIN); assert: `sw.Has(A)` true (visible) and `pm.IsPaired(A)` false (admission gates on-air only, D1; success #4).
- [ ] `internal/session/session_test.go::TestSubmitPIN_Correct_PairsClient` — input: A joined; act: `Process(SubmitPIN{A, correctPIN})`; assert: `pm.IsPaired(A)` true, effects empty (success #1).
- [ ] `internal/session/session_test.go::TestSubmitPIN_Incorrect_ErrorEffect` — input: A joined, limit 5; act: `Process(SubmitPIN{A, "wrong"})`; assert: one `Effect{To:A, ErrorMessage}` (ErrIncorrectPIN), `pm.State(A)==StatePending` (success #5).
- [ ] `internal/session/session_test.go::TestSubmitPIN_ExhaustLimit_DurablyRejected` — input: A joined, limit 3; act: three wrong PINs then a fourth submission; assert: `pm.State(A)==StateRejected`, the post-limit submission returns an `ErrorMessage` (ErrRejected) — durable refusal (success #5).
- [ ] `internal/session/session_test.go::TestSubmitPIN_UnknownClient_ErrorEffect` — input: fresh hub (Z never joined); act: `Process(SubmitPIN{Z, correctPIN})`; assert: one `Effect{To:Z, ErrorMessage}` (pairing.ErrUnknownClient), no panic (failure #5).
- [ ] `internal/session/session_test.go::TestSwitchSource_PairedClient_GoesOnAirAndBroadcasts` — input: A joined+paired; act: `Process(Inbound{From:A, Data:encode(switch-source A)})`; assert: `sw.OnAir()==(A,true)` and effects include `OnAirStateFrom(A,true)` addressed to A (success #2).
- [ ] `internal/session/session_test.go::TestSwitchSource_UnpairedTarget_ErrorEffect_NotOnAir` — input: A joined but NOT paired; act: switch-source naming A; assert: one `Effect{To:A, ErrorMessage}` (switcher.ErrNotPaired), `sw.OnAir()` live=false — A never reaches on-air (failure #2; invariant unpaired-cannot-go-on-air).
- [ ] `internal/session/session_test.go::TestSwitchSource_UnknownTarget_ErrorEffect` — input: A joined+paired, target G not connected; act: switch-source naming G; assert: `Effect{To:A, ErrorMessage}` (switcher.ErrUnknownClient), on-air unchanged (failure #4).
- [ ] `internal/session/session_test.go::TestSwitchSource_BroadcastAddressedToAllConnected` — input: A,B,C all joined+paired; act: switch-source naming A; assert: effects are exactly one `OnAirStateFrom(A,true)` per connected client (To A, To B, To C) in roster order — the exact Effect list (D5).
- [ ] `internal/session/session_test.go::TestDisconnect_OnAirClient_ClearsAndBroadcastsNobody` — input: A,B joined+paired, A on-air; act: `Process(Disconnect{A})`; assert: `sw.OnAir()==(zero,false)`, `sw.Has(A)` false, effects include `OnAirStateFrom(zero,false)` addressed to every remaining client (To B) (failure #1; success #3 — the worked example).
- [ ] `internal/session/session_test.go::TestDisconnect_NonOnAirClient_NoBroadcast` — input: A on-air, B present+paired; act: `Process(Disconnect{B})`; assert: `sw.OnAir()==(A,true)` unchanged, `sw.Has(B)` false, NO on-air-state Effect (success #6).
- [ ] `internal/session/session_test.go::TestDisconnect_UnknownClient_SafeNoop` — input: fresh hub (Z never joined); act: `Process(Disconnect{Z})`; assert: effects empty, no panic, hub/switcher state intact (failure #7).
- [ ] `internal/session/session_test.go::TestDisconnect_PairingStateRetained` — input: A joined+paired, then on-air; act: `Process(Disconnect{A})`; assert: `pm.State(A)==StatePaired` still (pairing durable, no deregister — D4).
- [ ] `internal/session/session_test.go::TestInbound_MalformedBytes_ErrorEffectNoPanic` — input: A joined; act: `Process(Inbound{From:A, Data:[]byte("{not json")})`; assert: one `Effect{To:A, ErrorMessage}`, `Process` does not panic (failure #6; invariant decode-never-panics).
- [ ] `internal/session/session_test.go::TestInbound_UnknownType_ErrorEffectNoPanic` — input: A joined; act: Inbound with an envelope of an unrecognized `type`; assert: `Effect{To:A, ErrorMessage}`, no panic (failure #6).
- [ ] `internal/session/session_test.go::TestInbound_UnexpectedServerMessage_ErrorEffect` — input: A joined; act: Inbound decoding to an `on-air-state` message (a server→client type sent by a client); assert: `Effect{To:A, ErrorMessage}` (unexpected type), no state change.
- [ ] `internal/session/session_test.go::TestInbound_OfferRelayedOpaquelyToPeers` — input: A,B,C joined; act: `Process(Inbound{From:A, Data:encode(offer SDP)})`; assert: effects relay the SAME `OfferMessage` (SDP unmodified) addressed to B and C only, never to A (D3).
- [ ] `internal/session/session_test.go::TestInbound_IceCandidateRelayedOpaquely` — input: A,B joined; act: Inbound ice-candidate from A; assert: one `Effect{To:B, IceCandidateMessage}` with the candidate string unmodified, none to A (D3).
