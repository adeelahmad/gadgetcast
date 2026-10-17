import { Button, Card, Icon, IconButton, Input, MetricPill, PortHandle, Select, Slider, Switch, WaveScope } from "../ds";
import { cfgNum, cfgStr, sigTint, sigVar, URL_RULE, URL_TYPES, viewOf } from "./nodes";
import type { Studio } from "./store";

export const VIDEO_FORMATS = ["MJPEG 1080p30", "MJPEG 720p30", "YUYV 480p30", "H.264 1080p30 (UVC 1.5)"];
export const AI_MODELS = ["gemini-live-2.5-flash", "gemini-2.5-flash", "gemini-2.5-pro"];

/** Right-hand inspector: config for the selected node. Every field writes the node's `config` and is saved with the pipeline. */
export function Inspector({ studio, floating = false }: { studio: Studio; floating?: boolean }) {
  const { s, act, catalog } = studio;
  const n = s.doc.nodes.find(x => x.id === s.selected);
  if (!n && floating) return null;
  if (!n) {
    return (
      <aside style={{ width: "var(--w-inspector)", flex: "0 0 auto", display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: "var(--sp-5)", padding: "var(--sp-9)", background: "var(--surface-panel)", borderLeft: "1px solid var(--line-border)", textAlign: "center" }}>
        <Icon name="touch_app" size="xl" style={{ color: "var(--text-faint)" }} />
        <p style={{ font: "var(--type-body)", color: "var(--text-faint)" }}>Select a node to configure it. Drag from an output port to patch a cable; right-click a node for quick actions.</p>
      </aside>
    );
  }
  const v = viewOf(n, catalog, s.doc.edges);
  const url = cfgStr(n, "url");
  const urlish = URL_TYPES.has(n.type);
  const bad = urlish && url.trim().length > 0 && !URL_RULE.test(url.trim());
  const set = (patch: Record<string, unknown>) => act.patchConfig(n.id, patch);
  const m = s.status?.metrics;
  const live = s.status?.running != null && s.status.running === s.doc.id;
  return (
    <aside className="gc-scroll" tabIndex={0} aria-label="Inspector" data-ui data-testid="inspector" style={floating
      ? { position: "absolute", right: 16, top: 64, bottom: 16, width: "var(--w-inspector)", zIndex: 30, overflow: "auto", background: "var(--surface-panel)", border: "1px solid var(--line-border)", borderRadius: "var(--r-lg)" }
      : { width: "var(--w-inspector)", flex: "0 0 auto", overflow: "auto", background: "var(--surface-panel)", borderLeft: "1px solid var(--line-border)" }}>
      <div style={{ display: "flex", alignItems: "center", gap: 10, height: 44, padding: "0 8px 0 12px", borderBottom: "1px solid var(--line-hairline)", position: "sticky", top: 0, background: "var(--surface-panel)", zIndex: 2 }}>
        <span style={{ width: 26, height: 26, borderRadius: "var(--r-md)", display: "grid", placeItems: "center", background: sigTint(v.cls) }}><Icon name={v.icon} size="sm" style={{ color: sigVar(v.cls) }} /></span>
        <span style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column" }}>
          <span style={{ font: "var(--fw-semibold) var(--fs-body)/1.2 var(--font-sans)", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{v.title}</span>
          <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>{v.group} · {n.id}</span>
        </span>
        <IconButton icon="open_in_full" label="Open node config" size="sm" onClick={() => act.setModal("config")} />
        <IconButton icon="delete" label="Delete node" size="sm" tone="var(--st-err)" onClick={() => act.removeNodes([n.id])} />
        {floating ? <IconButton icon="close" label="Close inspector" size="sm" onClick={() => act.select(null)} /> : null}
      </div>
      <div style={{ padding: "var(--sp-6) var(--sp-5)", display: "flex", flexDirection: "column", gap: "var(--sp-6)" }}>
        <Input label="Node name" value={n.title ?? v.title} onChange={e => act.patchNode(n.id, { title: e.target.value })} />
        {urlish ? (
          <Input label="Source path / URL" mono icon="link" value={url} invalid={bad}
            hint={bad ? "Unsupported scheme · use rtsp://, rtmp://, udp://, srt://, https:// or an absolute path" : "Validated by the Pi when the pipeline runs"}
            onChange={e => set({ url: e.target.value })} />
        ) : null}
        {v.cls === "video" || n.type === "uvc_out" ? <Select label="Format" value={cfgStr(n, "format", VIDEO_FORMATS[0])} options={VIDEO_FORMATS} onChange={e => set({ format: e.target.value })} /> : null}
        {v.cls === "audio" || n.type === "uac_out" || n.type === "browser_mic" ? (
          <>
            <Slider label="Gain" value={cfgNum(n, "gain", 0)} min={-24} max={12} unit="dB" tone="var(--sig-audio)" onChange={g => set({ gain: g })} />
            <WaveScope label="Signal" height={52} live={live && !!s.levels[0]} gain={live && s.levels[0] ? 1 : 0} />
          </>
        ) : null}
        {v.cls === "storage" ? (
          <>
            <Slider label="Upload interval" value={cfgNum(n, "interval", 30)} min={5} max={300} step={5} unit="s" onChange={x => set({ interval: x })} />
            <Select label="Payload" value={cfgStr(n, "payload", "Still frame (JPEG)")} options={["Still frame (JPEG)", "Segment (10 s MP4)", "Audio only (WAV)", "Transcript (JSON)"]} onChange={e => set({ payload: e.target.value })} />
          </>
        ) : null}
        {v.cls === "ai" ? (
          <>
            <Select label="Model" value={cfgStr(n, "model", AI_MODELS[0])} options={AI_MODELS} onChange={e => set({ model: e.target.value })} />
            <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-3)" }}>
              <span className="gc-caps">System instructions</span>
              <div style={{ font: "var(--type-meta)", color: "var(--text-muted)", background: "var(--surface-input)", border: "1px solid var(--line-strong)", borderRadius: "var(--r-sm)", padding: "var(--sp-4)", maxHeight: 88, overflow: "hidden" }}>{cfgStr(n, "prompt") || "No instructions yet."}</div>
              <Button size="sm" variant="secondary" icon="edit_note" onClick={() => act.setModal("config")}>Edit prompt template</Button>
            </div>
            <Slider label="Frame sample rate" value={cfgNum(n, "fps", 1)} min={0.2} max={5} step={0.2} unit="fps" tone="var(--sig-ai)" onChange={x => set({ fps: Math.round(x * 10) / 10 })} />
          </>
        ) : null}
        {v.cls === "bluetooth" ? (
          <>
            <Select label="Bluetooth profile" value={cfgStr(n, "btProfile", "a2dp_source")} options={["a2dp_source", "a2dp_sink", "hfp_handsfree", "ble_audio", "bt_camera_remote"]} onChange={e => set({ btProfile: e.target.value })} />
            <Select label="Codec" value={cfgStr(n, "btCodec", "sbc")} options={["sbc", "aac", "ldac", "aptx_hd", "lc3"]} onChange={e => set({ btCodec: e.target.value })} />
            <Input label="Paired device MAC" mono icon="bluetooth" value={cfgStr(n, "btMac")} placeholder="AC:80:0A:41:9B:2E" onChange={e => set({ btMac: e.target.value })} />
          </>
        ) : null}
        {n.type === "cond" ? <Input label="Condition" mono value={cfgStr(n, "condition", "drift_ppm > 40")} hint="Evaluated per data tick" onChange={e => set({ condition: e.target.value })} /> : null}
        {n.type === "setvar" ? (
          <>
            <Input label="Variable" mono value={cfgStr(n, "varKey")} placeholder="on_air_source" onChange={e => set({ varKey: e.target.value })} />
            <Input label="Value" mono value={cfgStr(n, "varValue")} onChange={e => set({ varValue: e.target.value })} />
          </>
        ) : null}
        {n.type === "exec" ? <Input label="Script path" mono icon="terminal" value={cfgStr(n, "script", "/opt/gadgetcast/hooks/take.py")} hint="Runs on the Pi as the gadgetcast user" onChange={e => set({ script: e.target.value })} /> : null}
        <Input label="GStreamer element" mono icon="account_tree" value={v.gst || "—"} hint={v.category + " · " + v.media} readOnly />

        <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-4)" }}>
          <span className="gc-caps">Ports</span>
          {v.inputs.map((p, i) => <PortHandle key={"i" + i} kind={p.kind} direction="in" label={p.label + " in · " + p.kind} connected={s.doc.edges.some(e => e.to.node === n.id && e.to.port === i)} />)}
          {v.outputs.map((p, i) => <PortHandle key={"o" + i} kind={p.kind} direction="in" label={p.label + " out · " + p.kind} connected={s.doc.edges.some(e => e.from.node === n.id && e.from.port === i)} />)}
          {!v.inputs.length && !v.outputs.length ? <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>No ports</span> : null}
        </div>

        <Switch checked={n.enabled} onChange={() => act.toggleNode(n.id)} label="Node enabled" description="Bypassed nodes keep their cables but pass nothing" />

        <Card eyebrow="On the Pi" title="System load" icon="speed" padding="var(--sp-5)" footer="Per-node figures are not reported by the v0 API">
          <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--sp-3)" }}>
            <MetricPill icon="memory" label="cpu" value={m ? m.cpuPct.toFixed(0) : "—"} unit="%" />
            <MetricPill icon="thermostat" label="soc" value={m?.tempC != null ? m.tempC.toFixed(0) : "—"} unit="°C" />
            <MetricPill icon="stacks" label="load" value={m ? m.loadAvg[0]?.toFixed(2) ?? "—" : "—"} />
          </div>
        </Card>
        <Button size="sm" variant="ghost" icon="link_off" onClick={() => act.clearLinks(n.id)}>Clear connections</Button>
      </div>
    </aside>
  );
}
