import { ApiError, type Backend, type EventsState } from "./backend";
import { openEventStream, wsUrl, type SocketFactory } from "./events";
import { ID_PATTERN, type DevicesResponse, type Health, type NodeGroup, type Pipeline, type ServerEvent, type Status } from "./types";

export type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export interface HttpBackendOptions {
  /** Prefix for every request; "" means same origin (the Pi serves the SPA). */
  base?: string;
  fetch?: FetchLike;
  socketFactory?: SocketFactory;
  version?: string;
}

function checkId(id: string): string {
  if (!ID_PATTERN.test(id)) throw new ApiError(400, "Pipeline id must match ^[a-zA-Z0-9_-]{1,64}$");
  return encodeURIComponent(id);
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: unknown };
    if (body && typeof body.error === "string") return body.error;
  } catch { /* not JSON */ }
  return "HTTP " + res.status + (res.statusText ? " " + res.statusText : "");
}

/** Typed client for every endpoint in docs/api.md. */
export function createHttpBackend(opts: HttpBackendOptions = {}): Backend {
  const base = opts.base ?? "";
  const f: FetchLike = opts.fetch ?? ((i, init) => fetch(i, init));

  async function call<T>(method: string, path: string, body?: unknown): Promise<T> {
    const init: RequestInit = { method, headers: { Accept: "application/json" } };
    if (body !== undefined) {
      init.body = JSON.stringify(body);
      init.headers = { Accept: "application/json", "Content-Type": "application/json" };
    }
    let res: Response;
    try { res = await f(base + path, init); } catch (e) {
      throw new ApiError(0, "The Pi did not answer (" + (e instanceof Error ? e.message : "network error") + ")");
    }
    if (!res.ok) throw new ApiError(res.status, await readError(res));
    if (res.status === 204) return undefined as T;
    const text = await res.text();
    return (text ? JSON.parse(text) : undefined) as T;
  }

  return {
    mode: "live",
    version: opts.version ?? "unknown",
    listNodes: () => call<NodeGroup[]>("GET", "/api/nodes"),
    listPipelines: () => call<Pipeline[]>("GET", "/api/pipelines"),
    getPipeline: async id => call<Pipeline>("GET", "/api/pipelines/" + checkId(id)),
    savePipeline: async p => call<Pipeline>("PUT", "/api/pipelines/" + checkId(p.id), p),
    deletePipeline: async id => call<void>("DELETE", "/api/pipelines/" + checkId(id)),
    runPipeline: async id => call<Status>("POST", "/api/pipelines/" + checkId(id) + "/run"),
    stopPipeline: () => call<Status>("POST", "/api/pipelines/stop"),
    getStatus: () => call<Status>("GET", "/api/status"),
    getDevices: () => call<DevicesResponse>("GET", "/api/devices"),
    subscribe(onEvent: (ev: ServerEvent) => void, onState?: (s: EventsState) => void) {
      const url = base ? base.replace(/^http/, "ws") + "/api/events" : wsUrl("/api/events");
      return openEventStream(url, onEvent, onState, opts.socketFactory);
    },
  };
}

/** GET /api/health with a timeout. Resolves to the health body, or null when the Pi is unreachable or unhealthy. */
export async function probeHealth(f: FetchLike, base = "", timeoutMs = 2500): Promise<Health | null> {
  const ctl = typeof AbortController !== "undefined" ? new AbortController() : null;
  const t = setTimeout(() => ctl?.abort(), timeoutMs);
  try {
    const res = await f(base + "/api/health", { headers: { Accept: "application/json" }, signal: ctl?.signal });
    if (!res.ok) return null;
    const body = (await res.json()) as Partial<Health>;
    if (!body || body.ok !== true) return null;
    return { ok: true, version: typeof body.version === "string" ? body.version : "unknown" };
  } catch {
    return null;
  } finally {
    clearTimeout(t);
  }
}
