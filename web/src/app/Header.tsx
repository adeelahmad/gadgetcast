import logoDark from "../ds/assets/logo-dark.svg";
import logoLight from "../ds/assets/logo.svg";
import { Badge, Button, Icon, IconButton, MetricPill, OnAirIndicator, Tooltip } from "../ds";
import type { Backend } from "../api/backend";
import type { Studio } from "./store";

export function fmtUptime(s: number): string {
  const d = Math.floor(s / 86400), h = Math.floor((s % 86400) / 3600), m = Math.floor((s % 3600) / 60);
  return d ? d + "d " + h + "h" : h ? h + "h " + m + "m" : m + "m";
}

export function Header({ studio, backend, reason }: { studio: Studio; backend: Backend; reason?: string }) {
  const { s, act } = studio;
  const running = s.status?.running ?? null;
  const live = running != null && running === s.doc.id;
  const runningName = running ? s.pipelines.find(p => p.id === running)?.name ?? running : undefined;
  const m = s.status?.metrics;
  const tally = running ? (s.status?.onAir ? "live" : "armed") : "off";
  return (
    <header style={{ height: "var(--h-header)", flex: "0 0 auto", display: "flex", alignItems: "center", gap: "var(--sp-5)", padding: "0 var(--sp-6)", background: "var(--surface-panel)", borderBottom: "1px solid var(--line-border)", zIndex: 20, minWidth: 0 }}>
      <img src={s.theme === "dark" ? logoDark : logoLight} alt="GadgetCast" style={{ height: 19, flex: "0 0 auto" }} />

      {backend.mode === "demo" ? (
        <Tooltip label="Offline demo" detail={(reason ?? "The Pi did not answer") + " · bundled data, nothing reaches a device"} side="bottom" tone="warn">
          <button type="button" onClick={() => act.setModal("settings")} data-testid="demo-badge" style={{ background: "none", border: 0, padding: 0, cursor: "pointer", display: "inline-flex" }}>
            <Badge tone="warn" icon="cloud_off">Offline demo</Badge>
          </button>
        </Tooltip>
      ) : (
        <Badge tone={s.events === "open" ? "ok" : "warn"} dot>{s.events === "open" ? "Pi v" + backend.version : "Reconnecting"}</Badge>
      )}

      <button type="button" onClick={() => act.setPresetMenu(!s.presetMenu)} data-testid="pipeline-chip"
        style={{ position: "relative", display: "flex", flex: "0 1 auto", minWidth: 0, whiteSpace: "nowrap", alignItems: "center", gap: "var(--sp-3)", height: "var(--ctl-md)", padding: "0 var(--sp-4)", background: "var(--surface-raised)", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)", cursor: "pointer", borderLeft: "2px solid " + (live ? "var(--sig-live)" : "var(--brand)") }}>
        <Icon name="account_tree" size="sm" style={{ color: live ? "var(--sig-live)" : "var(--text-faint)" }} />
        <span style={{ font: "var(--type-ui)", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis" }}>{s.doc.name}</span>
        {s.dirty ? <span title="Unsaved changes" style={{ width: 5, height: 5, borderRadius: "var(--r-pill)", background: "var(--st-warn)", flex: "0 0 auto" }} /> : null}
        <Icon name="expand_more" size="xs" style={{ color: "var(--text-faint)" }} />
      </button>

      <button type="button" onClick={() => act.setPalette(true)} data-testid="search"
        style={{ flex: "1 1 200px", minWidth: 0, maxWidth: 380, display: "flex", alignItems: "center", gap: "var(--sp-4)", height: "var(--ctl-md)", padding: "0 var(--sp-4)", background: "var(--surface-input)", border: "1px solid var(--line-strong)", borderRadius: "var(--r-sm)", cursor: "text" }}>
        <Icon name="search" size="sm" style={{ color: "var(--text-faint)" }} />
        <span style={{ flex: 1, minWidth: 0, textAlign: "left", font: "var(--type-ui)", color: "var(--text-disabled)", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>Search nodes, devices, pipelines…</span>
        <kbd>⌘K</kbd>
      </button>

      <div style={{ flex: 1 }} />

      <div style={{ display: "flex", flex: "0 0 auto", gap: "var(--sp-2)", padding: "0 var(--sp-3)", borderRight: "1px solid var(--line-hairline)" }}>
        <Tooltip label="Send my mic to the USB host" side="bottom"><IconButton icon="mic" label="Send my mic" size="sm" onClick={() => act.setModal("sendmic")} data-testid="send-mic" /></Tooltip>
        <Tooltip label="Capture screenshot" shortcut="⇧C" side="bottom"><IconButton icon="photo_camera" label="Screenshot" size="sm" onClick={() => act.openCapture("screenshot")} /></Tooltip>
        <Tooltip label="Record a clip" side="bottom"><IconButton icon="videocam" label="Record clip" size="sm" onClick={() => act.openCapture("video")} /></Tooltip>
        <Tooltip label="Capture audio clip" side="bottom"><IconButton icon="graphic_eq" label="Audio clip" size="sm" onClick={() => act.openCapture("audio")} /></Tooltip>
      </div>

      <MetricPill icon="memory" label="cpu" value={m ? m.cpuPct.toFixed(0) : "—"} unit="%" tone={m && m.cpuPct > 80 ? "var(--st-warn)" : undefined} />
      <MetricPill icon="thermostat" label="soc" value={m && m.tempC != null ? m.tempC.toFixed(0) : "—"} unit="°C" tone={m && m.tempC != null && m.tempC > 70 ? "var(--st-warn)" : undefined} />
      <OnAirIndicator state={tally} source={runningName} />
      <Button variant={running ? "danger" : "live"} size="sm" icon={running ? "stop" : "fiber_manual_record"} onClick={running ? act.stop : act.goOnAir} disabled={s.busy} data-testid="go-on-air">
        {running ? "Stop" : "Go on air"}
      </Button>
      <div style={{ display: "flex", gap: "var(--sp-2)", flex: "0 0 auto" }}>
        <IconButton icon={s.theme === "dark" ? "light_mode" : "dark_mode"} label="Toggle theme" size="sm" onClick={act.toggleTheme} data-testid="theme-toggle" />
        <IconButton icon="keyboard" label="Shortcuts" size="sm" onClick={() => act.setModal("keys")} />
        <IconButton icon="settings" label="Settings" size="sm" onClick={() => act.setModal("settings")} />
      </div>
    </header>
  );
}

export function StatusBar({ studio, backend }: { studio: Studio; backend: Backend }) {
  const { s } = studio;
  const m = s.status?.metrics;
  const item = (label: string, value: string, tone?: string) => (
    <span style={{ display: "flex", gap: "var(--sp-3)", alignItems: "center", font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)", whiteSpace: "nowrap" }}>
      <span style={{ textTransform: "uppercase", letterSpacing: "var(--ls-caps)" }}>{label}</span>
      <span style={{ color: tone || "var(--text-body)" }}>{value}</span>
    </span>
  );
  const running = s.status?.running ?? null;
  return (
    <footer style={{ height: "var(--h-statusbar)", flex: "0 0 auto", display: "flex", alignItems: "center", gap: "var(--sp-7)", padding: "0 var(--sp-6)", background: "var(--surface-panel)", borderTop: "1px solid var(--line-border)", overflow: "hidden" }}>
      <span style={{ display: "flex", alignItems: "center", gap: "var(--sp-3)", font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: backend.mode === "live" ? "var(--st-ok)" : "var(--st-warn)", whiteSpace: "nowrap" }}>
        <Icon name={backend.mode === "live" ? "lan" : "cloud_off"} size="xs" />
        {backend.mode === "live" ? window.location.host + " · api v" + backend.version : "offline demo · bundled data"}
      </span>
      {item("events", s.events, s.events === "open" ? "var(--st-ok)" : "var(--st-warn)")}
      {item("running", running ? running : "none", running ? "var(--sig-live)" : undefined)}
      {item("on air", s.status?.onAir ?? "—")}
      {item("clients", String(s.status?.clients ?? 0))}
      {m ? item("ram", Math.round(m.ramUsedMB) + " / " + Math.round(m.ramTotalMB) + " MB") : null}
      {m ? item("up", fmtUptime(m.uptimeS)) : null}
      <div style={{ flex: 1 }} />
      {item("autosave", s.autosave ? "on · 30 s" : "off", s.autosave ? "var(--st-ok)" : "var(--text-disabled)")}
      {item("graph", s.doc.nodes.length + " nodes · " + s.doc.edges.length + " cables")}
    </footer>
  );
}
