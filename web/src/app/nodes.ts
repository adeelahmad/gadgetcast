import type { PortKind } from "../api/types";
import { portsOf, type Catalog, type GraphEdge, type GraphNode, type ResolvedPort } from "../graph/model";

/** Signal family used to tint a node class's icon tile (bluetooth has its own hue). */
export type Sig = "video" | "audio" | "ai" | "data" | "bt";
export function sigOf(cls: string | undefined): Sig {
  return ({ ai: "ai", audio: "audio", logic: "data", bluetooth: "bt" } as Record<string, Sig>)[cls ?? ""] ?? "video";
}
export const sigVar = (cls: string | undefined) => "var(--sig-" + sigOf(cls) + ")";
export const sigTint = (cls: string | undefined, pct = 13) => "color-mix(in srgb, " + sigVar(cls) + " " + pct + "%, transparent)";

export interface NodeView {
  title: string;
  subtitle: string;
  icon: string;
  cls: string;
  group: string;
  category: string;
  media: string;
  gst: string;
  known: boolean;
  inputs: ResolvedPort[];
  outputs: ResolvedPort[];
}

export function cfgStr(n: GraphNode, key: string, fallback = ""): string {
  const v = n.config[key];
  return typeof v === "string" ? v : fallback;
}
export function cfgNum(n: GraphNode, key: string, fallback: number): number {
  const v = n.config[key];
  return typeof v === "number" && Number.isFinite(v) ? v : fallback;
}

export function viewOf(n: GraphNode, catalog: Catalog, edges: GraphEdge[]): NodeView {
  const t = catalog.get(n.type);
  const { inputs, outputs } = portsOf(n, catalog, edges);
  return {
    title: n.title || t?.title || n.type,
    subtitle: cfgStr(n, "url") || t?.subtitle || t?.url || t?.dev || (t ? "" : "unknown node type · " + n.type),
    icon: t?.icon ?? "help",
    cls: t?.cls ?? "logic",
    group: t?.group ?? "Unknown",
    category: t?.category ?? "unknown",
    media: t?.media ?? "data",
    gst: t?.gst ?? "",
    known: !!t,
    inputs, outputs,
  };
}

/** Node types whose config carries a source / destination URL or path. */
export const URL_TYPES = new Set(["rtsp", "rtmp", "rtp", "srt", "hls", "url", "file", "s3", "nfs", "cifs", "nvme", "rtmp_out", "record"]);
export const URL_RULE = /^(rtsp|rtsps|rtmp|rtmps|udp|srt|https?|s3|nfs|smb):\/\/|^\//;

export const KIND_LABEL: Record<PortKind, string> = { video: "video", audio: "audio", ai: "ai", data: "data" };
