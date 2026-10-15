import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Backend, EventsState } from "../api/backend";
import type { DevicesResponse, NodeGroup, Pipeline, PortKind, ServerEvent, Status } from "../api/types";
import type { LogLine, ToastTone } from "../ds";
import { graphToPipeline, parseImportedPipeline, pipelineToGraph } from "../graph/mapping";
import { hasGadgetSink, indexCatalog, newEdgeId, newNodeId, slugId, type Catalog, type GraphDoc, type GraphNode, type PortRef } from "../graph/model";
import { graphProblems } from "../graph/validate";

export type View = "canvas" | "mixer" | "diag" | "hardware";
export type RailMode = "nodes" | "library" | "devices" | "help";
export type ModalKind = null | "settings" | "keys" | "save" | "config" | "delete" | "discard" | "capture";
export type CaptureMode = "screenshot" | "video" | "audio";
export type Theme = "light" | "dark";
export interface ToastMsg { id: number; msg: string; icon?: string; tone: ToastTone }

export interface StudioState {
  theme: Theme;
  view: View;
  rail: RailMode;
  doc: GraphDoc;
  dirty: boolean;
  selected: string | null;
  groups: NodeGroup[];
  pipelines: Pipeline[];
  status: Status | null;
  devices: DevicesResponse | null;
  logs: LogLine[];
  follow: boolean;
  events: EventsState;
  modal: ModalKind;
  palette: boolean;
  presetMenu: boolean;
  toast: ToastMsg | null;
  autosave: boolean;
  captureMode: CaptureMode;
  /** Pipeline id the delete modal is asking about. */
  deleteTarget: string | null;
  /** Action the discard-changes modal will run on confirm. */
  pending: (() => void) | null;
  busy: boolean;
  levels: number[];
  peaks: number[];
}

const EMPTY_DOC: GraphDoc = { id: null, name: "Untitled pipeline", nodes: [], edges: [] };
const MAX_LOGS = 500;

function readTheme(): Theme {
  try { return localStorage.getItem("gadgetcast:theme") === "dark" ? "dark" : "light"; } catch { return "light"; }
}
function stamp(ts: string): string {
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return ts;
  return d.toTimeString().slice(0, 8) + "." + String(d.getMilliseconds()).padStart(3, "0");
}
let logSeq = 0;
function toLogLine(ev: Extract<ServerEvent, { type: "log" }>): LogLine {
  const level = ev.data.level === "error" ? "err" : ev.data.level === "warn" ? "warn" : "info";
  return { id: ++logSeq, t: stamp(ev.ts), src: ev.data.subsystem, level, msg: ev.data.msg };
}
const errText = (e: unknown) => (e instanceof Error ? e.message : String(e));

export function useStudio(backend: Backend) {
  const [s, setS] = useState<StudioState>(() => ({
    theme: readTheme(), view: "canvas", rail: "nodes", doc: EMPTY_DOC, dirty: false, selected: null,
    groups: [], pipelines: [], status: null, devices: null, logs: [], follow: true, events: "connecting",
    modal: null, palette: false, presetMenu: false, toast: null, autosave: true, captureMode: "screenshot",
    deleteTarget: null, pending: null, busy: false, levels: [0, 0], peaks: [0, 0],
  }));
  const ref = useRef(s);
  ref.current = s;
  const up = useCallback((patch: Partial<StudioState> | ((p: StudioState) => Partial<StudioState>)) => {
    setS(p => ({ ...p, ...(typeof patch === "function" ? patch(p) : patch) }));
  }, []);
  const catalog: Catalog = useMemo(() => indexCatalog(s.groups), [s.groups]);
  const catRef = useRef(catalog);
  catRef.current = catalog;

  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const toast = useCallback((msg: string, icon?: string, tone: ToastTone = "neutral") => {
    const id = Date.now();
    up({ toast: { id, msg, icon, tone } });
    if (toastTimer.current) clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => up(p => (p.toast?.id === id ? { toast: null } : {})), tone === "err" ? 4200 : 2400);
  }, [up]);

  /* ── doc edits ── */
  const editDoc = useCallback((fn: (d: GraphDoc) => GraphDoc, extra?: (p: StudioState) => Partial<StudioState>) => {
    up(p => ({ doc: fn(p.doc), dirty: true, ...(extra ? extra(p) : {}) }));
  }, [up]);

  const loadDoc = useCallback((p: Pipeline | null) => {
    if (!p) { up({ doc: { ...EMPTY_DOC }, dirty: false, selected: null, presetMenu: false }); return; }
    const { doc, dropped } = pipelineToGraph(p);
    up({ doc, dirty: false, selected: null, presetMenu: false });
    if (dropped) toast(dropped + " cable(s) referenced missing nodes and were dropped", "warning", "warn");
  }, [up, toast]);

  const refreshPipelines = useCallback(async () => {
    const list = await backend.listPipelines();
    up({ pipelines: list });
    return list;
  }, [backend, up]);

  const refreshDevices = useCallback(async () => {
    try { up({ devices: await backend.getDevices() }); } catch (e) { toast("Devices: " + errText(e), "error", "err"); }
  }, [backend, up, toast]);

  /* ── boot ── */
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const [groups, pipelines, status, devices] = await Promise.all([
          backend.listNodes(), backend.listPipelines(), backend.getStatus().catch(() => null), backend.getDevices().catch(() => null),
        ]);
        if (cancelled) return;
        up({ groups, pipelines, status, devices });
        const pick = pipelines.find(p => p.id === status?.running) ?? pipelines.find(p => p.isDefault) ?? pipelines[0] ?? null;
        loadDoc(pick);
      } catch (e) {
        if (!cancelled) toast("Could not load from the Pi: " + errText(e), "error", "err");
      }
    })();
    return () => { cancelled = true; };
  }, [backend, up, loadDoc, toast]);

  /* ── server push ── */
  useEffect(() => backend.subscribe(ev => {
    if (ev.type === "metrics") up(p => ({ status: p.status ? { ...p.status, metrics: ev.data } : { running: null, onAir: null, clients: 0, metrics: ev.data } }));
    else if (ev.type === "status") up({ status: ev.data });
    else up(p => ({ logs: p.logs.concat([toLogLine(ev)]).slice(-MAX_LOGS) }));
  }, events => up({ events })), [backend, up]);

  /* ── demo-only meter simulation (the v0 API reports no audio levels) ── */
  useEffect(() => {
    const sim = backend.simulatedLevels;
    if (!sim) return undefined;
    const t = setInterval(() => up(sim()), 700);
    return () => clearInterval(t);
  }, [backend, up]);

  /* ── theme ── */
  useEffect(() => {
    document.documentElement.setAttribute("data-theme", s.theme);
    try { localStorage.setItem("gadgetcast:theme", s.theme); } catch { /* storage unavailable */ }
  }, [s.theme]);

  /* ── save / run ── */
  const save = useCallback(async (opts: { name?: string; description?: string; isDefault?: boolean; asNew?: boolean; quiet?: boolean } = {}) => {
    const cur = ref.current;
    const name = (opts.name ?? cur.doc.name).trim() || "Untitled pipeline";
    const taken = cur.pipelines.map(p => p.id);
    const id = !opts.asNew && cur.doc.id ? cur.doc.id : slugId(name, taken);
    const doc: GraphDoc = { ...cur.doc, name, description: opts.description ?? cur.doc.description, isDefault: opts.isDefault ?? cur.doc.isDefault ?? false };
    up({ busy: true });
    try {
      const saved = await backend.savePipeline(graphToPipeline(doc, id));
      if (saved.isDefault) {
        const others = cur.pipelines.filter(p => p.isDefault && p.id !== saved.id);
        await Promise.all(others.map(p => backend.savePipeline({ ...p, isDefault: false })));
      }
      await refreshPipelines();
      up(p => ({ doc: { ...p.doc, id: saved.id, name: saved.name, description: saved.description, isDefault: saved.isDefault, updatedAt: saved.updatedAt }, dirty: p.doc === cur.doc ? false : p.dirty }));
      if (!opts.quiet) toast("Saved “" + saved.name + "” on the Pi", "save", "ok");
      return saved;
    } catch (e) {
      toast("Save failed: " + errText(e), "error", "err");
      return null;
    } finally {
      up({ busy: false });
    }
  }, [backend, up, toast, refreshPipelines]);

  const goOnAir = useCallback(async () => {
    const cur = ref.current;
    if (!hasGadgetSink(cur.doc, catRef.current)) { toast("Patch a USB gadget sink before going on air", "usb", "warn"); return; }
    let id = cur.doc.id;
    if (!id || cur.dirty) { const saved = await save({ quiet: true }); if (!saved) return; id = saved.id; }
    try {
      const st = await backend.runPipeline(id);
      up({ status: st });
      toast("“" + ref.current.doc.name + "” is running on the Pi", "usb", "live");
    } catch (e) { toast("Run failed: " + errText(e), "error", "err"); }
  }, [backend, save, toast, up]);

  const stop = useCallback(async () => {
    try { up({ status: await backend.stopPipeline() }); toast("Pipeline stopped · gadget released", "stop_circle"); } catch (e) { toast("Stop failed: " + errText(e), "error", "err"); }
  }, [backend, toast, up]);

  /* Guard: run `fn` now, or after the user agrees to discard unsaved edits. */
  const guard = useCallback((fn: () => void) => {
    if (ref.current.dirty && ref.current.doc.nodes.length) up({ modal: "discard", pending: fn, presetMenu: false, palette: false });
    else fn();
  }, [up]);

  const openPipeline = useCallback((id: string | null) => guard(async () => {
    if (!id) { loadDoc(null); return; }
    try { const p = await backend.getPipeline(id); loadDoc(p); toast("Loaded “" + p.name + "”", "account_tree"); }
    catch (e) { toast("Open failed: " + errText(e), "error", "err"); }
  }), [backend, guard, loadDoc, toast]);

  const act = useMemo(() => ({
    up, toast, save, goOnAir, stop, openPipeline, refreshDevices,
    setView: (view: View) => up({ view }),
    setRail: (rail: RailMode) => up({ rail }),
    setModal: (modal: ModalKind) => up({ modal, presetMenu: false }),
    setPalette: (palette: boolean) => up({ palette }),
    setPresetMenu: (presetMenu: boolean) => up({ presetMenu }),
    setTheme: (theme: Theme) => up({ theme }),
    toggleTheme: () => up(p => ({ theme: p.theme === "light" ? "dark" : "light" })),
    setFollow: (follow: boolean) => up({ follow }),
    setAutosave: (autosave: boolean) => up({ autosave }),
    openCapture: (captureMode: CaptureMode) => up({ modal: "capture", captureMode }),
    select: (selected: string | null) => up({ selected }),
    clearLogs: () => up({ logs: [] }),

    moveNodes: (list: { id: string; x: number; y: number }[]) => editDoc(d => {
      const m = new Map(list.map(l => [l.id, l]));
      return { ...d, nodes: d.nodes.map(n => { const l = m.get(n.id); return l ? { ...n, x: l.x, y: l.y } : n; }) };
    }),
    removeNodes: (ids: string[]) => editDoc(
      d => ({ ...d, nodes: d.nodes.filter(n => !ids.includes(n.id)), edges: d.edges.filter(e => !ids.includes(e.from.node) && !ids.includes(e.to.node)) }),
      p => ({ selected: p.selected && ids.includes(p.selected) ? null : p.selected })),
    removeEdge: (id: string) => editDoc(d => ({ ...d, edges: d.edges.filter(e => e.id !== id) })),
    patchNode: (id: string, patch: Partial<Pick<GraphNode, "title" | "enabled">>) => editDoc(d => ({ ...d, nodes: d.nodes.map(n => (n.id === id ? { ...n, ...patch } : n)) })),
    patchConfig: (id: string, patch: Record<string, unknown>) => editDoc(d => ({ ...d, nodes: d.nodes.map(n => (n.id === id ? { ...n, config: { ...n.config, ...patch } } : n)) })),
    toggleNode: (id: string) => editDoc(d => ({ ...d, nodes: d.nodes.map(n => (n.id === id ? { ...n, enabled: !n.enabled } : n)) })),
    clearLinks: (id: string) => editDoc(d => ({ ...d, edges: d.edges.filter(e => e.from.node !== id && e.to.node !== id) })),
    renameDoc: (name: string) => editDoc(d => ({ ...d, name })),
    duplicate: (ids: string[]) => {
      let last: string | null = null;
      editDoc(d => {
        const nodes = d.nodes.slice();
        ids.forEach(id => {
          const n = d.nodes.find(x => x.id === id); if (!n) return;
          const c: GraphNode = { ...n, id: newNodeId(n.type, nodes.map(x => x.id)), x: n.x + 32, y: n.y + 32, config: { ...n.config } };
          nodes.push(c); last = c.id;
        });
        return { ...d, nodes };
      }, () => ({ selected: last }));
    },
    addNode: (type: string, x = 416, y = 336, link?: { from: PortRef; kind: PortKind }) => {
      const t = catRef.current.get(type);
      editDoc(d => {
        const id = newNodeId(type, d.nodes.map(n => n.id));
        const config: Record<string, unknown> = {};
        if (t?.url) config.url = t.url;
        const n: GraphNode = { id, type, x, y, title: t?.title, enabled: true, config };
        const port = link ? (t?.in ?? []).findIndex(p => p[0] === link.kind) : -1;
        const edges = link && port >= 0 ? d.edges.concat([{ id: newEdgeId(d.edges.map(e => e.id)), from: link.from, to: { node: id, port }, kind: link.kind }]) : d.edges;
        return { ...d, nodes: d.nodes.concat([n]), edges };
      }, p => ({ selected: newNodeId(type, p.doc.nodes.map(n => n.id)) }));
    },
    connect: (from: PortRef, to: PortRef, kind: PortKind, replaceId?: string) => editDoc(d => {
      const edges = replaceId ? d.edges.filter(e => e.id !== replaceId) : d.edges;
      return { ...d, edges: edges.concat([{ id: replaceId ?? newEdgeId(d.edges.map(e => e.id)), from, to, kind }]) };
    }),

    revert: () => guard(async () => {
      const id = ref.current.doc.id;
      if (!id) { loadDoc(null); return; }
      try { loadDoc(await backend.getPipeline(id)); toast("Reverted to the saved pipeline", "restart_alt"); } catch (e) { toast(errText(e), "error", "err"); }
    }),
    confirmDiscard: () => { const fn = ref.current.pending; up({ modal: null, pending: null, dirty: false }); fn?.(); },
    askDelete: (id: string) => up({ modal: "delete", deleteTarget: id }),
    deletePipeline: async (id: string) => {
      try {
        await backend.deletePipeline(id);
        const list = await refreshPipelines();
        up({ modal: null, deleteTarget: null });
        if (ref.current.doc.id === id) up(p => ({ doc: { ...p.doc, id: null }, dirty: true }));
        toast("Deleted “" + (list.find(p => p.id === id)?.name ?? id) + "”", "delete");
      } catch (e) { toast("Delete failed: " + errText(e), "error", "err"); }
    },
    setDefault: async (id: string) => {
      try {
        const list = ref.current.pipelines;
        const target = await backend.getPipeline(id);
        await backend.savePipeline({ ...target, isDefault: true });
        await Promise.all(list.filter(p => p.isDefault && p.id !== id).map(p => backend.savePipeline({ ...p, isDefault: false })));
        await refreshPipelines();
        if (ref.current.doc.id === id) up(p => ({ doc: { ...p.doc, isDefault: true } }));
        toast("“" + target.name + "” loads on boot", "push_pin", "ok");
      } catch (e) { toast("Could not set the boot default: " + errText(e), "error", "err"); }
    },
    exportJson: () => {
      const cur = ref.current;
      const p = graphToPipeline(cur.doc, cur.doc.id ?? slugId(cur.doc.name, []));
      const blob = new Blob([JSON.stringify(p, null, 2)], { type: "application/json" });
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob); a.download = p.id + ".json"; a.click();
      setTimeout(() => URL.revokeObjectURL(a.href), 1000);
      toast("Pipeline exported as JSON", "download");
    },
    importJson: (file: File) => guard(() => {
      const r = new FileReader();
      r.onload = () => {
        try {
          const p = parseImportedPipeline(JSON.parse(String(r.result)));
          const { doc } = pipelineToGraph(p);
          const problems = graphProblems(doc, catRef.current);
          up({ doc: { ...doc, id: null, isDefault: false }, dirty: true, selected: null });
          if (problems.length) toast("Imported with " + problems.length + " problem(s): " + problems[0], "warning", "warn");
          else toast("Imported “" + doc.name + "” · save to keep it on the Pi", "upload", "ok");
        } catch (e) { toast("Invalid pipeline JSON: " + errText(e), "error", "err"); }
      };
      r.readAsText(file);
    }),
  }), [up, toast, save, goOnAir, stop, openPipeline, refreshDevices, editDoc, guard, loadDoc, backend, refreshPipelines]);

  /* ── autosave: write an already-saved pipeline back to the Pi every 30 s while it has edits ── */
  useEffect(() => {
    if (!s.autosave) return undefined;
    const t = setInterval(() => {
      const cur = ref.current;
      if (cur.dirty && cur.doc.id && cur.pipelines.some(p => p.id === cur.doc.id) && !cur.busy) void save({ quiet: true });
    }, 30000);
    return () => clearInterval(t);
  }, [s.autosave, save]);

  return { s, act, catalog };
}

export type Studio = ReturnType<typeof useStudio>;
export type Actions = Studio["act"];
