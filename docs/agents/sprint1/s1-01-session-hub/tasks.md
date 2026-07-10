---
type: tasks
story: S1-01
---

# Tasks — S1-01 session-hub

Wave 1 (parallel). Package `internal/session`. Stdlib only; imports only `internal/core`,
`internal/signaling`, `internal/pairing`, `internal/switcher`. Single-threaded, event-
driven — no goroutines/channels/`time.Now()`. Constructed as `session.New(pm, sw)` over an
injected `*pairing.Manager` and `*switcher.Switcher` (tests build `sw := switcher.New(pm)`;
`pm.IsPaired` satisfies `switcher.PairingOracle` with zero adapter). `Process(Event)
[]Effect` is one call in, one deterministic Effect list out.

Event model (an `Event` interface with concrete variants): `Inbound{From core.ClientID;
Data []byte}` (raw signaling bytes), `SubmitPIN{Client core.ClientID; PIN string}`
(out-of-band pairing channel — PIN is not one of the seven wire messages), `Disconnect{
Client core.ClientID}`. Effect model: `Effect{To core.ClientID; Msg signaling.Message}`.
A broadcast expands to one Effect per roster client in roster order (D5). The hub keeps its
own ordered roster (switcher exposes no enumeration) mirroring switcher membership.

## T1 — Hub construction, event/effect types, and join admission

Define `session.New(pm *pairing.Manager, sw *switcher.Switcher) *Hub`, the `Event`/`Effect`
types, and `Process`. On an `Inbound` decoding to a `JoinMessage`: `switcher.AddClient`
(admit — visible) + `pairing.Register` + append to the roster. A join for a `ClientID`
already connected surfaces `switcher.ErrDuplicateClient` as an addressed `ErrorMessage`
Effect and leaves state unchanged (D1; intake failures #3; success #1/#4).

- Atomic: construction + the join-admit path + the duplicate-join guard + the roster.
- Testable: `session_test.go` covers join-admits-and-registers (visible in switcher, state
  Pending, no on-air change), duplicate-join-errors, and unpaired-join-is-visible-but-
  unpaired (admission timing D1).

## T2 — PIN pairing events

On a `SubmitPIN` event: `pairing.SubmitPIN(Client, PIN)` and map the outcome to Effects — a
correct PIN pairs the client (no broadcast; observe via `pm.IsPaired`); `ErrIncorrectPIN`,
`ErrRejected`, and `ErrUnknownClient` (a PIN for a `ClientID` that never joined) each yield
an addressed `ErrorMessage` Effect, never a panic or silent success (intake failures #5;
success #5).

- Atomic: the SubmitPIN dispatch + the four outcome→Effect mappings.
- Testable: `session_test.go` covers correct-PIN-pairs, incorrect-PIN-error-effect,
  exhaust-limit-durably-rejected, and unknown-client-PIN-error-effect (never-joined).

## T3 — switch-source on-air routing + on-air-state broadcast

On an `Inbound` decoding to a `SwitchSourceMessage{Target}`: `switcher.SetOnAir(Target)`.
On success, broadcast `signaling.OnAirStateFrom(switcher.OnAir())` expanded to one addressed
Effect per roster client (D5, decision #8). `ErrUnknownClient` (target not connected) and
`ErrNotPaired` (admitted but unpaired) each yield an addressed `ErrorMessage` Effect and
leave on-air unchanged — an unpaired target NEVER reaches on-air (intake failures #2/#4;
success #2/#4; invariant: unpaired-cannot-go-on-air).

- Atomic: the SetOnAir call + the broadcast fan-out + the unknown/unpaired guards.
- Testable: `session_test.go` covers paired-goes-on-air-and-broadcasts, unpaired-target-
  error-not-on-air, unknown-target-error, and broadcast-addressed-to-all-connected (exact
  Effect list in roster order).

## T4 — Disconnect handling

On a `Disconnect{Client}` event: if the client is not in the roster, a safe no-op (empty
Effects, no panic). Otherwise capture whether it was on-air (`switcher.OnAir()`), call
`switcher.RemoveClient` (which clears on-air per its S0-04 contract), and remove it from the
roster. If it WAS on-air, broadcast `signaling.OnAirStateFrom(core.ClientID{}, false)` to
every remaining roster client; if it was not, produce no on-air-state Effect. Pairing state
is retained — no deregister (D4; intake failures #1/#7; success #3/#6).

- Atomic: the roster/no-op guard + the was-on-air capture + RemoveClient + the conditional
  nobody broadcast.
- Testable: `session_test.go` covers disconnect-on-air-clears-and-broadcasts-nobody,
  disconnect-non-on-air-no-broadcast, disconnect-unknown-safe-noop, and pairing-state-
  retained-after-disconnect.

## T5 — Decode robustness + opaque offer/answer/ice relay

On an `Inbound` event the hub calls `signaling.Decode(Data)` and NEVER panics: a
`*DecodeError` (malformed bytes, unknown type) yields an addressed `ErrorMessage` Effect to
the sender; an unexpected server→client message arriving inbound (`on-air-state`, `error`)
yields an `ErrorMessage` Effect. A decoded `OfferMessage`/`AnswerMessage`/
`IceCandidateMessage` is relayed VERBATIM (opaque — SDP/ICE never parsed) as an addressed
Effect to every OTHER connected peer (D3; intake failure #6; invariant: decode-never-panics).

- Atomic: the decode-error path + the unexpected-type path + the opaque relay fan-out.
- Testable: `session_test.go` covers malformed-bytes-error-effect-no-panic, unknown-type-
  error-effect-no-panic, unexpected-server-message-error-effect, offer-relayed-opaquely-to-
  peers, and ice-candidate-relayed-opaquely.
