import { useEffect } from "react";
import { Badge, PresetRow, Tabs, Toast } from "../ds";
import type { Backend } from "../api/backend";
import { Canvas } from "./Canvas";
import { HardwareView } from "./HardwareView";
import { Header, StatusBar } from "./Header";
import { Inspector } from "./Inspector";
import { Modals, Palette } from "./Modals";
import { DiagnosticsView, MixerView } from "./Panels";
import { Rail, Sidebar } from "./Sidebar";
import { useStudio, type View } from "./store";

export function App({ backend, reason }: { backend: Backend; reason?: string }) {
  const studio = useStudio(backend);
  const { s, act } = studio;

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement;
      const typing = /input|textarea|select/i.test(t.tagName) || t.isContentEditable;
      const mod = e.metaKey || e.ctrlKey;
      const k = e.key.toLowerCase();
      if (mod && k === "k") { e.preventDefault(); act.setPalette(true); return; }
      if (mod && k === "s") { e.preventDefault(); if (s.doc.id) void act.save(); else act.setModal("save"); return; }
      if (mod && e.key === "Enter") { e.preventDefault(); if (s.status?.running) void act.stop(); else void act.goOnAir(); return; }
      if (e.key === "Escape") { act.up({ palette: false, presetMenu: false, ...(s.modal && s.modal !== "capture" ? { modal: null, pending: null } : {}) }); return; }
      if (typing || mod || e.altKey || s.palette || s.modal) return;
      const views: Record<string, View> = { "1": "canvas", "2": "mixer", "3": "diag", "4": "hardware" };
      if (views[e.key] && !e.shiftKey) act.setView(views[e.key]);
      else if (k === "r" && !e.shiftKey) act.revert();
      else if (e.key === "C" && e.shiftKey) act.openCapture("screenshot");
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [s.doc.id, s.status?.running, s.modal, s.palette, act]);

  const running = s.status?.running ?? null;
  return (
    <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", background: "var(--surface-app)", overflow: "hidden" }}>
      <Header studio={studio} backend={backend} reason={reason} />
      <div style={{ height: "var(--h-subbar)", flex: "0 0 auto", display: "flex", alignItems: "center", gap: "var(--sp-6)", padding: "0 var(--sp-6)", background: "var(--surface-panel)", borderBottom: "1px solid var(--line-border)" }}>
        <Tabs<View> size="sm" value={s.view} onChange={act.setView} style={{ border: 0 }}
          items={[
            { value: "canvas", label: "Patch canvas", icon: "account_tree" },
            { value: "mixer", label: "Audio + Vision mixer", icon: "tune" },
            { value: "diag", label: "Diagnostics", icon: "monitor_heart", count: s.logs.length },
            { value: "hardware", label: "Hardware", icon: "memory" },
          ]} />
        <div style={{ flex: 1 }} />
        {backend.mode === "demo" ? <Badge tone="warn" dot>Simulated media</Badge> : <Badge tone={s.events === "open" ? "ok" : "warn"} dot>{s.events === "open" ? "Events live" : "Events " + s.events}</Badge>}
        <Badge tone="neutral" icon="groups">{s.status?.clients ?? 0} clients</Badge>
        {s.dirty ? <Badge tone="warn" icon="edit">Unsaved changes</Badge> : s.doc.id ? <Badge tone="ok" icon="cloud_done">Saved</Badge> : <Badge tone="neutral" icon="cloud_off">Not saved</Badge>}
      </div>
      <div style={{ flex: 1, display: "flex", minHeight: 0, position: "relative" }}>
        {s.view === "canvas" ? (
          <>
            <Canvas studio={studio} />
            <Sidebar studio={studio} backend={backend} floating />
            <Inspector studio={studio} floating />
          </>
        ) : (
          <>
            <Rail studio={studio} />
            <Sidebar studio={studio} backend={backend} />
            {s.view === "mixer" ? <MixerView studio={studio} /> : s.view === "diag" ? <DiagnosticsView studio={studio} backend={backend} /> : <HardwareView studio={studio} />}
          </>
        )}
        {s.presetMenu ? (
          <>
            <div style={{ position: "absolute", inset: 0, zIndex: 69 }} onClick={() => act.setPresetMenu(false)} />
            <div role="menu" style={{ position: "absolute", left: 220, top: 6, width: 288, zIndex: 70, padding: "var(--sp-3)", background: "var(--surface-raised)", border: "1px solid var(--line-strong)", borderRadius: "var(--r-md)", boxShadow: "var(--el-4)", maxHeight: "70%", overflow: "auto" }} className="gc-scroll">
              <div className="gc-caps" style={{ padding: "var(--sp-3) var(--sp-4)" }}>Pipelines on the Pi</div>
              {s.pipelines.map(p => (
                <PresetRow key={p.id} name={p.name} meta={p.nodes.length + " nodes · " + p.edges.length + " cables"} selected={p.id === s.doc.id} running={running === p.id} isDefault={!!p.isDefault} onClick={() => act.openPipeline(p.id)} />
              ))}
              <PresetRow name="New pipeline" icon="add" meta="empty canvas" onClick={() => act.openPipeline(null)} />
            </div>
          </>
        ) : null}
      </div>
      <StatusBar studio={studio} backend={backend} />
      {s.palette ? <Palette studio={studio} /> : null}
      <Modals studio={studio} backend={backend} reason={reason} />
      {s.toast ? (
        <div style={{ position: "absolute", bottom: 40, left: "50%", transform: "translateX(-50%)", zIndex: 95 }}>
          <Toast key={s.toast.id} tone={s.toast.tone} icon={s.toast.icon} onDismiss={() => act.up({ toast: null })}>{s.toast.msg}</Toast>
        </div>
      ) : null}
    </div>
  );
}
