import { useMemo, useState, type ReactNode } from "react";
import { Badge, Button, Card, Icon, IconButton, Input, LevelMeter, LogStream, MetricPill, Slider, Switch, WaveScope } from "../ds";
import type { Backend } from "../api/backend";
import { generateGadgetScript } from "../graph/gadgetscript";
import { fmtUptime } from "./Header";
import { cfgNum, viewOf } from "./nodes";
import type { Studio } from "./store";

function Monitor({ label, tone, live, badge, height = 180, detail }: { label: string; tone: string; live: boolean; badge?: ReactNode; height?: number; detail: string }) {
  return (
    <div style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", gap: "var(--sp-3)" }}>
      <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-3)" }}>
        <span className="gc-caps" style={{ color: tone }}>{label}</span>
        <div style={{ flex: 1, height: 1, background: "var(--line-hairline)" }} />
        {badge}
      </div>
      <div style={{ position: "relative", height, borderRadius: "var(--r-md)", overflow: "hidden", border: "1px solid " + (tone === "var(--sig-live)" && live ? "var(--sig-live-dim)" : "var(--line-border)"), background: "var(--surface-input)", display: "flex", alignItems: "center", justifyContent: "center" }}>
        {live ? <div className="gc-pulse" style={{ position: "absolute", inset: 0, background: "color-mix(in srgb, " + tone + " 8%, transparent)" }} /> : null}
        <span style={{ font: "var(--type-meta)", color: "var(--text-faint)", textAlign: "center", zIndex: 1 }}>
          {live ? "PIPELINE RUNNING" : "NO SIGNAL"}<br /><span style={{ fontSize: "var(--fs-tick)" }}>{detail}</span>
        </span>
      </div>
    </div>
  );
}

/** Audio + vision mixer. Sources and strips come from the current graph; gain writes the node's config. */
export function MixerView({ studio }: { studio: Studio }) {
  const { s, act, catalog } = studio;
  const running = s.status?.running ?? null;
  const live = running != null && running === s.doc.id;
  const views = useMemo(() => s.doc.nodes.map(n => ({ n, v: viewOf(n, catalog, s.doc.edges) })), [s.doc, catalog]);
  const sources = views.filter(x => x.v.inputs.length === 0 && x.v.outputs.some(p => p.kind === "video"));
  const strips = views.filter(x => x.v.inputs.length === 0 && x.v.outputs.some(p => p.kind === "audio")).slice(0, 6);
  const [preview, setPreview] = useState<string | null>(null);
  const pv = sources.find(x => x.n.id === preview) ?? sources[0];
  const metered = !!s.levels[0];
  return (
    <div className="gc-scroll gc-stack" style={{ flex: 1, minWidth: 0, overflow: "auto", padding: "var(--sp-7)", display: "flex", flexDirection: "column", gap: "var(--sp-7)", background: "var(--surface-app)" }}>
      <div style={{ display: "flex", gap: "var(--sp-7)" }}>
        <Monitor label="Preview" tone="var(--sig-video)" live={false} detail="video preview is not streamed to the Studio in v0" badge={pv ? <Badge tone="video">{pv.v.title}</Badge> : <Badge tone="neutral">No video source</Badge>} />
        <div style={{ display: "flex", flexDirection: "column", justifyContent: "center", gap: "var(--sp-4)", paddingTop: "var(--sp-8)" }}>
          {live ? <Button variant="danger" icon="stop" onClick={act.stop}>Stop</Button> : <Button variant="primary" icon="switch_video" onClick={act.goOnAir} disabled={s.busy}>Take</Button>}
          <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)", maxWidth: 96, textAlign: "center" }}>Take runs this pipeline on the Pi</span>
        </div>
        <Monitor label="Program → USB" tone="var(--sig-live)" live={live} detail={live ? s.doc.name : "nothing routed to the host"}
          badge={<Badge tone="live" solid={live} dot={!live}>{live ? (s.status?.onAir ? "On air" : "Armed") : "Off air"}</Badge>} />
      </div>

      <div style={{ display: "flex", gap: "var(--sp-4)", flexWrap: "wrap" }}>
        {sources.map(x => (
          <button type="button" key={x.n.id} onClick={() => setPreview(x.n.id)} data-selected={pv?.n.id === x.n.id ? "true" : "false"}
            style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-4) var(--sp-5)", cursor: "pointer", background: pv?.n.id === x.n.id ? "var(--surface-selected)" : "var(--surface-card)", borderRadius: "var(--r-md)", border: "1px solid " + (pv?.n.id === x.n.id ? "var(--sig-video)" : "var(--line-border)") }}>
            <Icon name={x.v.icon} size="sm" style={{ color: "var(--sig-video)" }} />
            <span style={{ font: "var(--type-ui)", color: "var(--text-primary)" }}>{x.v.title}</span>
          </button>
        ))}
        {!sources.length ? <p style={{ font: "var(--type-body)", color: "var(--text-faint)" }}>This pipeline has no video sources. Add one on the patch canvas.</p> : null}
      </div>

      <Card eyebrow="Audio mixer" title="Source strips" icon="tune" padding="var(--sp-6)"
        actions={<Badge tone="audio" icon="graphic_eq">{metered ? "simulated levels" : "levels not reported in v0"}</Badge>}>
        <div style={{ display: "flex", gap: "var(--sp-9)", flexWrap: "wrap" }}>
          {strips.map((x, i) => {
            const gain = cfgNum(x.n, "gain", 0);
            return (
              <div key={x.n.id} style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: "var(--sp-4)", width: 92 }}>
                <Icon name={x.v.icon} size="md" style={{ color: "var(--sig-audio)" }} />
                <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-muted)", textAlign: "center" }}>{x.v.title}</span>
                <div style={{ display: "flex", gap: "var(--sp-4)", alignItems: "flex-end" }}>
                  <LevelMeter orientation="vertical" height={96} showScale={false} channels={[s.levels[0] * (1 - i * 0.12), (s.levels[1] ?? 0) * (1 - i * 0.15)]} />
                  <div style={{ height: 96, width: 24, display: "flex", alignItems: "center", justifyContent: "center" }}>
                    <Slider value={gain} min={-24} max={12} tone="var(--sig-audio)" aria-label={x.v.title + " gain"} onChange={g => act.patchConfig(x.n.id, { gain: g })} style={{ width: 96, transform: "rotate(-90deg)", flex: "0 0 auto" }} />
                  </div>
                </div>
                <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)" }}>{gain} dB</span>
              </div>
            );
          })}
          {!strips.length ? <p style={{ font: "var(--type-body)", color: "var(--text-faint)" }}>No audio sources in this pipeline.</p> : null}
          <div style={{ flex: 1, minWidth: 240, display: "flex", flexDirection: "column", gap: "var(--sp-5)" }}>
            <WaveScope label="Program bus · oscilloscope" height={80} live={metered} gain={metered ? 1 : 0} />
            <WaveScope label="Spectrum" mode="spectrum" height={64} live={metered} gain={metered ? 1 : 0} tone="var(--sig-audio)" />
            <LevelMeter label="UAC2 mic out (what the host hears)" channels={s.levels} peaks={s.peaks} width={320} />
          </div>
        </div>
      </Card>
    </div>
  );
}

const SUBSYSTEMS = ["all", "server", "session", "pipeline"] as const;

export function DiagnosticsView({ studio, backend }: { studio: Studio; backend: Backend }) {
  const { s, act, catalog } = studio;
  const [filter, setFilter] = useState<(typeof SUBSYSTEMS)[number]>("all");
  const [level, setLevel] = useState<"all" | "warn" | "err">("all");
  const [q, setQ] = useState("");
  const ql = q.trim().toLowerCase();
  const lines = s.logs.filter(l => (filter === "all" || l.src === filter) && (level === "all" || (level === "err" ? l.level === "err" : l.level === "err" || l.level === "warn")) && (!ql || l.msg.toLowerCase().includes(ql)));
  const script = useMemo(() => generateGadgetScript(s.doc, catalog), [s.doc, catalog]);
  const m = s.status?.metrics;
  const download = (name: string, text: string, type: string) => {
    const a = document.createElement("a");
    a.href = URL.createObjectURL(new Blob([text], { type })); a.download = name; a.click();
    setTimeout(() => URL.revokeObjectURL(a.href), 1000);
  };
  const chip = (id: string, active: boolean, onClick: () => void) => (
    <button type="button" key={id} className="gc-row" onClick={onClick}
      style={{ height: "var(--ctl-xs)", padding: "0 var(--sp-4)", borderRadius: "var(--r-xs)", cursor: "pointer", border: "1px solid " + (active ? "var(--brand)" : "var(--line-border)"), background: active ? "var(--surface-selected)" : "transparent", font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: active ? "var(--brand)" : "var(--text-faint)" }}>{id}</button>
  );
  const running = s.status?.running ?? null;
  return (
    <div style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", minHeight: 0, background: "var(--surface-app)" }}>
      <div style={{ height: "var(--h-subbar)", display: "flex", alignItems: "center", gap: "var(--sp-3)", padding: "0 var(--sp-5)", borderBottom: "1px solid var(--line-border)", flex: "0 0 auto" }}>
        {SUBSYSTEMS.map(x => chip(x, filter === x, () => setFilter(x)))}
        <span style={{ width: 1, height: 16, background: "var(--line-hairline)", margin: "0 var(--sp-2)" }} />
        {chip("warn+", level === "warn", () => setLevel(level === "warn" ? "all" : "warn"))}
        {chip("errors", level === "err", () => setLevel(level === "err" ? "all" : "err"))}
        <div style={{ width: 200, marginLeft: "var(--sp-3)" }}><Input size="sm" icon="search" placeholder="Search log" value={q} onChange={e => setQ(e.target.value)} aria-label="Search log" /></div>
        <div style={{ flex: 1 }} />
        <Switch size="sm" checked={s.follow} onChange={act.setFollow} label="Follow tail" />
        <IconButton icon="delete_sweep" label="Clear log view" size="sm" onClick={act.clearLogs} />
        <IconButton icon="download" label="Download log" size="sm" onClick={() => download("gadgetcast-log.txt", s.logs.map(l => l.t + "  " + l.src.toUpperCase().padEnd(9) + l.level.toUpperCase().padEnd(5) + " " + l.msg).join("\n"), "text/plain")} />
      </div>
      <div className="gc-scroll gc-stack" style={{ flex: 1, overflow: "auto", padding: "var(--sp-7)", display: "flex", flexDirection: "column", gap: "var(--sp-7)" }}>
        <div style={{ display: "flex", gap: "var(--sp-3)", flexWrap: "wrap" }}>
          <MetricPill icon="memory" label="cpu" value={m ? m.cpuPct.toFixed(1) : "—"} unit="%" tone={m && m.cpuPct > 80 ? "var(--st-warn)" : undefined} />
          <MetricPill icon="memory_alt" label="ram" value={m ? Math.round(m.ramUsedMB) + " / " + Math.round(m.ramTotalMB) : "—"} unit="MB" />
          <MetricPill icon="thermostat" label="soc" value={m?.tempC != null ? m.tempC.toFixed(1) : "—"} unit="°C" tone={m?.tempC != null && m.tempC > 70 ? "var(--st-warn)" : undefined} />
          <MetricPill icon="stacks" label="load" value={m ? m.loadAvg.map(x => x.toFixed(2)).join(" · ") : "—"} />
          <MetricPill icon="schedule" label="uptime" value={m ? fmtUptime(m.uptimeS) : "—"} />
          <MetricPill icon="groups" label="clients" value={s.status?.clients ?? 0} />
          <MetricPill icon="sensors" label="events" value={s.events} tone={s.events === "open" ? "var(--st-ok)" : "var(--st-warn)"} />
        </div>
        <Card eyebrow={backend.mode === "live" ? "Realtime · GET /api/events" : "Offline demo · simulated"} title="Server, session and pipeline events" icon="terminal" padding="var(--sp-5)"
          actions={<Badge tone={running ? "ok" : "neutral"} dot>{running ? "running · " + running : "idle"}</Badge>}>
          <LogStream lines={lines} height={240} follow={s.follow} empty={s.logs.length ? "No lines match the filter." : "Waiting for the Pi to log something."} />
        </Card>
        <Card eyebrow="GadgetScript" title="Generated from the graph" icon="code" padding="0"
          actions={<>
            <Button size="xs" variant="ghost" icon="content_copy" onClick={() => { navigator.clipboard?.writeText(script).then(() => act.toast("GadgetScript copied", "content_copy"), () => act.toast("Clipboard unavailable", "error", "err")); }}>Copy</Button>
            <Button size="xs" variant="ghost" icon="download" onClick={() => download((s.doc.id ?? "pipeline") + ".sh", script, "text/x-shellscript")}>Download</Button>
          </>}
          footer="A readable sketch of the configfs gadget and gst-launch chains · the Pi builds the real pipeline">
          <pre className="gc-scroll" data-testid="gadgetscript" style={{ margin: 0, padding: "var(--sp-6)", maxHeight: 280, overflow: "auto", background: "var(--surface-canvas)", font: "var(--type-meta)", color: "var(--text-muted)", whiteSpace: "pre" }}>{script}</pre>
        </Card>
      </div>
    </div>
  );
}
