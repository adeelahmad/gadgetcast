import { useRef, useState, type ReactNode } from "react";
import { Badge, Button, DeviceRow, Icon, IconButton, Input, PresetRow, Tooltip } from "../ds";
import type { Backend } from "../api/backend";
import { sigTint, sigVar } from "./nodes";
import type { RailMode, Studio } from "./store";

const RAILS: [string, RailMode, string][] = [["widgets", "nodes", "Node palette"], ["bookmarks", "library", "Pipeline library"], ["devices_other", "devices", "Hardware & clients"], ["help", "help", "Shortcuts"]];

export const SHORTCUTS: [string, string][] = [
  ["⌘K", "Command palette"], ["⌘S", "Save pipeline"], ["⌘↵", "Go on air / stop"], ["R", "Revert to saved"],
  ["1 · 2 · 3 · 4", "Canvas · Mixer · Diagnostics · Hardware"], ["⇧C", "Quick capture"], ["Del", "Delete selection"],
  ["⌘D", "Duplicate selection"], ["⌘A", "Select all nodes"], ["V · H", "Select · hand tool"], ["Space", "Hold to pan"], ["⇧1", "Zoom to fit"],
];

export function Rail({ studio }: { studio: Studio }) {
  const { s, act } = studio;
  return (
    <nav style={{ width: "var(--w-rail)", flex: "0 0 auto", display: "flex", flexDirection: "column", alignItems: "center", gap: "var(--sp-3)", padding: "var(--sp-4) 0", background: "var(--surface-app)", borderRight: "1px solid var(--line-border)" }}>
      {RAILS.map(([icon, id, label]) => (
        <Tooltip key={id} label={label} side="right"><IconButton icon={icon} label={label} size="md" active={s.rail === id} onClick={() => act.setRail(id)} /></Tooltip>
      ))}
      <div style={{ flex: 1 }} />
      <Tooltip label="Patch canvas" side="right"><IconButton icon="account_tree" label="Patch canvas" size="md" onClick={() => act.setView("canvas")} /></Tooltip>
    </nav>
  );
}

export function Sidebar({ studio, backend, floating = false }: { studio: Studio; backend: Backend; floating?: boolean }) {
  const { s, act } = studio;
  const [collapsed, setCollapsed] = useState(false);
  const [q, setQ] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);
  const ql = q.trim().toLowerCase();

  const head = (title: string, right?: ReactNode) => (
    <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", height: "var(--h-subbar)", flex: "0 0 auto", padding: "0 var(--sp-5)", borderBottom: "1px solid var(--line-hairline)" }}>
      <span className="gc-caps" style={{ flex: 1 }}>{title}</span>{right}
    </div>
  );

  let body: ReactNode;
  if (s.rail === "nodes") {
    body = (
      <>
        {head("Node palette", <Badge tone="neutral">{s.groups.reduce((n, g) => n + g.items.length, 0)}</Badge>)}
        <div style={{ padding: "var(--sp-4) var(--sp-5)" }}><Input icon="search" size="sm" placeholder="Filter nodes" value={q} onChange={e => setQ(e.target.value)} aria-label="Filter nodes" /></div>
        <div className="gc-scroll" style={{ flex: 1, overflow: "auto", padding: "0 var(--sp-4) var(--sp-6)" }}>
          {s.groups.map(g => {
            const items = g.items.filter(i => !ql || (i.title + " " + (i.subtitle ?? "") + " " + g.group).toLowerCase().includes(ql));
            if (!items.length) return null;
            return (
              <div key={g.group} style={{ marginTop: "var(--sp-5)" }}>
                <div className="gc-caps" style={{ padding: "0 var(--sp-3) var(--sp-3)" }}>{g.group}</div>
                <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-1)" }}>
                  {items.map(i => (
                    <div key={i.type} className="gc-row" draggable data-node-type={i.type}
                      onDragStart={e => { e.dataTransfer.setData("text/gc-node", i.type); e.dataTransfer.effectAllowed = "copy"; }}
                      onDoubleClick={() => act.addNode(i.type)}
                      title="Drag onto the canvas, or double-click to drop at centre"
                      style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "6px 8px", borderRadius: "var(--r-sm)", cursor: "grab" }}>
                      <span style={{ width: 28, height: 28, flex: "0 0 auto", borderRadius: "var(--r-sm)", display: "grid", placeItems: "center", background: sigTint(i.cls) }}>
                        <Icon name={i.icon} size="sm" style={{ color: sigVar(i.cls) }} />
                      </span>
                      <span style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column" }}>
                        <span style={{ font: "var(--type-ui)", color: "var(--text-primary)" }}>{i.title}</span>
                        <span style={{ font: "var(--fw-regular) var(--fs-meta)/1.2 var(--font-sans)", color: "var(--text-faint)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{i.subtitle || i.url || i.dev}</span>
                      </span>
                      <Icon name="drag_indicator" size="xs" style={{ color: "var(--text-disabled)" }} />
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
          {!s.groups.length ? <p style={{ padding: "var(--sp-6) var(--sp-3)", font: "var(--type-body)", color: "var(--text-faint)" }}>Loading the node catalogue from the Pi…</p> : null}
        </div>
      </>
    );
  } else if (s.rail === "library") {
    body = (
      <>
        {head("Pipeline library", <IconButton icon="add" label="New pipeline" size="sm" onClick={() => act.openPipeline(null)} />)}
        <div className="gc-scroll gc-stack" style={{ flex: 1, overflow: "auto", padding: "var(--sp-4)", display: "flex", flexDirection: "column", gap: "var(--sp-1)" }}>
          {s.pipelines.map(p => (
            <PresetRow key={p.id} name={p.name} meta={p.nodes.length + " nodes · " + p.edges.length + " cables" + (p.isDefault ? " · boot default" : "")}
              selected={s.doc.id === p.id} running={s.status?.running === p.id} isDefault={!!p.isDefault}
              onClick={() => act.openPipeline(p.id)}
              actions={<span style={{ display: "flex", gap: 2 }} onClick={e => e.stopPropagation()}>
                <IconButton icon="push_pin" label="Load on boot" size="xs" active={!!p.isDefault} onClick={() => act.setDefault(p.id)} />
                <IconButton icon="delete" label={"Delete " + p.name} size="xs" onClick={() => act.askDelete(p.id)} />
              </span>} />
          ))}
          {!s.pipelines.length ? <p style={{ padding: "var(--sp-5) var(--sp-3)", font: "var(--type-body)", color: "var(--text-faint)" }}>No pipelines on the Pi yet. Build one on the canvas and save it.</p> : null}
          <div style={{ display: "flex", gap: "var(--sp-3)", padding: "var(--sp-5) var(--sp-3) 0" }}>
            <Button size="sm" variant="secondary" icon="save" onClick={() => act.setModal("save")} fullWidth>Save as…</Button>
            <Button size="sm" variant="ghost" icon="download" onClick={act.exportJson}>JSON</Button>
          </div>
          <button type="button" className="gc-row" onClick={() => fileRef.current?.click()}
            style={{ display: "flex", alignItems: "center", gap: "var(--sp-3)", padding: "var(--sp-4) var(--sp-3)", cursor: "pointer", background: "transparent", border: 0, borderRadius: "var(--r-sm)", textAlign: "left" }}>
            <Icon name="upload" size="sm" style={{ color: "var(--text-faint)" }} />
            <span style={{ font: "var(--type-ui)", color: "var(--text-muted)" }}>Import pipeline JSON</span>
          </button>
          <input ref={fileRef} type="file" accept="application/json,.json" style={{ display: "none" }}
            onChange={e => { const f = e.target.files?.[0]; if (f) act.importJson(f); e.target.value = ""; }} />
        </div>
      </>
    );
  } else if (s.rail === "devices") {
    const d = s.devices;
    body = (
      <>
        {head("Hardware & clients", <IconButton icon="refresh" label="Rescan" size="sm" onClick={act.refreshDevices} />)}
        <div className="gc-scroll" style={{ flex: 1, overflow: "auto" }}>
          <div className="gc-caps" style={{ padding: "var(--sp-5) var(--sp-5) var(--sp-3)" }}>Capture devices</div>
          {d?.devices.map(x => <DeviceRow key={x.id} name={x.name} detail={x.path} icon={x.kind === "video" ? "videocam" : "mic"} transport={x.kind} status="online" />)}
          {d && !d.devices.length ? <p style={{ padding: "0 var(--sp-5)", font: "var(--type-body)", color: "var(--text-faint)" }}>No V4L2 or ALSA devices reported.</p> : null}
          <div className="gc-caps" style={{ padding: "var(--sp-6) var(--sp-5) var(--sp-3)" }}>Browser clients</div>
          {d?.clients.map(c => <DeviceRow key={c.id} name={c.id} detail={c.onAir ? "routing media to the USB host" : "paired · idle"} icon="smartphone" transport="LAN" status={c.onAir ? "online" : "pairing"} statusLabel={c.onAir ? "On air" : "Idle"} />)}
          {d && !d.clients.length ? <p style={{ padding: "0 var(--sp-5)", font: "var(--type-body)", color: "var(--text-faint)" }}>No browser clients connected.</p> : null}
          {!d ? <p style={{ padding: "var(--sp-5)", font: "var(--type-body)", color: "var(--text-faint)" }}>Device list unavailable.</p> : null}
          <div style={{ padding: "var(--sp-6) var(--sp-5)" }}>
            <Button size="sm" variant="secondary" icon="memory" onClick={() => act.setView("hardware")} fullWidth>Open hardware view</Button>
          </div>
        </div>
      </>
    );
  } else {
    body = (
      <>
        {head("Shortcuts")}
        <div className="gc-scroll gc-stack" style={{ flex: 1, overflow: "auto", padding: "var(--sp-5)", display: "flex", flexDirection: "column", gap: "var(--sp-4)" }}>
          {SHORTCUTS.map(([k, label]) => (
            <div key={k} style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)" }}>
              <kbd style={{ minWidth: 64, textAlign: "center" }}>{k}</kbd>
              <span style={{ font: "var(--type-ui)", color: "var(--text-muted)" }}>{label}</span>
            </div>
          ))}
          <p style={{ font: "var(--type-body)", color: "var(--text-faint)", paddingTop: "var(--sp-5)" }}>
            {backend.mode === "live"
              ? "Connected to the Pi at " + window.location.host + ". Every browser client pairs with a PIN before it can route media to the USB host."
              : "Offline demo: the Pi did not answer, so this session runs on bundled data. Nothing you do here reaches a device."}
          </p>
        </div>
      </>
    );
  }

  if (!floating) {
    return <aside style={{ width: "var(--w-sidebar)", flex: "0 0 auto", display: "flex", flexDirection: "column", minHeight: 0, background: "var(--surface-panel)", borderRight: "1px solid var(--line-border)" }}>{body}</aside>;
  }
  if (collapsed) {
    return (
      <div data-ui style={{ position: "absolute", left: 16, top: 16, zIndex: 30, display: "flex", flexDirection: "column", gap: 2, padding: 4, background: "var(--surface-panel)", border: "1px solid var(--line-border)", borderRadius: "var(--r-lg)" }}>
        {RAILS.map(([icon, id, label]) => <Tooltip key={id} label={label} side="right"><IconButton icon={icon} label={label} size="sm" active={s.rail === id} onClick={() => { act.setRail(id); setCollapsed(false); }} /></Tooltip>)}
      </div>
    );
  }
  return (
    <aside data-ui data-testid="sidebar" style={{ position: "absolute", left: 16, top: 16, bottom: 16, width: "var(--w-sidebar)", zIndex: 30, display: "flex", flexDirection: "column", minHeight: 0, overflow: "hidden", background: "var(--surface-panel)", border: "1px solid var(--line-border)", borderRadius: "var(--r-lg)" }}>
      <div style={{ display: "flex", alignItems: "center", gap: 2, padding: 6, borderBottom: "1px solid var(--line-hairline)" }}>
        {RAILS.map(([icon, id, label]) => <Tooltip key={id} label={label} side="bottom"><IconButton icon={icon} label={label} size="sm" active={s.rail === id} onClick={() => act.setRail(id)} /></Tooltip>)}
        <span style={{ flex: 1 }} />
        <Tooltip label="Collapse panel" side="bottom"><IconButton icon="left_panel_close" label="Collapse" size="sm" onClick={() => setCollapsed(true)} /></Tooltip>
      </div>
      {body}
    </aside>
  );
}
