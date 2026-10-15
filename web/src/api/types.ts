/* Wire types for the Pi-side HTTP / WebSocket API (docs/api.md, v0). Keep these byte-for-byte with the
   contract: the Go server is built against the same document. */

export type PortKind = "video" | "audio" | "ai" | "data";
/** [kind, format label], e.g. ["video", "H.264"]. */
export type Port = [PortKind, string];

export interface NodeType {
  type: string;
  title: string;
  icon: string;
  cls: string;
  category: string;
  media?: "video" | "audio" | "both" | "ai" | "data";
  gst?: string;
  subtitle?: string;
  url?: string;
  dev?: string;
  in?: Port[];
  out?: Port[];
}
export interface NodeGroup { group: string; items: NodeType[] }

export interface PipelineNode {
  id: string;
  type: string;
  x: number;
  y: number;
  title?: string;
  enabled?: boolean;
  config?: Record<string, unknown>;
}
export interface PipelineEdge { id: string; from: string; fromPort: number; to: string; toPort: number; kind: PortKind }
export interface Pipeline {
  id: string;
  name: string;
  description?: string;
  isDefault?: boolean;
  nodes: PipelineNode[];
  edges: PipelineEdge[];
  updatedAt?: string;
}

export interface Metrics { cpuPct: number; ramUsedMB: number; ramTotalMB: number; tempC: number | null; uptimeS: number; loadAvg: number[] }
export interface Status { running: string | null; onAir: string | null; clients: number; metrics: Metrics }
export interface Device { id: string; kind: "video" | "audio"; name: string; path: string }
export interface Client { id: string; onAir: boolean }
export interface DevicesResponse { devices: Device[]; clients: Client[] }
export interface Health { ok: boolean; version: string }

export type LogEventLevel = "info" | "warn" | "error";
export type LogSubsystem = "server" | "session" | "pipeline";
export interface LogEventData { level: LogEventLevel; subsystem: LogSubsystem; msg: string }

/** One frame on GET /api/events. */
export type ServerEvent =
  | { type: "metrics"; ts: string; data: Metrics }
  | { type: "status"; ts: string; data: Status }
  | { type: "log"; ts: string; data: LogEventData };

/** `id` path segments must match this. */
export const ID_PATTERN = /^[a-zA-Z0-9_-]{1,64}$/;
