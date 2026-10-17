/* Send this device's microphone, and optionally its camera or screen, to the Pi: the host computer sees them
   as the GadgetCast USB mic and webcam. */
import { useEffect, useRef, useState } from "react";
import { Badge, Button, Icon, Input, Modal, Switch, Tabs, type TabItem } from "../ds";
import { wsUrl } from "../api/events";
import { startMicSender, type SendState, type SenderPeer, type SenderSocket } from "../api/sender";

const PIN_KEY = "gadgetcast.pin";
const readPin = () => { try { return sessionStorage.getItem(PIN_KEY) ?? ""; } catch { return ""; } };
const savePin = (p: string) => { try { sessionStorage.setItem(PIN_KEY, p); } catch { /* private mode */ } };

export type SendSource = "mic" | "camera" | "screen";

const MIC: MediaTrackConstraints = { echoCancellation: true, noiseSuppression: true, autoGainControl: true };
const canShareScreen = () => typeof navigator !== "undefined" && !!navigator.mediaDevices && "getDisplayMedia" in navigator.mediaDevices;

/** Source choices this browser supports (phones cannot share their screen). */
export function sourceItems(): TabItem<SendSource>[] {
  const items: TabItem<SendSource>[] = [{ value: "mic", label: "Mic", icon: "mic" }, { value: "camera", label: "Camera + mic", icon: "videocam" }];
  if (canShareScreen()) items.push({ value: "screen", label: "Screen + mic", icon: "screen_share" });
  return items;
}

async function openSource(src: SendSource): Promise<MediaStream> {
  if (src === "mic") return navigator.mediaDevices.getUserMedia({ audio: MIC });
  if (src === "camera") {
    return navigator.mediaDevices.getUserMedia({ audio: MIC, video: { width: { ideal: 1280 }, height: { ideal: 720 }, frameRate: { ideal: 30 }, facingMode: "user" } });
  }
  const screen = await navigator.mediaDevices.getDisplayMedia({ video: { frameRate: { ideal: 30 } } });
  try {
    const mic = await navigator.mediaDevices.getUserMedia({ audio: MIC });
    return new MediaStream([...screen.getVideoTracks(), ...mic.getAudioTracks()]);
  } catch {
    return screen; // screen only, if the mic was refused
  }
}

/* Plays the host's audio. The element starts playing inside the Start tap, because phones only allow
   audio playback from a user gesture; the track arrives later. */
function hostAudioPlayer(): (track: MediaStreamTrack) => () => void {
  const stream = new MediaStream();
  const audio = new Audio();
  audio.srcObject = stream;
  void audio.play().catch(() => { /* retried when the track arrives */ });
  return track => {
    stream.addTrack(track);
    void audio.play().catch(() => { /* the browser blocked autoplay */ });
    return () => { audio.pause(); stream.removeTrack(track); };
  };
}

export const HEAR_LABEL = "Play the computer's sound here";
export const HEAR_HINT = "Hear the call through this device. Use headphones to avoid echo.";

const LABEL: Record<SendState, string> = { idle: "Off", connecting: "Connecting", "on-air": "On air", error: "Stopped", stopped: "Off" };

/** Sender state shared by the Studio dialog and the phone page (/send). */
export function useMicSender() {
  const [pin, setPin] = useState(readPin);
  const [source, setSource] = useState<SendSource>("mic");
  const [hear, setHear] = useState(false);
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
      getMedia: () => openSource(source),
      play: hear ? hostAudioPlayer() : undefined,
    }, pin.trim(), { onState: (s, d) => { setState(s); setDetail(d); } });
  };
  const stop = () => { stopRef.current?.(); stopRef.current = null; };
  return { pin, setPin, source, setSource, hear, setHear, state, detail, supported, busy, start, stop };
}

export const STATE_LABEL = LABEL;

export function SendMic({ live, onClose }: { live: boolean; onClose: () => void }) {
  const { pin, setPin, source, setSource, hear, setHear, state, detail, supported, busy, start, stop } = useMicSender();
  const close = () => { stop(); onClose(); };

  const why = !live ? "Connect to a GadgetCast Pi to send audio; the offline demo has no device."
    : !supported ? "This browser cannot send audio (it needs WebRTC over https://)." : null;

  return (
    <Modal open eyebrow="This device · mic, camera or screen" title="Send to the USB host" icon="mic" width={480} onClose={close}
      footer={<>
        <span style={{ flex: 1, alignSelf: "center", font: "var(--type-meta)", color: "var(--text-faint)" }}>Keep this tab open while sending</span>
        <Button variant="ghost" onClick={close}>Close</Button>
        {busy ? <Button variant="danger" icon="stop" onClick={stop}>Stop</Button>
          : <Button variant="live" icon="mic" onClick={start} disabled={!!why || pin.trim() === ""}>Start sending</Button>}
      </>}>
      <span style={{ font: "var(--type-body)", color: "var(--text-body)" }}>
        The computer the Pi is plugged into sees <b>GadgetCast</b> as a microphone and webcam. Pick it in Zoom, Teams or OBS there.
      </span>
      {!busy ? <Tabs<SendSource> size="sm" value={source} onChange={setSource} items={sourceItems()} /> : null}
      <Input label="Pairing PIN" value={pin} onChange={e => setPin(e.target.value)} inputMode="numeric" autoComplete="one-time-code" mono
        hint="Printed by the installer, and in gadgetcast-pin.txt on the SD card" disabled={busy} />
      <Switch checked={hear} onChange={setHear} disabled={busy} label={HEAR_LABEL} description={HEAR_HINT} />
      <div role="status" aria-live="polite" style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)" }}>
        <Badge tone={state === "on-air" ? "live" : state === "error" ? "err" : "neutral"} dot solid={state === "on-air"}>{LABEL[state]}</Badge>
        <span style={{ font: "var(--type-meta)", color: state === "error" ? "var(--st-err)" : "var(--text-muted)" }}>
          {why ?? detail ?? (state === "on-air" ? (source === "mic" ? "The host is hearing this microphone." : "The host is seeing and hearing this device.") : state === "connecting" ? "Pairing and connecting…" : "")}
        </span>
      </div>
      <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>On a phone, open <b>{window.location.host}/send</b> for a one-button version.</span>
      {state === "error" ? <span style={{ display: "flex", gap: "var(--sp-3)", alignItems: "center", font: "var(--type-meta)", color: "var(--text-faint)" }}><Icon name="info" size="sm" />Check the PIN, then start again.</span> : null}
    </Modal>
  );
}
