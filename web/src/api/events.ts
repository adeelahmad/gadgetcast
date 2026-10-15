import type { EventsState } from "./backend";
import type { ServerEvent } from "./types";

/** Minimal WebSocket surface, so tests can inject a fake. */
export interface SocketLike {
  onopen: ((ev: unknown) => void) | null;
  onclose: ((ev: unknown) => void) | null;
  onerror: ((ev: unknown) => void) | null;
  onmessage: ((ev: { data: unknown }) => void) | null;
  close(): void;
}
export type SocketFactory = (url: string) => SocketLike;

/** Parse one frame; returns null for anything that is not a well-formed contract event. */
export function parseEvent(raw: unknown): ServerEvent | null {
  if (typeof raw !== "string") return null;
  let v: unknown;
  try { v = JSON.parse(raw); } catch { return null; }
  if (!v || typeof v !== "object") return null;
  const o = v as { type?: unknown; ts?: unknown; data?: unknown };
  if (typeof o.ts !== "string" || !o.data || typeof o.data !== "object") return null;
  if (o.type === "metrics" || o.type === "status" || o.type === "log") return o as ServerEvent;
  return null;
}

/** Backoff for reconnect attempt n (0-based): 0.5 s, 1 s, 2 s, 4 s … capped at 15 s. */
export function backoffMs(attempt: number): number {
  return Math.min(15000, 500 * 2 ** Math.max(0, attempt));
}

/** ws(s)://host/path for the page's origin. */
export function wsUrl(path: string, loc: { protocol: string; host: string } = window.location): string {
  return (loc.protocol === "https:" ? "wss://" : "ws://") + loc.host + path;
}

/**
 * A self-healing subscription to GET /api/events. Reconnects with exponential backoff until closed;
 * a successful open resets the backoff.
 */
export function openEventStream(
  url: string,
  onEvent: (ev: ServerEvent) => void,
  onState: (s: EventsState) => void = () => {},
  factory: SocketFactory = u => new WebSocket(u) as unknown as SocketLike,
  timers: { set: (fn: () => void, ms: number) => unknown; clear: (h: unknown) => void } = {
    set: (fn, ms) => setTimeout(fn, ms),
    clear: h => clearTimeout(h as ReturnType<typeof setTimeout>),
  },
): () => void {
  let closed = false;
  let attempt = 0;
  let sock: SocketLike | null = null;
  let timer: unknown = null;

  const connect = () => {
    if (closed) return;
    onState(attempt === 0 ? "connecting" : "reconnecting");
    let s: SocketLike;
    try { s = factory(url); } catch { schedule(); return; }
    sock = s;
    s.onopen = () => { attempt = 0; onState("open"); };
    s.onmessage = m => { const ev = parseEvent(m.data); if (ev) onEvent(ev); };
    s.onerror = () => { /* onclose follows */ };
    s.onclose = () => { if (sock === s) { sock = null; schedule(); } };
  };
  const schedule = () => {
    if (closed) return;
    onState("reconnecting");
    const delay = backoffMs(attempt++);
    timer = timers.set(connect, delay);
  };

  connect();
  return () => {
    closed = true;
    if (timer != null) timers.clear(timer);
    const s = sock; sock = null;
    if (s) { s.onclose = null; s.onmessage = null; s.close(); }
  };
}
