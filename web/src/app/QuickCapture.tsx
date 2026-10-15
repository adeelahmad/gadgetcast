/* Quick capture: grabs a frame, a short clip or an audio snippet from this browser's camera, microphone or
   screen with getUserMedia / getDisplayMedia, entirely locally. Nothing is uploaded to the Pi: the v0 API
   has no capture endpoint. (Idea from the Gemini prototype, whose version faked the result.) */
import { useEffect, useRef, useState } from "react";
import { Badge, Button, Icon, Modal, Select, Tabs } from "../ds";
import type { CaptureMode } from "./store";

type Source = "camera" | "screen";
interface Result { url: string; name: string; size: number; kind: CaptureMode }

const stampName = () => new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19);
const fmtSize = (b: number) => (b > 1e6 ? (b / 1e6).toFixed(1) + " MB" : Math.max(1, Math.round(b / 1e3)) + " kB");

function pickMime(audioOnly: boolean): string | undefined {
  if (typeof MediaRecorder === "undefined") return undefined;
  const c = audioOnly ? ["audio/webm;codecs=opus", "audio/webm", "audio/mp4"] : ["video/webm;codecs=vp9,opus", "video/webm", "video/mp4"];
  return c.find(t => MediaRecorder.isTypeSupported(t));
}

export function QuickCapture({ mode: initial, onClose, onDone }: { mode: CaptureMode; onClose: () => void; onDone: (msg: string) => void }) {
  const [mode, setMode] = useState<CaptureMode>(initial);
  const [source, setSource] = useState<Source>("camera");
  const [secs, setSecs] = useState("5");
  const [stream, setStream] = useState<MediaStream | null>(null);
  const [busy, setBusy] = useState<null | "starting" | "recording">(null);
  const [left, setLeft] = useState(0);
  const [err, setErr] = useState<string | null>(null);
  const [result, setResult] = useState<Result | null>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const stopStream = () => { streamRef.current?.getTracks().forEach(t => t.stop()); streamRef.current = null; setStream(null); };
  const clearResult = () => setResult(r => { if (r) URL.revokeObjectURL(r.url); return null; });
  useEffect(() => () => { stopStream(); }, []);
  useEffect(() => () => { if (result) URL.revokeObjectURL(result.url); }, [result]);
  useEffect(() => { if (videoRef.current) videoRef.current.srcObject = stream; }, [stream]);

  const supported = typeof navigator !== "undefined" && !!navigator.mediaDevices;
  const start = async () => {
    setErr(null); clearResult(); setBusy("starting"); stopStream();
    try {
      if (!supported) throw new Error("Media capture needs a secure context (https:// or localhost).");
      let s: MediaStream;
      if (mode === "audio") s = await navigator.mediaDevices.getUserMedia({ audio: true });
      else if (source === "screen") s = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: mode === "video" });
      else s = await navigator.mediaDevices.getUserMedia({ video: { width: { ideal: 1920 }, height: { ideal: 1080 } }, audio: mode === "video" });
      streamRef.current = s; setStream(s);
      s.getVideoTracks()[0]?.addEventListener("ended", stopStream);
    } catch (e) {
      setErr(e instanceof Error ? (e.name === "NotAllowedError" ? "Permission denied by the browser." : e.message) : String(e));
    } finally { setBusy(null); }
  };

  const grabFrame = () => {
    const v = videoRef.current; if (!v || !v.videoWidth) { setErr("No video frame yet."); return; }
    const c = document.createElement("canvas");
    c.width = v.videoWidth; c.height = v.videoHeight;
    c.getContext("2d")?.drawImage(v, 0, 0);
    c.toBlob(b => {
      if (!b) { setErr("Could not encode the frame."); return; }
      setResult({ url: URL.createObjectURL(b), name: "gadgetcast-frame-" + stampName() + ".png", size: b.size, kind: "screenshot" });
    }, "image/png");
  };

  const record = () => {
    const s = streamRef.current; if (!s) return;
    if (typeof MediaRecorder === "undefined") { setErr("This browser cannot record media."); return; }
    const audioOnly = mode === "audio";
    const mime = pickMime(audioOnly);
    let rec: MediaRecorder;
    try { rec = new MediaRecorder(s, mime ? { mimeType: mime } : undefined); } catch (e) { setErr(e instanceof Error ? e.message : String(e)); return; }
    const chunks: Blob[] = [];
    rec.ondataavailable = e => { if (e.data.size) chunks.push(e.data); };
    rec.onstop = () => {
      const type = rec.mimeType || mime || (audioOnly ? "audio/webm" : "video/webm");
      const b = new Blob(chunks, { type });
      const ext = type.includes("mp4") ? (audioOnly ? "m4a" : "mp4") : "webm";
      setResult({ url: URL.createObjectURL(b), name: "gadgetcast-" + (audioOnly ? "audio" : "clip") + "-" + stampName() + "." + ext, size: b.size, kind: mode });
      setBusy(null);
    };
    const total = Number(secs);
    setBusy("recording"); setLeft(total);
    rec.start(250);
    let t = total;
    const tick = setInterval(() => { t -= 1; setLeft(t); if (t <= 0) { clearInterval(tick); if (rec.state !== "inactive") rec.stop(); } }, 1000);
  };

  const download = () => {
    if (!result) return;
    const a = document.createElement("a"); a.href = result.url; a.download = result.name; a.click();
    onDone("Saved " + result.name + " to this computer");
  };

  const close = () => { stopStream(); onClose(); };
  const wantsVideo = mode !== "audio";
  return (
    <Modal open eyebrow="Quick capture · this browser" title="Capture from camera, mic or screen" icon="photo_camera" width={600} onClose={close}
      footer={<>
        <span style={{ flex: 1, alignSelf: "center", font: "var(--type-meta)", color: "var(--text-faint)" }}>Stays on this computer · nothing is sent to the Pi</span>
        <Button variant="ghost" onClick={close}>Close</Button>
        {result ? <Button variant="primary" icon="download" onClick={download}>Download {result.kind === "screenshot" ? "PNG" : "clip"}</Button> : null}
      </>}>
      <Tabs<CaptureMode> size="sm" value={mode} onChange={m => { setMode(m); clearResult(); stopStream(); }}
        items={[{ value: "screenshot", label: "Screenshot", icon: "photo_camera" }, { value: "video", label: "Video clip", icon: "videocam" }, { value: "audio", label: "Audio clip", icon: "graphic_eq" }]} />
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "var(--sp-5)" }}>
        {wantsVideo ? <Select label="Source" value={source} options={[{ value: "camera", label: "Camera (getUserMedia)" }, { value: "screen", label: "Screen, window or tab" }]} onChange={e => { setSource(e.target.value === "screen" ? "screen" : "camera"); stopStream(); }} />
          : <Select label="Source" value="mic" options={[{ value: "mic", label: "Microphone (getUserMedia)" }]} disabled />}
        {mode !== "screenshot" ? <Select label="Length" value={secs} options={[{ value: "5", label: "5 s" }, { value: "10", label: "10 s" }, { value: "30", label: "30 s" }]} onChange={e => setSecs(e.target.value)} /> : <span />}
      </div>

      <div style={{ position: "relative", height: 240, borderRadius: "var(--r-md)", overflow: "hidden", border: "1px solid " + (busy === "recording" ? "var(--sig-live-dim)" : "var(--line-border)"), background: "var(--surface-input)", display: "grid", placeItems: "center" }}>
        {stream && wantsVideo ? <video ref={videoRef} autoPlay muted playsInline style={{ width: "100%", height: "100%", objectFit: "contain", background: "var(--ink-1000)" }} /> : null}
        {stream && !wantsVideo ? <span style={{ display: "flex", alignItems: "center", gap: "var(--sp-3)", font: "var(--type-meta)", color: "var(--sig-audio)" }}><Icon name="mic" size="md" />MICROPHONE OPEN</span> : null}
        {!stream ? <span style={{ font: "var(--type-meta)", color: "var(--text-disabled)", textAlign: "center" }}>NO SIGNAL<br /><span style={{ fontSize: "var(--fs-tick)" }}>open a source to preview it</span></span> : null}
        {busy === "recording" ? <span style={{ position: "absolute", top: 8, left: 8 }}><Badge tone="live" solid>REC · {left} s</Badge></span> : null}
      </div>

      {err ? <span role="alert" style={{ display: "flex", gap: "var(--sp-3)", alignItems: "center", font: "var(--type-ui)", color: "var(--st-err)" }}><Icon name="error" size="sm" />{err}</span> : null}
      {result ? (
        <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-4) var(--sp-5)", border: "1px solid var(--st-ok-dim)", borderRadius: "var(--r-sm)", background: "color-mix(in srgb, var(--st-ok) 6%, transparent)" }}>
          {result.kind === "screenshot" ? <img src={result.url} alt="Captured frame" style={{ height: 40, borderRadius: "var(--r-xs)", border: "1px solid var(--line-border)" }} /> : <Icon name={result.kind === "audio" ? "graphic_eq" : "movie"} size="lg" style={{ color: "var(--st-ok)" }} />}
          <span style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column" }}>
            <span style={{ font: "var(--type-meta)", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{result.name}</span>
            <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)" }}>{fmtSize(result.size)}</span>
          </span>
          {result.kind === "audio" ? <audio src={result.url} controls style={{ height: 28 }} /> : null}
        </div>
      ) : null}

      <div style={{ display: "flex", gap: "var(--sp-4)" }}>
        {!stream ? <Button variant="secondary" icon={mode === "audio" ? "mic" : source === "screen" ? "screen_share" : "videocam"} onClick={start} disabled={busy === "starting"}>{busy === "starting" ? "Waiting for permission…" : "Open source"}</Button>
          : <Button variant="ghost" icon="close" onClick={stopStream} disabled={busy === "recording"}>Release source</Button>}
        {stream && mode === "screenshot" ? <Button variant="primary" icon="photo_camera" onClick={grabFrame}>Capture frame</Button> : null}
        {stream && mode !== "screenshot" ? <Button variant="live" icon="fiber_manual_record" onClick={record} disabled={busy === "recording"}>{busy === "recording" ? "Recording…" : "Record " + secs + " s"}</Button> : null}
      </div>
    </Modal>
  );
}
