import type { NodeGroup, NodeType, PortKind } from "../api/types";

/* The Studio's internal graph model. It differs from the wire Pipeline in two ways: edges address ports
   as {node, port} pairs (what the canvas and validation want), and ports are resolved from the catalogue
   rather than stored. mapping.ts converts in both directions. */

export interface PortRef { node: string; port: number }
export interface GraphNode {
  id: string;
  type: string;
  x: number;
  y: number;
  /** User-facing name; falls back to the catalogue title when absent. */
  title?: string;
  enabled: boolean;
  config: Record<string, unknown>;
}
export interface GraphEdge { id: string; from: PortRef; to: PortRef; kind: PortKind }
export interface GraphDoc {
  /** Server id; null until first saved. */
  id: string | null;
  name: string;
  description?: string;
  isDefault?: boolean;
  updatedAt?: string;
  nodes: GraphNode[];
  edges: GraphEdge[];
}
export interface ResolvedPort { kind: PortKind; label: string }

export type Catalog = Map<string, NodeType & { group: string }>;

export function indexCatalog(groups: NodeGroup[]): Catalog {
  const m: Catalog = new Map();
  groups.forEach(g => g.items.forEach(i => m.set(i.type, { ...i, group: g.group })));
  return m;
}

/**
 * Ports for a node. Known types read the catalogue; unknown types (a newer server, an imported file)
 * reconstruct ports from the cables that reference them so the graph still renders and round-trips.
 */
export function portsOf(node: GraphNode, catalog: Catalog, edges: GraphEdge[]): { inputs: ResolvedPort[]; outputs: ResolvedPort[] } {
  const t = catalog.get(node.type);
  if (t) {
    return {
      inputs: (t.in ?? []).map(([kind, label]) => ({ kind, label })),
      outputs: (t.out ?? []).map(([kind, label]) => ({ kind, label })),
    };
  }
  const inputs: ResolvedPort[] = [], outputs: ResolvedPort[] = [];
  edges.forEach(e => {
    if (e.to.node === node.id) inputs[e.to.port] = { kind: e.kind, label: "in " + e.to.port };
    if (e.from.node === node.id) outputs[e.from.port] = { kind: e.kind, label: "out " + e.from.port };
  });
  const fill = (arr: ResolvedPort[]) => Array.from({ length: arr.length }, (_, i) => arr[i] ?? { kind: "data" as PortKind, label: "—" });
  return { inputs: fill(inputs), outputs: fill(outputs) };
}

export function nodeTitle(node: GraphNode, catalog: Catalog): string {
  return node.title || catalog.get(node.type)?.title || node.type;
}

/** Fresh unique id for a node of `type`. */
export function newNodeId(type: string, taken: Iterable<string>): string {
  const set = new Set(taken);
  const stem = type.replace(/[^a-zA-Z0-9_-]/g, "_").slice(0, 48) || "node";
  let i = 1;
  while (set.has(stem + "_" + i)) i++;
  return stem + "_" + i;
}

export function newEdgeId(taken: Iterable<string>): string {
  const set = new Set(taken);
  let i = 1;
  while (set.has("e" + i)) i++;
  return "e" + i;
}

/** Pipeline id from a human name: lower-case slug, unique against `taken`, matching ^[a-zA-Z0-9_-]{1,64}$. */
export function slugId(name: string, taken: Iterable<string>): string {
  const set = new Set(taken);
  const stem = name.toLowerCase().normalize("NFKD").replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 56) || "pipeline";
  if (!set.has(stem)) return stem;
  let i = 2;
  while (set.has(stem + "-" + i)) i++;
  return stem + "-" + i;
}

/** A graph needs at least one USB gadget sink and one cable to be worth running. */
export function hasGadgetSink(doc: GraphDoc, catalog: Catalog): boolean {
  return doc.nodes.some(n => catalog.get(n.type)?.cls === "sink") && doc.edges.length > 0;
}
