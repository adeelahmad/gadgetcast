import type { Pipeline, PipelineEdge, PipelineNode, PortKind } from "../api/types";
import type { GraphDoc, GraphEdge, GraphNode } from "./model";

const KINDS: ReadonlySet<string> = new Set<PortKind>(["video", "audio", "ai", "data"]);

/** Wire Pipeline → internal graph. Cables that reference a missing node or carry an unknown kind are dropped and counted. */
export function pipelineToGraph(p: Pipeline): { doc: GraphDoc; dropped: number } {
  const nodes: GraphNode[] = p.nodes.map((n: PipelineNode) => {
    const g: GraphNode = { id: n.id, type: n.type, x: n.x, y: n.y, enabled: n.enabled !== false, config: { ...(n.config ?? {}) } };
    if (n.title !== undefined) g.title = n.title;
    return g;
  });
  const ids = new Set(nodes.map(n => n.id));
  let dropped = 0;
  const edges: GraphEdge[] = [];
  p.edges.forEach((e: PipelineEdge) => {
    if (!ids.has(e.from) || !ids.has(e.to) || !KINDS.has(e.kind)) { dropped++; return; }
    edges.push({ id: e.id, from: { node: e.from, port: e.fromPort }, to: { node: e.to, port: e.toPort }, kind: e.kind });
  });
  const doc: GraphDoc = { id: p.id, name: p.name, nodes, edges };
  if (p.description !== undefined) doc.description = p.description;
  if (p.isDefault !== undefined) doc.isDefault = p.isDefault;
  if (p.updatedAt !== undefined) doc.updatedAt = p.updatedAt;
  return { doc, dropped };
}

/** Internal graph → wire Pipeline, ready for PUT /api/pipelines/{id}. Positions are rounded to whole px. */
export function graphToPipeline(doc: GraphDoc, id: string = doc.id ?? ""): Pipeline {
  const nodes: PipelineNode[] = doc.nodes.map(n => {
    const out: PipelineNode = { id: n.id, type: n.type, x: Math.round(n.x), y: Math.round(n.y), enabled: n.enabled, config: { ...n.config } };
    if (n.title !== undefined) out.title = n.title;
    return out;
  });
  const edges: PipelineEdge[] = doc.edges.map(e => ({ id: e.id, from: e.from.node, fromPort: e.from.port, to: e.to.node, toPort: e.to.port, kind: e.kind }));
  const p: Pipeline = { id, name: doc.name, nodes, edges };
  if (doc.description !== undefined) p.description = doc.description;
  if (doc.isDefault !== undefined) p.isDefault = doc.isDefault;
  if (doc.updatedAt !== undefined) p.updatedAt = doc.updatedAt;
  return p;
}

/**
 * Parse an exported JSON file. Accepts the wire Pipeline shape, plus the design-system prototype's
 * export ({name, nodes:[{id,type,x,y,title,...}], edges:[{from:{node,port},to:{node,port},kind}]}).
 */
export function parseImportedPipeline(raw: unknown): Pipeline {
  if (!raw || typeof raw !== "object") throw new Error("Not a pipeline file");
  const o = raw as Record<string, unknown>;
  if (!Array.isArray(o.nodes) || !Array.isArray(o.edges)) throw new Error("Pipeline JSON needs nodes[] and edges[]");
  const nodes: PipelineNode[] = (o.nodes as Record<string, unknown>[]).map((n, i) => {
    if (typeof n.type !== "string") throw new Error("Node " + i + " has no type");
    const node: PipelineNode = {
      id: typeof n.id === "string" ? n.id : "n" + (i + 1), type: n.type,
      x: typeof n.x === "number" ? n.x : 0, y: typeof n.y === "number" ? n.y : 0,
      enabled: n.enabled !== false,
      config: n.config && typeof n.config === "object" ? (n.config as Record<string, unknown>) : {},
    };
    if (typeof n.title === "string") node.title = n.title;
    return node;
  });
  const edges: PipelineEdge[] = (o.edges as Record<string, unknown>[]).map((e, i) => {
    const kind = e.kind as PortKind;
    if (!KINDS.has(kind)) throw new Error("Cable " + i + " has an unknown kind");
    const id = typeof e.id === "string" ? e.id : "e" + (i + 1);
    if (typeof e.from === "string" && typeof e.to === "string") {
      return { id, from: e.from, fromPort: Number(e.fromPort) || 0, to: e.to, toPort: Number(e.toPort) || 0, kind };
    }
    const f = e.from as { node?: unknown; port?: unknown } | undefined, t = e.to as { node?: unknown; port?: unknown } | undefined;
    if (!f || !t || typeof f.node !== "string" || typeof t.node !== "string") throw new Error("Cable " + i + " has no endpoints");
    return { id, from: f.node, fromPort: Number(f.port) || 0, to: t.node, toPort: Number(t.port) || 0, kind };
  });
  return {
    id: typeof o.id === "string" ? o.id : "",
    name: typeof o.name === "string" && o.name.trim() ? o.name : "Imported pipeline",
    description: typeof o.description === "string" ? o.description : undefined,
    nodes, edges,
  };
}
