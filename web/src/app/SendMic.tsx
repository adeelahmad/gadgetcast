/* Send this device's microphone to the Pi: the host computer hears it as the GadgetCast USB mic. */
import { useEffect, useRef, useState } from "react";
import { Badge, Button, Icon, Input, Modal } from "../ds";
import { wsUrl } from "../api/events";
import { startMicSender, type SendState, type SenderPeer, type SenderSocket } from "../api/sender";

const PIN_KEY = "gadgetcast.pin";
const readPin = () => { try { return sessionStorage.getItem(PIN_KEY) ?? ""; } catch { return ""; } };
const savePin = (p: string) => { try { sessionStorage.setItem(PIN_KEY, p); } catch { /* private mode */ } };

const LABEL: Record<SendState, string> = { idle: "Off", connecting: "Connecting", "on-air": "On air", error: "Stopped", stopped: "Off" };

/** Sender state shared by the Studio dialog and the phone page (/send). */
export function useMicSender() {
  const [pin, setPin] = useState(readPin);
  const [state, setState] = useState<SendState>("idle");
  const [detail, setDetail] = useState<string | undefined>();
  const stopRef = useRef<(() => void) | null>(null);
  useEffect(() => () => stopRef.current?.(), []);

  const supported = typeof RTCPeerConnection !== "undefined" && !!navigator.mediaDevices;
  const busy = state === "connecting" || state === "on-air";

  const start = () => {
    savePin(pin);
    setDetail(undefined);
    stopRef.current = startMicSender({
      url: wsUrl("/ws"),
      socket: url => new WebSocket(url) as unknown as SenderSocket,
      peer: () => new RTCPeerConnection() as unknown as SenderPeer,
      getMic: () => navigator.mediaDevices.getUserMedia({ audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true } }),
    }, pin.trim(), { onState: (s, d) => { setState(s); setDetail(d); } });
  };
  const stop = () => { stopRef.current?.(); stopRef.current = null; };
  return { pin, setPin, state, detail, supported, busy, start, stop };
}

export const STATE_LABEL = LABEL;

export function SendMic({ live, onClose }: { live: boolean; onClose: () => void }) {
  const { pin, setPin, state, detail, supported, busy, start, stop } = useMicSender();
  const close = () => { stop(); onClose(); };

  const why = !live ? "Connect to a GadgetCast Pi to send audio; the offline demo has no device."
    : !supported ? "This browser cannot send audio (it needs WebRTC over https://)." : null;

  return (
    <Modal open eyebrow="This device · microphone" title="Send my mic to the USB host" icon="mic" width={480} onClose={close}
      footer={<>
        <span style={{ flex: 1, alignSelf: "center", font: "var(--type-meta)", color: "var(--text-faint)" }}>Keep this tab open while sending</span>
        <Button variant="ghost" onClick={close}>Close</Button>
        {busy ? <Button variant="danger" icon="stop" onClick={stop}>Stop</Button>
          : <Button variant="live" icon="mic" onClick={start} disabled={!!why || pin.trim() === ""}>Start sending</Button>}
      </>}>
      <span style={{ font: "var(--type-body)", color: "var(--text-body)" }}>
        The computer the Pi is plugged into hears this microphone as <b>GadgetCast</b>. Pick it as the microphone in Zoom, Teams or OBS there.
      </span>
      <Input label="Pairing PIN" value={pin} onChange={e => setPin(e.target.value)} inputMode="numeric" autoComplete="one-time-code" mono
        hint="Printed by the installer, and in gadgetcast-pin.txt on the SD card" disabled={busy} />
      <div role="status" aria-live="polite" style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)" }}>
        <Badge tone={state === "on-air" ? "live" : state === "error" ? "err" : "neutral"} dot solid={state === "on-air"}>{LABEL[state]}</Badge>
        <span style={{ font: "var(--type-meta)", color: state === "error" ? "var(--st-err)" : "var(--text-muted)" }}>
          {why ?? detail ?? (state === "on-air" ? "The host is hearing this microphone." : state === "connecting" ? "Pairing and connecting…" : "")}
        </span>
      </div>
      <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>On a phone, open <b>{window.location.host}/send</b> for a one-button version.</span>
      {state === "error" ? <span style={{ display: "flex", gap: "var(--sp-3)", alignItems: "center", font: "var(--type-meta)", color: "var(--text-faint)" }}><Icon name="info" size="sm" />Check the PIN, then start again.</span> : null}
    </Modal>
  );
}
