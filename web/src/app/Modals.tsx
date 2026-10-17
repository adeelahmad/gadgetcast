import { useEffect, useMemo, useRef, useState } from "react";
import { Badge, Button, Icon, Input, Modal, Select, Slider, Switch } from "../ds";
import type { Backend } from "../api/backend";
import { AI_MODELS, VIDEO_FORMATS } from "./Inspector";
import { cfgNum, cfgStr, viewOf } from "./nodes";
import { QuickCapture } from "./QuickCapture";
import { SendMic } from "./SendMic";
import { SHORTCUTS } from "./Sidebar";
import type { Studio } from "./store";

interface Row { group: string; label: string; icon: string; meta?: string; fn: () => void }

/** ⌘K command palette: actions, pipelines on the Pi, the node catalogue and hardware. Arrow keys + Enter. */
export function Palette({ studio }: { studio: Studio }) {
  const { s, act } = studio;
  const [q, setQ] = useState("");
  const [idx, setIdx] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);
  const running = s.status?.running ?? null;
  const rows = useMemo(() => {
    const r: Row[] = [];
    ([
      [running ? "Stop pipeline" : "Go on air", running ? "stop_circle" : "fiber_manual_record", running ? act.stop : act.goOnAir],
      ["Save pipeline", "save", () => (s.doc.id ? act.save() : act.setModal("save"))],
      ["Save as…", "save_as", () => act.setModal("save")],
      ["New pipeline", "note_add", () => act.openPipeline(null)],
      ["Revert to saved", "restart_alt", act.revert],
      ["Export pipeline JSON", "download", act.exportJson],
      ["Toggle theme", "contrast", act.toggleTheme],
      ["Quick capture", "photo_camera", () => act.openCapture("screenshot")],
      ["Open patch canvas", "account_tree", () => act.setView("canvas")],
      ["Open mixer", "tune", () => act.setView("mixer")],
      ["Open diagnostics", "monitor_heart", () => act.setView("diag")],
      ["Open hardware", "memory", () => act.setView("hardware")],
      ["Settings", "settings", () => act.setModal("settings")],
    ] as [string, string, () => void][]).forEach(([label, icon, fn]) => r.push({ group: "Action", label, icon, fn }));
    s.pipelines.forEach(p => r.push({ group: "Pipeline", label: p.name, icon: "account_tree", meta: p.nodes.length + " nodes", fn: () => act.openPipeline(p.id) }));
    s.groups.forEach(g => g.items.forEach(i => r.push({ group: "Node", label: i.title, icon: i.icon, meta: g.group, fn: () => { act.setView("canvas"); act.addNode(i.type); } })));
    s.devices?.devices.forEach(d => r.push({ group: "Hardware", label: d.name, icon: d.kind === "video" ? "videocam" : "mic", meta: d.path, fn: () => act.setView("hardware") }));
    return r;
  }, [s.pipelines, s.groups, s.devices, s.doc.id, running, act]);
  const ql = q.trim().toLowerCase();
  const hits = rows.filter(r => !ql || (r.label + " " + (r.meta ?? "") + " " + r.group).toLowerCase().includes(ql)).slice(0, 12);
  useEffect(() => setIdx(0), [q]);
  useEffect(() => { listRef.current?.querySelector("[data-active='true']")?.scrollIntoView({ block: "nearest" }); }, [idx]);
  const run = (r: Row | undefined) => { if (!r) return; act.setPalette(false); r.fn(); };
  return (
    <div onClick={e => { if (e.target === e.currentTarget) act.setPalette(false); }} data-testid="palette"
      style={{ position: "absolute", inset: 0, background: "var(--scrim)", zIndex: 90, display: "flex", justifyContent: "center", paddingTop: 96 }}>
      <div role="dialog" aria-modal="true" aria-label="Command palette" style={{ width: 560, maxWidth: "calc(100% - 32px)", maxHeight: 460, display: "flex", flexDirection: "column", background: "var(--surface-panel)", border: "1px solid var(--line-strong)", borderRadius: "var(--r-lg)", boxShadow: "var(--el-4)", overflow: "hidden", alignSelf: "flex-start" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-5) var(--sp-6)", borderBottom: "1px solid var(--line-hairline)" }}>
          <Icon name="search" size="md" style={{ color: "var(--brand)" }} />
          <input autoFocus value={q} onChange={e => setQ(e.target.value)} placeholder="Search nodes, hardware, pipelines, actions…" aria-label="Command"
            onKeyDown={e => {
              if (e.key === "ArrowDown") { e.preventDefault(); setIdx(i => Math.min(hits.length - 1, i + 1)); }
              else if (e.key === "ArrowUp") { e.preventDefault(); setIdx(i => Math.max(0, i - 1)); }
              else if (e.key === "Enter") { e.preventDefault(); run(hits[idx]); }
            }}
            style={{ flex: 1, background: "transparent", border: 0, outline: "none", font: "var(--type-body)", color: "var(--text-primary)", boxShadow: "none" }} />
          <kbd>esc</kbd>
        </div>
        <div ref={listRef} className="gc-scroll" style={{ overflow: "auto", padding: "var(--sp-3)" }}>
          {hits.map((r, i) => (
            <button type="button" key={r.group + r.label + i} className="gc-row" data-selected={i === idx ? "true" : "false"} data-active={i === idx ? "true" : "false"}
              onMouseEnter={() => setIdx(i)} onClick={() => run(r)}
              style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", width: "100%", height: "var(--ctl-lg)", padding: "0 var(--sp-5)", background: "transparent", border: 0, borderRadius: "var(--r-sm)", cursor: "pointer", textAlign: "left" }}>
              <Icon name={r.icon} size="sm" style={{ color: "var(--text-faint)" }} />
              <span style={{ font: "var(--type-ui)", color: "var(--text-primary)", flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{r.label}</span>
              {r.meta ? <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)", whiteSpace: "nowrap" }}>{r.meta}</span> : null}
              <Badge tone="neutral">{r.group}</Badge>
            </button>
          ))}
          {!hits.length ? <div style={{ padding: "var(--sp-8)", textAlign: "center", font: "var(--type-body)", color: "var(--text-faint)" }}>No matches on the Pi.</div> : null}
        </div>
      </div>
    </div>
  );
}

function SaveModal({ studio }: { studio: Studio }) {
  const { s, act } = studio;
  const exists = !!s.doc.id && s.pipelines.some(p => p.id === s.doc.id);
  const [name, setName] = useState(exists ? s.doc.name + " copy" : s.doc.name);
  const [description, setDescription] = useState(s.doc.description ?? "");
  const [isDefault, setIsDefault] = useState(false);
  const close = () => act.setModal(null);
  const go = async () => { const saved = await act.save({ name, description, isDefault, asNew: true }); if (saved) close(); };
  return (
    <Modal open eyebrow="Pipeline library" title="Save pipeline as" icon="save" width={460} onClose={close}
      footer={<><Button variant="ghost" onClick={close}>Cancel</Button><Button variant="primary" icon="save" onClick={go} disabled={!name.trim() || s.busy}>Save on the Pi</Button></>}>
      <Input label="Name" value={name} onChange={e => setName(e.target.value)} autoFocus onKeyDown={e => { if (e.key === "Enter") void go(); }} />
      <Input label="Description" placeholder="What this pipeline is for" value={description} onChange={e => setDescription(e.target.value)} />
      <Switch checked={isDefault} onChange={setIsDefault} label="Load on boot" description="The Pi arms this pipeline when it powers up" />
      <div style={{ display: "flex", gap: "var(--sp-3)", flexWrap: "wrap" }}>
        <Badge tone="neutral" icon="account_tree">{s.doc.nodes.length} nodes</Badge>
        <Badge tone="neutral" icon="cable">{s.doc.edges.length} cables</Badge>
      </div>
    </Modal>
  );
}

export function Modals({ studio, backend, reason }: { studio: Studio; backend: Backend; reason?: string }) {
  const { s, act, catalog } = studio;
  const n = s.doc.nodes.find(x => x.id === s.selected);
  const close = () => act.setModal(null);
  const defaultId = s.pipelines.find(p => p.isDefault)?.id ?? "";

  if (s.modal === "settings") {
    return (
      <Modal open eyebrow="GadgetCast" title="Settings" icon="settings" width={560} onClose={close}
        footer={<Button variant="secondary" onClick={close}>Done</Button>}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "var(--sp-6)" }}>
          <Select label="Theme" value={s.theme} options={[{ value: "light", label: "Light (default)" }, { value: "dark", label: "Dark (studio)" }]} onChange={e => act.setTheme(e.target.value === "dark" ? "dark" : "light")} />
          <Select label="Pipeline on boot" value={defaultId} options={[{ value: "", label: "None" }, ...s.pipelines.map(p => ({ value: p.id, label: p.name }))]} onChange={e => { if (e.target.value) void act.setDefault(e.target.value); }} />
        </div>
        <Switch checked={s.autosave} onChange={act.setAutosave} label="Auto-save to the Pi" description="Writes edits to an already-saved pipeline every 30 s" />
        <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-3)", padding: "var(--sp-5)", background: "var(--surface-input)", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)" }}>
          <span className="gc-caps">Connection</span>
          {backend.mode === "live" ? (
            <span style={{ font: "var(--type-meta)", color: "var(--text-body)" }}>{window.location.host} · API v{backend.version} · events {s.events}</span>
          ) : (
            <>
              <span style={{ font: "var(--type-meta)", color: "var(--st-warn)" }}>Offline demo · {reason ?? "the Pi did not answer"}</span>
              <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>Pipelines, devices and logs are bundled sample data held in this tab. Nothing reaches a device.</span>
              <div><Button size="sm" variant="secondary" icon="refresh" onClick={() => window.location.reload()}>Retry the Pi</Button></div>
            </>
          )}
        </div>
      </Modal>
    );
  }
  if (s.modal === "keys") {
    return (
      <Modal open eyebrow="Keyboard" title="Shortcuts" icon="keyboard" width={480} onClose={close} footer={<Button variant="secondary" onClick={close}>Done</Button>}>
        <div>
          {SHORTCUTS.map(([k, label]) => (
            <div key={k} style={{ display: "flex", alignItems: "center", gap: "var(--sp-5)", padding: "var(--sp-3) 0", borderBottom: "1px solid var(--line-hairline)" }}>
              <span style={{ flex: 1, font: "var(--type-ui)", color: "var(--text-primary)" }}>{label}</span>
              <kbd>{k}</kbd>
            </div>
          ))}
        </div>
      </Modal>
    );
  }
  if (s.modal === "save") return <SaveModal studio={studio} />;
  if (s.modal === "sendmic") return <SendMic live={backend.mode === "live"} onClose={close} />;
  if (s.modal === "capture") return <QuickCapture mode={s.captureMode} onClose={close} onDone={(msg) => act.toast(msg, "download", "ok")} />;
  if (s.modal === "delete" && s.deleteTarget) {
    const p = s.pipelines.find(x => x.id === s.deleteTarget);
    return (
      <Modal open eyebrow="Pipeline library" title={"Delete “" + (p?.name ?? s.deleteTarget) + "”?"} icon="delete" width={440} onClose={close}
        footer={<><Button variant="ghost" onClick={close}>Keep it</Button><Button variant="danger" icon="delete" onClick={() => act.deletePipeline(s.deleteTarget!)}>Delete from the Pi</Button></>}>
        <p style={{ font: "var(--type-body)", color: "var(--text-muted)" }}>The pipeline is removed from the Pi for every client. A running pipeline must be stopped first.</p>
      </Modal>
    );
  }
  if (s.modal === "discard") {
    return (
      <Modal open eyebrow="Unsaved changes" title="Discard edits to this pipeline?" icon="edit_off" width={440} onClose={() => act.up({ modal: null, pending: null })}
        footer={<><Button variant="ghost" onClick={() => act.up({ modal: null, pending: null })}>Keep editing</Button><Button variant="danger" onClick={act.confirmDiscard}>Discard</Button></>}>
        <p style={{ font: "var(--type-body)", color: "var(--text-muted)" }}>“{s.doc.name}” has changes that are not on the Pi yet.</p>
      </Modal>
    );
  }
  if (s.modal === "config" && n) {
    const v = viewOf(n, catalog, s.doc.edges);
    const set = (patch: Record<string, unknown>) => act.patchConfig(n.id, patch);
    return (
      <Modal open eyebrow={v.cls === "ai" ? "AI node" : v.group + " node"} title={v.title} icon={v.icon} width={720} onClose={close}
        footer={<Button variant="primary" icon="check" onClick={close}>Done</Button>}>
        <Input label="Node name" value={n.title ?? v.title} onChange={e => act.patchNode(n.id, { title: e.target.value })} />
        {v.cls === "ai" ? (
          <>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: "var(--sp-5)" }}>
              <Select label="Model" value={cfgStr(n, "model", AI_MODELS[0])} options={AI_MODELS} onChange={e => set({ model: e.target.value })} />
              <Select label="Modalities" value={cfgStr(n, "modalities", "Audio + frames")} options={["Audio + frames", "Audio only", "Frames only", "Text only"]} onChange={e => set({ modalities: e.target.value })} />
              <Select label="Voice" value={cfgStr(n, "voice", "Aoede")} options={["Aoede", "Charon", "Fenrir", "Kore", "Puck"]} onChange={e => set({ voice: e.target.value })} />
            </div>
            <label style={{ display: "flex", flexDirection: "column", gap: "var(--sp-3)" }}>
              <span className="gc-caps">System instructions</span>
              <textarea value={cfgStr(n, "prompt")} onChange={e => set({ prompt: e.target.value })} rows={7} className="gc-field gc-scroll" style={{ padding: "var(--sp-5)", font: "var(--type-meta)", lineHeight: 1.6, resize: "vertical" }} />
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: "var(--sp-3)" }}>
              <span className="gc-caps">Prompt template · per frame</span>
              <textarea value={cfgStr(n, "template", "Frame at {{timestamp}} from {{source}}. Transcript so far: {{transcript}}.\nReply only if something actionable changed.")}
                onChange={e => set({ template: e.target.value })} rows={3} className="gc-field" style={{ padding: "var(--sp-5)", font: "var(--type-meta)", lineHeight: 1.6, resize: "vertical" }} />
            </label>
            <div style={{ display: "flex", gap: "var(--sp-5)" }}>
              <Slider label="Frame sample rate" value={cfgNum(n, "fps", 1)} min={0.2} max={5} step={0.2} unit="fps" tone="var(--sig-ai)" onChange={x => set({ fps: Math.round(x * 10) / 10 })} style={{ flex: 1 }} />
              <Slider label="Interrupt sensitivity" value={cfgNum(n, "sensitivity", 0.4)} min={0} max={1} step={0.05} tone="var(--sig-ai)" onChange={x => set({ sensitivity: Math.round(x * 100) / 100 })} style={{ flex: 1 }} />
            </div>
            <p style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>Model calls run on the Pi with its own credentials. The Studio never holds an API key.</p>
          </>
        ) : (
          <>
            <Input label="Source path / URL" mono value={cfgStr(n, "url")} onChange={e => set({ url: e.target.value })} hint="Validated before the node joins the running graph" />
            <Select label="Format" value={cfgStr(n, "format", VIDEO_FORMATS[0])} options={VIDEO_FORMATS} onChange={e => set({ format: e.target.value })} />
            <label style={{ display: "flex", flexDirection: "column", gap: "var(--sp-3)" }}>
              <span className="gc-caps">Raw config (JSON)</span>
              <pre className="gc-field gc-scroll" style={{ margin: 0, padding: "var(--sp-5)", font: "var(--type-meta)", maxHeight: 160, overflow: "auto" }}>{JSON.stringify(n.config, null, 2)}</pre>
            </label>
          </>
        )}
      </Modal>
    );
  }
  return null;
}
