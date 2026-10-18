import type { Catalog, GraphDoc, GraphEdge, ResolvedPort } from "./model";
import { portsOf } from "./model";

/** React Flow handle ids are "in-<index>" / "out-<index>". */
export function handleId(direction: "in" | "out", index: number): string { return direction + "-" + index; }
export function parseHandle(h: string | null | undefined): { direction: "in" | "out"; index: number } | null {
  if (!h) return null;
  const m = /^(in|out)-(\d+)$/.exec(h);
  return m ? { direction: m[1] as "in" | "out", index: Number(m[2]) } : null;
}

export interface Connection { source: string | null; sourceHandle: string | null; target: string | null; targetHandle: string | null }
export type PortLookup = (nodeId: string, direction: "in" | "out", index: number) => ResolvedPort | undefined;

/**
 * Why a cable may not be patched, or null when it may. A hard rule, not a warning:
 * same kind only, no self-loop, one cable per input, output → input only.
 * `ignoreEdgeId` lets an edge being re-routed vacate the input it currently holds.
 */
export function connectionError(c: Connection, portOf: PortLookup, edges: GraphEdge[], ignoreEdgeId?: string): string | null {
  if (!c.source || !c.target) return "Drop the cable on a port";
  const sh = parseHandle(c.sourceHandle), th = parseHandle(c.targetHandle);
  if (!sh || !th) return "Drop the cable on a port";
  if (sh.direction !== "out" || th.direction !== "in") return "Patch from an output to an input";
  if (c.source === c.target) return "A node cannot patch into itself";
  const out = portOf(c.source, "out", sh.index), inp = portOf(c.target, "in", th.index);
  if (!out || !inp) return "That port does not exist";
  if (out.kind !== inp.kind) return "Cannot link " + out.kind.toUpperCase() + " output to " + inp.kind.toUpperCase() + " input";
  if (edges.some(e => e.id !== ignoreEdgeId && e.to.node === c.target && e.to.port === th.index)) return "That input is already patched";
  return null;
}

export function lookupFor(doc: GraphDoc, catalog: Catalog): PortLookup {
  return (nodeId, direction, index) => {
    const n = doc.nodes.find(x => x.id === nodeId);
    if (!n) return undefined;
    const p = portsOf(n, catalog, doc.edges);
    return (direction === "in" ? p.inputs : p.outputs)[index];
  };
}

/** Check a whole graph (e.g. an imported file). Returns human-readable problems; empty means valid. */
export function graphProblems(doc: GraphDoc, catalog: Catalog): string[] {
  const problems: string[] = [];
  const lookup = lookupFor(doc, catalog);
  const ids = new Set<string>();
  doc.nodes.forEach(n => {
    if (ids.has(n.id)) problems.push("Duplicate node id " + n.id);
    ids.add(n.id);
    if (!catalog.has(n.type)) problems.push("Unknown node type “" + n.type + "” (" + n.id + ")");
  });
  doc.edges.forEach((e, i) => {
    const err = connectionError({ source: e.from.node, sourceHandle: handleId("out", e.from.port), target: e.to.node, targetHandle: handleId("in", e.to.port) }, lookup, doc.edges.slice(0, i));
    if (err) problems.push(e.id + ": " + err);
    else {
      const out = lookup(e.from.node, "out", e.from.port);
      if (out && out.kind !== e.kind) problems.push(e.id + ": cable kind " + e.kind + " does not match its " + out.kind + " port");
    }
  });
  return problems;
}

export interface PatchTarget { node: string; port: number; label: string }

/** Every input the given output may be patched to, for patching without a pointer (Inspector). */
export function patchTargets(doc: GraphDoc, catalog: Catalog, nodeId: string, outPort: number): PatchTarget[] {
  const lookup = lookupFor(doc, catalog), out: PatchTarget[] = [];
  for (const n of doc.nodes) {
    portsOf(n, catalog, doc.edges).inputs.forEach((p, port) => {
      const c = { source: nodeId, sourceHandle: handleId("out", outPort), target: n.id, targetHandle: handleId("in", port) };
      if (!connectionError(c, lookup, doc.edges)) out.push({ node: n.id, port, label: (catalog.get(n.type)?.title ?? n.type) + " (" + n.id + ") · " + p.label });
    });
  }
  return out;
}
