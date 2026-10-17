/* /send: a one-screen page for phones. Enter the PIN, tap once, and the host hears this phone's mic. */
import logo from "../ds/assets/logo.svg";
import logoDark from "../ds/assets/logo-dark.svg";
import { Badge, Button, Input, Tabs } from "../ds";
import { STATE_LABEL, sourceItems, useMicSender, type SendSource } from "./SendMic";

export function SendPage() {
  const { pin, setPin, source, setSource, state, detail, supported, busy, start, stop } = useMicSender();
  const dark = document.documentElement.getAttribute("data-theme") === "dark";
  const msg = !supported ? "This browser cannot send audio. Use Chrome, Safari or Firefox over https://."
    : detail ?? (state === "on-air" ? (source === "mic" ? "The computer is hearing this phone." : "The computer is seeing and hearing this phone.") + " Keep this page open." : state === "connecting" ? "Connecting…" : state === "error" ? "" : "Pick GadgetCast as the microphone or camera on the computer.");
  return (
    <main style={{ minHeight: "100dvh", display: "flex", alignItems: "center", justifyContent: "center", padding: "var(--sp-8) 16px", background: "var(--surface-canvas)" }}>
      <div style={{ width: "100%", maxWidth: 420, display: "flex", flexDirection: "column", gap: "var(--sp-8)", alignItems: "stretch" }}>
        <img src={dark ? logoDark : logo} alt="GadgetCast" style={{ height: 28, alignSelf: "center" }} />
        <h1 style={{ margin: 0, textAlign: "center", font: "var(--type-title, var(--type-ui))", fontSize: 22, color: "var(--text-primary)" }}>Use this phone as a webcam or mic</h1>
        {!busy ? <Tabs<SendSource> size="sm" value={source} onChange={setSource} items={sourceItems()} /> : null}
        <Input label="Pairing PIN" value={pin} onChange={e => setPin(e.target.value)} inputMode="numeric" autoComplete="one-time-code" mono size="lg" disabled={busy} />
        {busy ? <Button variant="danger" size="lg" icon="stop" onClick={stop}>Stop</Button>
          : <Button variant="live" size="lg" icon="mic" onClick={start} disabled={!supported || pin.trim() === ""}>Start sending</Button>}
        <div role="status" aria-live="polite" style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: "var(--sp-4)", textAlign: "center" }}>
          <Badge tone={state === "on-air" ? "live" : state === "error" ? "err" : "neutral"} dot solid={state === "on-air"}>{STATE_LABEL[state]}</Badge>
          <span style={{ font: "var(--type-body)", color: state === "error" ? "var(--st-err)" : "var(--text-muted)" }}>{msg}</span>
        </div>
        <a href="/" style={{ alignSelf: "center", font: "var(--type-meta)", color: "var(--text-faint)" }}>Open GadgetCast Studio</a>
      </div>
    </main>
  );
}
