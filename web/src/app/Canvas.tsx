/* Patch canvas on React Flow v11. React Flow owns pan / zoom / node drag / marquee / minimap / connection
   radius / edge reconnection; the design system owns every pixel through custom node + edge types
   (PipelineNode, PatchWire) and the connection line. Extras: typed-port validation with a rejection bubble,
   quick-add on a cable dropped on empty canvas, palette drop, floating chrome, context menus. */
import { memo, useCallback, useEffect, useMemo, useRef, useState, type CSSProperties, type DragEvent, type MouseEvent as ReactMouseEvent } from "react";
import ReactFlow, {
  Background, BackgroundVariant, EdgeLabelRenderer, Handle, MiniMap, Position, ReactFlowProvider, SelectionMode, useReactFlow, useStore, useViewport,
  type Connection, type ConnectionLineComponentProps, type Edge, type EdgeChange, type EdgeProps, type Node, type NodeChange, type NodeProps, type OnConnectStartParams,
} from "reactflow";
import "reactflow/dist/style.css";
import { Badge, Button, Icon, IconButton, Input, LevelMeter, PatchWire, PipelineNode, Tooltip, WaveScope, type SignalKind } from "../ds";
import type { NodeType, PortKind } from "../api/types";
import type { GraphNode, PortRef } from "../graph/model";
import { hasGadgetSink } from "../graph/model";
import { connectionError, handleId, lookupFor, parseHandle } from "../graph/validate";
import { cfgStr, sigTint, sigVar, viewOf, type NodeView } from "./nodes";
import type { Studio } from "./store";

const GRID = 16, NODE_W = 224;
const snapG = (v: number) => Math.round(v / GRID) * GRID;
/** Shared with the connection-line component, which React Flow renders outside our props. */
const HOT: { kind: SignalKind | null; reject: string | null } = { kind: null, reject: null };
const handleStyle: CSSProperties = {
  position: "absolute", top: 0, left: 0, right: "auto", bottom: "auto", width: "100%", height: "100%", minWidth: 0, minHeight: 0,
  transform: "none", background: "transparent", border: 0, borderRadius: 0, opacity: 1,
};
const panel: CSSProperties = { background: "var(--surface-panel)", border: "1px solid var(--line-border)", borderRadius: "var(--r-lg)" };

interface GcNodeData {
  n: GraphNode;
  view: NodeView;
  hotKind: SignalKind | null;
  live: boolean;
  inConnected: boolean[];
  outConnected: boolean[];
  levels: number[];
  peaks: number[];
}
interface GcEdgeData { kind: PortKind; flowing: boolean; dimmed: boolean; onDelete: (id: string) => void }

/* ── custom node ── */
const GcNode = memo(function GcNode({ data, selected }: NodeProps<GcNodeData>) {
  const { n, view } = data;
  const renderHandle = (dir: "in" | "out", i: number) => (
    <Handle type={dir === "in" ? "target" : "source"} position={dir === "in" ? Position.Left : Position.Right} id={handleId(dir, i)} style={handleStyle} />
  );
  const prompt = cfgStr(n, "prompt");
  return (
    <PipelineNode title={view.title} subtitle={view.subtitle} icon={view.icon} nodeClass={view.cls} hotKind={data.hotKind} renderHandle={renderHandle}
      inputs={view.inputs.map((p, i) => ({ ...p, connected: data.inConnected[i] }))}
      outputs={view.outputs.map((p, i) => ({ ...p, connected: data.outConnected[i] }))}
      state={!view.known ? "error" : n.enabled && data.live ? "live" : "idle"} disabled={!n.enabled} selected={selected}
      badge={view.cls === "ai" ? <Badge tone="ai" icon="auto_awesome">{data.live ? "Streaming" : "Configured"}</Badge> : !view.known ? <Badge tone="err" icon="help">Unknown type</Badge> : null}>
      {n.type === "uac_out" || n.type === "mixer_a" ? <LevelMeter channels={data.levels} peaks={data.peaks} showScale={false} width={180} /> : null}
      {n.type === "browser_mic" || n.type === "bt_src" ? <WaveScope height={32} live={data.live} gain={data.live ? 1 : 0.15} /> : null}
      {view.cls === "ai" && prompt ? (
        <div style={{ font: "var(--fw-regular) var(--fs-meta)/1.4 var(--font-sans)", color: "var(--text-muted)", background: "var(--surface-input)", borderRadius: "var(--r-sm)", padding: "6px 8px", maxHeight: 36, overflow: "hidden" }}>{prompt.slice(0, 70) + (prompt.length > 70 ? "…" : "")}</div>
      ) : null}
    </PipelineNode>
  );
});

/* ── custom edge ── */
function GcEdge({ id, sourceX, sourceY, targetX, targetY, selected, data }: EdgeProps<GcEdgeData>) {
  const [hov, setHov] = useState(false);
  if (!data) return null;
  return (
    <>
      <PatchWire from={{ x: sourceX, y: sourceY }} to={{ x: targetX, y: targetY }} kind={data.kind} flowing={data.flowing} dimmed={data.dimmed} selected={selected} hovered={hov}
        onPointerEnter={() => setHov(true)} onPointerLeave={() => setHov(false)} onClick={() => {}} />
      {selected ? (
        <EdgeLabelRenderer>
          <button type="button" className="nopan" onClick={() => data.onDelete(id)} title="Delete cable" aria-label="Delete cable"
            style={{ position: "absolute", transform: "translate(-50%,-50%) translate(" + (sourceX + targetX) / 2 + "px," + (sourceY + targetY) / 2 + "px)", pointerEvents: "all", width: 22, height: 22, borderRadius: "50%", border: "1px solid var(--brand)", background: "var(--surface-panel)", color: "var(--brand)", display: "grid", placeItems: "center", cursor: "pointer", padding: 0 }}>
            <Icon name="close" size="xs" />
          </button>
        </EdgeLabelRenderer>
      ) : null}
    </>
  );
}

function ConnLine({ fromX, fromY, toX, toY, connectionStatus }: ConnectionLineComponentProps) {
  return <PatchWire from={{ x: fromX, y: fromY }} to={{ x: toX, y: toY }} kind={HOT.kind || "data"} invalid={connectionStatus === "invalid"} flowing={connectionStatus === "valid"} />;
}

const nodeTypes = { gc: GcNode };
const edgeTypes = { gc: GcEdge };

/* ── floating chrome (needs React Flow context) ── */
function Chrome({ studio, tool, setTool, showMap, setShowMap, openQuick, fitAll, valid }: {
  studio: Studio; tool: "select" | "pan"; setTool: (t: "select" | "pan") => void; showMap: boolean; setShowMap: (f: (m: boolean) => boolean) => void;
  openQuick: () => void; fitAll: () => void; valid: boolean;
}) {
  const { s, act } = studio;
  const rf = useReactFlow();
  const { zoom } = useViewport();
  const sep = <span style={{ width: 1, height: 18, background: "var(--line-hairline)", margin: "0 4px" }} />;
  const running = s.status?.running ?? null;
  return (
    <>
      <div data-ui className="nopan nowheel" style={{ position: "absolute", left: "50%", bottom: 16, transform: "translateX(-50%)", display: "flex", alignItems: "center", gap: 2, padding: 4, ...panel, zIndex: 5 }}>
        <Tooltip label="Select" shortcut="V"><IconButton icon="near_me" label="Select" size="sm" active={tool === "select"} onClick={() => setTool("select")} /></Tooltip>
        <Tooltip label="Hand, or hold Space" shortcut="H"><IconButton icon="pan_tool" label="Pan" size="sm" active={tool === "pan"} onClick={() => setTool("pan")} /></Tooltip>
        {sep}
        <Tooltip label="Add node" shortcut="Dbl-click"><IconButton icon="add_box" label="Add node" size="sm" onClick={openQuick} /></Tooltip>
        {sep}
        <IconButton icon="remove" label="Zoom out" size="sm" onClick={() => rf.zoomOut({ duration: 120 })} />
        <button type="button" onClick={() => rf.zoomTo(1, { duration: 120 })} title="Reset zoom"
          style={{ font: "var(--fw-medium) var(--fs-ui)/1 var(--font-sans)", color: "var(--text-body)", width: 44, height: 28, background: "transparent", border: 0, borderRadius: "var(--r-sm)", cursor: "pointer" }}>{Math.round(zoom * 100)}%</button>
        <IconButton icon="add" label="Zoom in" size="sm" onClick={() => rf.zoomIn({ duration: 120 })} />
        <Tooltip label="Zoom to fit" shortcut="⇧1"><IconButton icon="fit_screen" label="Fit" size="sm" onClick={fitAll} /></Tooltip>
        {sep}
        <Tooltip label="Minimap"><IconButton icon="map" label="Minimap" size="sm" active={showMap} onClick={() => setShowMap(m => !m)} /></Tooltip>
      </div>
      <div data-ui className="nopan" style={{ position: "absolute", right: 16, top: 16, display: "flex", alignItems: "center", gap: 8, padding: "4px 4px 4px 12px", ...panel, zIndex: 5 }}>
        <Badge tone={valid ? "ok" : "warn"} dot>{valid ? "Graph valid" : "No gadget sink"}</Badge>
        <span style={{ font: "var(--fw-regular) var(--fs-ui)/1 var(--font-sans)", color: "var(--text-faint)", whiteSpace: "nowrap" }}>{s.doc.nodes.length} nodes · {s.doc.edges.length} cables</span>
        {sep}
        <Button size="sm" variant="ghost" icon="restart_alt" onClick={act.revert} disabled={!s.dirty}>Revert</Button>
        <Button size="sm" variant="secondary" icon="save" onClick={() => (s.doc.id ? act.save() : act.setModal("save"))} disabled={s.busy}>Save</Button>
        {running && running === s.doc.id
          ? <Button size="sm" variant="danger" icon="stop" onClick={act.stop}>Stop</Button>
          : <Button size="sm" variant="primary" icon="switch_video" onClick={act.goOnAir} disabled={s.busy}>Take</Button>}
      </div>
    </>
  );
}

interface QuickState { sx: number; sy: number; wx: number; wy: number; q: string; from?: PortRef; kind?: SignalKind | null }
interface MenuState { x: number; y: number; node?: string; edge?: string }

function CanvasInner({ studio }: { studio: Studio }) {
  const { s, act, catalog } = studio;
  const rf = useReactFlow();
  const wrapRef = useRef<HTMLDivElement>(null);
  const [tool, setTool] = useState<"select" | "pan">("select");
  const [sel, setSel] = useState<string[]>([]);
  const [selEdge, setSelEdge] = useState<string | null>(null);
  const [hotKind, setHotKind] = useState<SignalKind | null>(null);
  const [menu, setMenu] = useState<MenuState | null>(null);
  const [quick, setQuick] = useState<QuickState | null>(null);
  const [bubble, setBubble] = useState<{ x: number; y: number; why: string } | null>(null);
  const [ghost, setGhost] = useState<{ x: number; y: number } | null>(null);
  const [showMap, setShowMap] = useState(true);
  const pendingFrom = useRef<PortRef | null>(null);
  const reconnecting = useRef<string | null>(null);
  const bubbleTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const inited = useStore(st => st.nodeInternals.size > 0 && Array.from(st.nodeInternals.values()).every(n => n.width));
  const fitted = useRef<string | null | undefined>(undefined);

  const doc = s.doc;
  const live = s.status?.running != null && s.status.running === doc.id;
  const lookup = useMemo(() => lookupFor(doc, catalog), [doc, catalog]);
  const size = () => wrapRef.current ? wrapRef.current.getBoundingClientRect() : { width: 1200, height: 800, left: 0, top: 0 };
  const local = (cx: number, cy: number) => { const r = size(); return { x: cx - r.left, y: cy - r.top }; };

  /* Fit is hand-computed from node bounds so the graph clears the floating sidebar and inspector
     (fitView is async, and a synchronous getViewport after it reads stale numbers). */
  const fitAll = useCallback(() => {
    const nodes = rf.getNodes(); if (!nodes.length) return;
    const bx = Math.min(...nodes.map(n => n.position.x)), by = Math.min(...nodes.map(n => n.position.y));
    const bx2 = Math.max(...nodes.map(n => n.position.x + (n.width || NODE_W))), by2 = Math.max(...nodes.map(n => n.position.y + (n.height || 80)));
    const bw = Math.max(1, bx2 - bx), bh = Math.max(1, by2 - by);
    const { width: w, height: h } = size();
    const L = 284, R = s.selected ? 320 : 32, T = 72, B = 72;
    const aw = Math.max(200, w - L - R), ah = Math.max(160, h - T - B);
    const k = Math.min(1.2, Math.max(0.55, Math.min(aw / bw, ah / bh)));
    rf.setViewport({ x: L + (aw - bw * k) / 2 - bx * k, y: T + (ah - bh * k) / 2 - by * k, zoom: k }, { duration: 0 });
  }, [rf, s.selected]);
  useEffect(() => {
    if (inited && fitted.current !== doc.id) { fitted.current = doc.id; fitAll(); }
  }, [inited, doc.id, fitAll]);

  /* ── React Flow data derived from the graph ── */
  const rfNodes: Node<GcNodeData>[] = useMemo(() => doc.nodes.map(n => {
    const view = viewOf(n, catalog, doc.edges);
    return {
      id: n.id, type: "gc", position: { x: n.x, y: n.y }, selected: sel.includes(n.id) || s.selected === n.id,
      data: {
        n, view, hotKind, live, levels: s.levels, peaks: s.peaks,
        inConnected: view.inputs.map((_, i) => doc.edges.some(e => e.to.node === n.id && e.to.port === i)),
        outConnected: view.outputs.map((_, i) => doc.edges.some(e => e.from.node === n.id && e.from.port === i)),
      },
    };
  }), [doc.nodes, doc.edges, catalog, sel, s.selected, hotKind, live, s.levels, s.peaks]);

  const rfEdges: Edge<GcEdgeData>[] = useMemo(() => {
    const byId = new Map(doc.nodes.map(n => [n.id, n]));
    return doc.edges.map(ed => {
      const a = byId.get(ed.from.node), b = byId.get(ed.to.node);
      return {
        id: ed.id, type: "gc", source: ed.from.node, sourceHandle: handleId("out", ed.from.port), target: ed.to.node, targetHandle: handleId("in", ed.to.port),
        selected: selEdge === ed.id, updatable: true,
        data: {
          kind: ed.kind, flowing: !!(live && a?.enabled && b?.enabled),
          dimmed: !!s.selected && s.selected !== ed.from.node && s.selected !== ed.to.node && selEdge !== ed.id,
          onDelete: (id: string) => { act.removeEdge(id); setSelEdge(null); },
        },
      };
    });
  }, [doc.edges, doc.nodes, live, s.selected, selEdge, act]);

  const onNodesChange = (changes: NodeChange[]) => {
    const moves: { id: string; x: number; y: number }[] = [];
    changes.forEach(c => { if (c.type === "position" && c.position) moves.push({ id: c.id, x: c.position.x, y: c.position.y }); });
    if (moves.length) act.moveNodes(moves);
    const selCh = changes.filter((c): c is Extract<NodeChange, { type: "select" }> => c.type === "select");
    if (selCh.length) {
      setSel(prev => {
        let next = prev.slice();
        selCh.forEach(c => { next = c.selected ? (next.includes(c.id) ? next : next.concat([c.id])) : next.filter(x => x !== c.id); });
        const last = selCh.find(c => c.selected);
        if (last) { act.select(last.id); setSelEdge(null); } else if (!next.length) act.select(null);
        return next;
      });
    }
    const rm = changes.filter(c => c.type === "remove").map(c => (c as { id: string }).id);
    if (rm.length) { act.removeNodes(rm); setSel(x => x.filter(id => !rm.includes(id))); }
  };
  const onEdgesChange = (changes: EdgeChange[]) => {
    changes.forEach(c => {
      if (c.type === "select") { if (c.selected) { setSelEdge(c.id); setSel([]); act.select(null); } else setSelEdge(x => (x === c.id ? null : x)); }
      if (c.type === "remove") act.removeEdge(c.id);
    });
  };

  /* ── typed connections: a hard rule, refused with a reason ── */
  const why = (c: Connection) => connectionError(c, lookup, doc.edges, reconnecting.current ?? undefined);
  const isValidConnection = (c: Connection) => { const w = why(c); HOT.reject = w; return !w; };
  const showBubble = (x: number, y: number, text: string) => {
    setBubble({ x, y, why: text });
    if (bubbleTimer.current) clearTimeout(bubbleTimer.current);
    bubbleTimer.current = setTimeout(() => setBubble(null), 2400);
  };
  const onConnectStart = (_: unknown, p: OnConnectStartParams) => {
    const h = parseHandle(p.handleId);
    const port = p.nodeId && h ? lookup(p.nodeId, h.direction, h.index) : undefined;
    HOT.kind = port ? port.kind : null; HOT.reject = null;
    pendingFrom.current = p.handleType === "source" && p.nodeId && h ? { node: p.nodeId, port: h.index } : null;
    setHotKind(HOT.kind); setQuick(null); setMenu(null);
  };
  const portRefs = (c: Connection) => {
    const sh = parseHandle(c.sourceHandle), th = parseHandle(c.targetHandle);
    if (!c.source || !c.target || !sh || !th) return null;
    return { from: { node: c.source, port: sh.index }, to: { node: c.target, port: th.index } };
  };
  const onConnect = (c: Connection) => {
    const r = portRefs(c); const out = r ? lookup(r.from.node, "out", r.from.port) : undefined;
    if (r && out && !why(c)) act.connect(r.from, r.to, out.kind);
    HOT.reject = null;
  };
  const onConnectEnd = (e: MouseEvent | TouchEvent) => {
    const pt = "changedTouches" in e ? e.changedTouches[0] : e;
    const t = e.target as Element | null;
    const sp = local(pt.clientX, pt.clientY);
    if (t?.classList?.contains("react-flow__pane") && pendingFrom.current) {
      const w = rf.screenToFlowPosition({ x: pt.clientX, y: pt.clientY });
      setQuick({ sx: sp.x, sy: sp.y, wx: w.x, wy: w.y - 20, from: pendingFrom.current, kind: HOT.kind, q: "" });
    } else if (t?.closest?.(".react-flow__handle") && HOT.reject) {
      showBubble(sp.x, sp.y, HOT.reject);
    }
    HOT.kind = null; HOT.reject = null; setHotKind(null); pendingFrom.current = null;
  };
  const onEdgeUpdateStart = (_: unknown, ed: Edge) => {
    reconnecting.current = ed.id;
    const h = parseHandle(ed.sourceHandle);
    const port = h ? lookup(ed.source, "out", h.index) : undefined;
    HOT.kind = port ? port.kind : null; setHotKind(HOT.kind);
  };
  const onEdgeUpdate = (old: Edge, c: Connection) => {
    const w = why(c);
    if (w) { HOT.reject = w; return; }
    const r = portRefs(c); const out = r ? lookup(r.from.node, "out", r.from.port) : undefined;
    if (r && out) act.connect(r.from, r.to, out.kind, old.id);
  };
  const onEdgeUpdateEnd = (e: MouseEvent | TouchEvent) => {
    const pt = "changedTouches" in e ? e.changedTouches[0] : e;
    if (HOT.reject) { const sp = local(pt.clientX, pt.clientY); showBubble(sp.x, sp.y, HOT.reject); }
    reconnecting.current = null; HOT.kind = null; HOT.reject = null; setHotKind(null);
  };

  /* ── quick add, menus, palette drop ── */
  const openQuick = (sx?: number, sy?: number) => {
    const r = size();
    const px = sx ?? r.width / 2, py = sy ?? r.height / 2;
    const w = rf.screenToFlowPosition({ x: r.left + px, y: r.top + py });
    setQuick({ sx: px, sy: sy == null ? py - 60 : py, wx: w.x - NODE_W / 2, wy: w.y - 24, q: "" });
  };
  const onDoubleClick = (e: ReactMouseEvent) => {
    if ((e.target as Element).classList.contains("react-flow__pane")) { const p = local(e.clientX, e.clientY); openQuick(p.x, p.y); }
  };
  const ctx = (e: ReactMouseEvent, extra: Partial<MenuState>) => { e.preventDefault(); const p = local(e.clientX, e.clientY); setMenu({ ...p, ...extra }); };
  const onDragOver = (e: DragEvent) => {
    if (!e.dataTransfer.types.includes("text/gc-node")) return;
    e.preventDefault(); e.dataTransfer.dropEffect = "copy";
    const w = rf.screenToFlowPosition({ x: e.clientX, y: e.clientY });
    setGhost({ x: snapG(w.x - NODE_W / 2), y: snapG(w.y - 24) });
  };
  const onDrop = (e: DragEvent) => {
    e.preventDefault(); setGhost(null);
    const type = e.dataTransfer.getData("text/gc-node"); if (!type) return;
    const w = rf.screenToFlowPosition({ x: e.clientX, y: e.clientY });
    act.addNode(type, snapG(w.x - NODE_W / 2), snapG(w.y - 24));
  };
  const allTypes: (NodeType & { group: string })[] = useMemo(() => s.groups.flatMap(g => g.items.map(i => ({ ...i, group: g.group }))), [s.groups]);
  const quickItems = quick ? allTypes
    .filter(i => !quick.kind || (i.in ?? []).some(p => p[0] === quick.kind))
    .filter(i => !quick.q || (i.title + " " + i.group + " " + (i.subtitle ?? "")).toLowerCase().includes(quick.q.toLowerCase())).slice(0, 9) : [];
  const pick = (i: NodeType) => {
    if (!quick) return;
    act.addNode(i.type, snapG(quick.wx), snapG(quick.wy), quick.from && quick.kind ? { from: quick.from, kind: quick.kind } : undefined);
    setQuick(null);
  };

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement;
      if (/input|textarea|select/i.test(t.tagName)) return;
      const mod = e.metaKey || e.ctrlKey;
      if (mod && e.key.toLowerCase() === "d" && sel.length) { e.preventDefault(); act.duplicate(sel); }
      else if (mod && e.key.toLowerCase() === "a") { e.preventDefault(); setSel(doc.nodes.map(n => n.id)); }
      else if (!mod && e.key.toLowerCase() === "v") setTool("select");
      else if (!mod && e.key.toLowerCase() === "h") setTool("pan");
      else if (e.shiftKey && (e.key === "!" || e.code === "Digit1")) { e.preventDefault(); fitAll(); }
      else if (e.key === "Escape") { setQuick(null); setMenu(null); }
    };
    window.addEventListener("keydown", down);
    return () => window.removeEventListener("keydown", down);
  }, [sel, doc.nodes, fitAll, act]);

  const menuNode = menu?.node ? doc.nodes.find(n => n.id === menu.node) : undefined;
  const targets = menu?.node ? (sel.length > 1 && sel.includes(menu.node) ? sel : [menu.node]) : [];
  const menuItems: [string, string, string | null, () => void][] = !menu ? [] : menu.node ? [
    ["tune", "Configure", null, () => { act.select(menu.node!); act.setModal("config"); }],
    ["content_copy", targets.length > 1 ? "Duplicate " + targets.length + " nodes" : "Duplicate", "⌘D", () => act.duplicate(targets)],
    [menuNode?.enabled ? "pause_circle" : "play_circle", menuNode?.enabled ? "Bypass node" : "Enable node", null, () => targets.forEach(id => act.toggleNode(id))],
    ["link_off", "Disconnect all", null, () => targets.forEach(id => act.clearLinks(id))],
    ["terminal", "View logs", null, () => act.setView("diag")],
    ["delete", targets.length > 1 ? "Delete " + targets.length + " nodes" : "Delete", "⌫", () => { act.removeNodes(targets); setSel([]); }],
  ] : menu.edge ? [["link_off", "Delete cable", "⌫", () => act.removeEdge(menu.edge!)]] : [
    ["add", "Add node here", "Dbl-click", () => openQuick(menu.x, menu.y)],
    ["fit_screen", "Zoom to fit", "⇧1", fitAll],
    ["select_all", "Select all", "⌘A", () => setSel(doc.nodes.map(n => n.id))],
    ["restart_alt", "Revert to saved", "R", act.revert],
  ];
  const valid = hasGadgetSink(doc, catalog);
  const vp = rf.getViewport();

  return (
    <div ref={wrapRef} onDragOver={onDragOver} onDragLeave={() => setGhost(null)} onDrop={onDrop} onDoubleClick={onDoubleClick} data-testid="canvas"
      style={{ flex: 1, minWidth: 0, minHeight: 0, position: "relative", background: "var(--surface-canvas)" }}>
      <ReactFlow nodes={rfNodes} edges={rfEdges} nodeTypes={nodeTypes} edgeTypes={edgeTypes} connectionLineComponent={ConnLine}
        onNodesChange={onNodesChange} onEdgesChange={onEdgesChange}
        onConnectStart={onConnectStart} onConnect={onConnect} onConnectEnd={onConnectEnd} isValidConnection={isValidConnection} connectionRadius={26}
        edgesUpdatable onEdgeUpdate={onEdgeUpdate} onEdgeUpdateStart={onEdgeUpdateStart} onEdgeUpdateEnd={onEdgeUpdateEnd}
        onNodeContextMenu={(e, n) => { if (!sel.includes(n.id)) { setSel([n.id]); act.select(n.id); } ctx(e, { node: n.id }); }}
        onEdgeContextMenu={(e, ed) => ctx(e, { edge: ed.id })} onPaneContextMenu={e => ctx(e, {})}
        onPaneClick={() => { setMenu(null); setQuick(null); setSelEdge(null); setSel([]); act.select(null); }}
        onNodeDragStart={() => { setMenu(null); setQuick(null); }}
        snapToGrid snapGrid={[GRID, GRID]} minZoom={0.25} maxZoom={2.5} zoomOnDoubleClick={false}
        panOnScroll zoomOnScroll={false} zoomOnPinch panActivationKeyCode="Space" panOnDrag={tool === "pan" ? true : [1, 2]}
        selectionOnDrag={tool === "select"} selectionMode={SelectionMode.Partial} deleteKeyCode={["Backspace", "Delete"]} multiSelectionKeyCode={["Meta", "Control", "Shift"]}
        proOptions={{ hideAttribution: true }} style={{ background: "transparent" }}>
        <Background variant={BackgroundVariant.Dots} gap={GRID} size={1.1} color="var(--grid-dot)" />
        {showMap && doc.nodes.length ? (
          <MiniMap pannable zoomable nodeColor={n => sigVar((n.data as GcNodeData).view.cls)} nodeStrokeWidth={0} nodeBorderRadius={2} maskColor="var(--surface-hover)"
            style={{ width: 168, height: 104, margin: 16, marginRight: s.selected ? 320 : 16, ...panel, overflow: "hidden" }} />
        ) : null}
        <Chrome studio={studio} tool={tool} setTool={setTool} showMap={showMap} setShowMap={setShowMap} openQuick={() => openQuick()} fitAll={fitAll} valid={valid} />
      </ReactFlow>

      {ghost ? <div style={{ position: "absolute", left: ghost.x * vp.zoom + vp.x, top: ghost.y * vp.zoom + vp.y, width: NODE_W * vp.zoom, height: 64 * vp.zoom, borderRadius: "var(--r-md)", border: "1.5px dashed var(--brand)", background: "var(--surface-selected)", pointerEvents: "none" }} /> : null}

      {!doc.nodes.length ? (
        <div style={{ position: "absolute", inset: 0, display: "grid", placeItems: "center", pointerEvents: "none" }}>
          <div style={{ textAlign: "center", display: "flex", flexDirection: "column", alignItems: "center", gap: 10 }}>
            <span style={{ width: 44, height: 44, borderRadius: "var(--r-lg)", display: "grid", placeItems: "center", background: "var(--surface-selected)" }}><Icon name="account_tree" size="md" style={{ color: "var(--brand)" }} /></span>
            <p style={{ font: "var(--type-heading)", color: "var(--text-primary)" }}>Empty pipeline</p>
            <p style={{ font: "var(--type-body)", color: "var(--text-muted)", maxWidth: 300 }}>Drag a node in from the palette, or double-click anywhere to add one.</p>
          </div>
        </div>
      ) : null}

      {sel.length > 1 ? (
        <div data-ui style={{ position: "absolute", left: "50%", top: 16, transform: "translateX(-50%)", display: "flex", alignItems: "center", gap: 4, padding: "4px 4px 4px 12px", ...panel, zIndex: 20 }}>
          <span style={{ font: "var(--type-ui)", color: "var(--text-body)" }}>{sel.length} selected</span>
          <span style={{ width: 1, height: 18, background: "var(--line-hairline)", margin: "0 4px" }} />
          <Tooltip label="Duplicate" shortcut="⌘D"><IconButton icon="content_copy" label="Duplicate" size="sm" onClick={() => act.duplicate(sel)} /></Tooltip>
          <Tooltip label="Bypass"><IconButton icon="pause_circle" label="Bypass" size="sm" onClick={() => sel.forEach(id => act.toggleNode(id))} /></Tooltip>
          <Tooltip label="Delete" shortcut="⌫"><IconButton icon="delete" label="Delete" size="sm" tone="var(--st-err)" onClick={() => { act.removeNodes(sel); setSel([]); }} /></Tooltip>
        </div>
      ) : null}

      {bubble ? (
        <div role="alert" data-testid="reject-bubble" style={{ position: "absolute", left: bubble.x + 14, top: bubble.y - 14, zIndex: 40, maxWidth: 260, display: "flex", gap: 8, alignItems: "center", padding: "8px 12px", background: "var(--surface-panel)", border: "1px solid var(--st-err)", borderRadius: "var(--r-lg)", font: "var(--type-ui)", color: "var(--st-err)" }}>
          <Icon name="block" size="xs" /><span>{bubble.why}</span>
        </div>
      ) : null}

      {quick ? (
        <div data-ui onDoubleClick={e => e.stopPropagation()}
          style={{ position: "absolute", left: Math.max(8, Math.min(quick.sx, size().width - 300)), top: Math.max(8, Math.min(quick.sy, size().height - 380)), width: 288, ...panel, zIndex: 60, overflow: "hidden" }}>
          <div style={{ padding: 8, borderBottom: "1px solid var(--line-hairline)" }}>
            <Input autoFocus icon="search" size="sm" placeholder={quick.kind ? "Add a " + quick.kind + " node…" : "Search nodes…"} value={quick.q} aria-label="Search nodes"
              onChange={e => setQuick({ ...quick, q: e.target.value })} onKeyDown={e => { if (e.key === "Enter" && quickItems[0]) pick(quickItems[0]); if (e.key === "Escape") setQuick(null); }} />
          </div>
          <div style={{ maxHeight: 300, overflow: "auto", padding: 6 }} className="gc-scroll">
            {quickItems.length ? quickItems.map((i, idx) => (
              <button type="button" key={i.type} className="gc-row" onClick={() => pick(i)} data-selected={idx === 0 && quick.q ? "true" : "false"}
                style={{ display: "flex", alignItems: "center", gap: 10, width: "100%", padding: "7px 8px", background: "transparent", border: 0, borderRadius: "var(--r-sm)", cursor: "pointer", textAlign: "left" }}>
                <span style={{ width: 26, height: 26, borderRadius: "var(--r-md)", display: "grid", placeItems: "center", background: sigTint(i.cls) }}><Icon name={i.icon} size="sm" style={{ color: sigVar(i.cls) }} /></span>
                <span style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", gap: 1 }}>
                  <span style={{ font: "var(--fw-medium) var(--fs-body)/1.2 var(--font-sans)", color: "var(--text-primary)" }}>{i.title}</span>
                  <span style={{ font: "var(--fw-regular) var(--fs-meta)/1.2 var(--font-sans)", color: "var(--text-faint)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{i.group} · {i.subtitle || i.url || i.dev}</span>
                </span>
                {quick.kind ? <Icon name="link" size="xs" style={{ color: "var(--text-disabled)" }} /> : null}
              </button>
            )) : <p style={{ padding: 16, font: "var(--type-ui)", color: "var(--text-faint)", textAlign: "center" }}>No matching nodes</p>}
          </div>
          {quick.kind ? <div style={{ padding: "8px 12px", borderTop: "1px solid var(--line-hairline)", font: "var(--type-meta)", color: "var(--text-faint)" }}>Picks a node with a {quick.kind} input and patches the cable in.</div> : null}
        </div>
      ) : null}

      {menu ? (
        <div data-ui role="menu" style={{ position: "absolute", left: Math.min(menu.x, size().width - 216), top: Math.max(8, Math.min(menu.y, size().height - 36 * menuItems.length - 24)), width: 208, padding: 6, ...panel, zIndex: 70 }}>
          {menuItems.map(([icon, label, key, fn], i) => {
            const danger = /delete/i.test(label);
            return (
              <button type="button" role="menuitem" key={i} className="gc-menuitem" onClick={() => { fn(); setMenu(null); }}
                style={{ display: "flex", alignItems: "center", gap: 10, width: "100%", height: 32, padding: "0 8px", background: "transparent", border: 0, borderRadius: "var(--r-sm)", cursor: "pointer", font: "var(--type-ui)", color: danger ? "var(--st-err)" : "var(--text-body)", textAlign: "left" }}>
                <Icon name={icon} size="sm" style={{ color: danger ? "var(--st-err)" : "var(--text-faint)" }} /><span style={{ flex: 1 }}>{label}</span>
                {key ? <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>{key}</span> : null}
              </button>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}

export function Canvas({ studio }: { studio: Studio }) {
  return <ReactFlowProvider><CanvasInner studio={studio} /></ReactFlowProvider>;
}
