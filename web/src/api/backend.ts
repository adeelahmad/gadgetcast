import type { DevicesResponse, NodeGroup, Pipeline, ServerEvent, Status } from "./types";

export type BackendMode = "live" | "demo";
/** State of the push channel (GET /api/events). */
export type EventsState = "connecting" | "open" | "reconnecting";

/**
 * Everything the Studio needs from the Pi. Two implementations exist: the HTTP/WebSocket client
 * (client.ts) and the bundled offline demo (mock.ts). Nothing outside src/api knows which one it holds,
 * apart from `mode` (for the header badge) and the optional simulation hook below.
 */
export interface Backend {
  readonly mode: BackendMode;
  /** Server version from /api/health, or "demo". */
  readonly version: string;

  listNodes(): Promise<NodeGroup[]>;
  listPipelines(): Promise<Pipeline[]>;
  getPipeline(id: string): Promise<Pipeline>;
  /** PUT: the server sets updatedAt and returns the saved pipeline. */
  savePipeline(p: Pipeline): Promise<Pipeline>;
  deletePipeline(id: string): Promise<void>;
  runPipeline(id: string): Promise<Status>;
  stopPipeline(): Promise<Status>;
  getStatus(): Promise<Status>;
  getDevices(): Promise<DevicesResponse>;

  /** Subscribe to server push. Returns an unsubscribe function. */
  subscribe(onEvent: (ev: ServerEvent) => void, onState?: (s: EventsState) => void): () => void;

  /**
   * Demo only: synthetic audio levels for the meters (0..1 per channel). The v0 API reports no levels,
   * so the live backend leaves this undefined and meters stay at rest.
   */
  simulatedLevels?: () => { levels: number[]; peaks: number[] };
}

export class ApiError extends Error {
  constructor(readonly status: number, message: string) {
    super(message);
    this.name = "ApiError";
  }
}
