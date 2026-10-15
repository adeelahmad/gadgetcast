import { ApiError, type Backend, type EventsState } from "./backend";
import { CATALOG, DEVICES, LOG_SEEDS, seedPipelines } from "./mockData";
import { ID_PATTERN, type DevicesResponse, type Metrics, type Pipeline, type ServerEvent, type Status } from "./types";

export interface MockOptions {
  /** Artificial response delay, ms. */
  latencyMs?: number;
  random?: () => number;
  /** Event cadence, ms (metrics every 2 s per the contract). */
  metricsEveryMs?: number;
  logEveryMs?: number;
}

const clone = <T,>(v: T): T => (v === undefined ? v : (JSON.parse(JSON.stringify(v)) as T));

/**
 * The offline demo: an in-memory stand-in for the Pi that honours the same contract as the HTTP client,
 * including validation errors and the events stream. Used automatically when /api/health fails.
 */
export function createMockBackend(opts: MockOptions = {}): Backend {
  const rnd = opts.random ?? Math.random;
  const latency = opts.latencyMs ?? 60;
  const pipelines = new Map<string, Pipeline>(seedPipelines().map(p => [p.id, p]));
  const startedAt = Date.now() - 3 * 3600 * 1000;
  let running: string | null = null;
  let onAir: string | null = null;
  const listeners = new Set<(ev: ServerEvent) => void>();

  const delay = <T,>(v: T): Promise<T> => new Promise(res => setTimeout(() => res(clone(v)), latency));
  const fail = (status: number, msg: string) => new Promise<never>((_, rej) => setTimeout(() => rej(new ApiError(status, msg)), latency));
  const now = () => new Date().toISOString();

  const metrics = (): Metrics => ({
    cpuPct: Math.round((running ? 30 + rnd() * 30 : 4 + rnd() * 6) * 10) / 10,
    ramUsedMB: Math.round(running ? 610 + rnd() * 80 : 420 + rnd() * 20),
    ramTotalMB: 3796,
    tempC: Math.round((running ? 56 + rnd() * 9 : 48 + rnd() * 4) * 10) / 10,
    uptimeS: Math.round((Date.now() - startedAt) / 1000),
    loadAvg: [running ? 1.2 + rnd() : 0.2 + rnd() * 0.2, 0.9, 0.7].map(v => Math.round(v * 100) / 100),
  });
  const status = (): Status => ({ running, onAir, clients: DEVICES.clients.length, metrics: metrics() });
  const emit = (ev: ServerEvent) => listeners.forEach(l => l(ev));
  const log = (level: "info" | "warn" | "error", subsystem: "server" | "session" | "pipeline", msg: string) =>
    emit({ type: "log", ts: now(), data: { level, subsystem, msg } });

  const devices = (): DevicesResponse => ({
    devices: DEVICES.devices,
    clients: DEVICES.clients.map(c => ({ ...c, onAir: c.id === onAir })),
  });

  return {
    mode: "demo",
    version: "demo",
    listNodes: () => delay(CATALOG),
    listPipelines: () => delay(Array.from(pipelines.values())),
    getPipeline: id => {
      const p = pipelines.get(id);
      return p ? delay(p) : fail(404, "No pipeline “" + id + "”");
    },
    savePipeline: p => {
      if (!ID_PATTERN.test(p.id)) return fail(400, "Pipeline id must match ^[a-zA-Z0-9_-]{1,64}$");
      if (!p.name || !p.name.trim()) return fail(400, "Pipeline name is required");
      const saved: Pipeline = { ...clone(p), updatedAt: now() };
      if (saved.isDefault) pipelines.forEach(o => { if (o.id !== saved.id) o.isDefault = false; });
      pipelines.set(saved.id, saved);
      log("info", "server", "pipeline “" + saved.name + "” saved (" + saved.nodes.length + " nodes)");
      return delay(saved);
    },
    deletePipeline: id => {
      if (!pipelines.has(id)) return fail(404, "No pipeline “" + id + "”");
      if (running === id) return fail(409, "Stop the pipeline before deleting it");
      pipelines.delete(id);
      return delay(undefined);
    },
    runPipeline: id => {
      const p = pipelines.get(id);
      if (!p) return fail(404, "No pipeline “" + id + "”");
      running = id;
      onAir = DEVICES.clients[0]?.id ?? null;
      const s = status();
      setTimeout(() => {
        emit({ type: "status", ts: now(), data: s });
        log("info", "pipeline", "gst: “" + p.name + "” → PLAYING (simulated)");
        log("info", "session", "signal: " + (onAir ?? "client") + " is on air");
      }, 0);
      return delay(s);
    },
    stopPipeline: () => {
      running = null; onAir = null;
      const s = status();
      setTimeout(() => { emit({ type: "status", ts: now(), data: s }); log("info", "pipeline", "gst: pipeline → NULL, gadget released"); }, 0);
      return delay(s);
    },
    getStatus: () => delay(status()),
    getDevices: () => delay(devices()),
    subscribe(onEvent: (ev: ServerEvent) => void, onState?: (s: EventsState) => void) {
      listeners.add(onEvent);
      onState?.("open");
      const seed = LOG_SEEDS.slice(0, 6);
      seed.forEach(d => onEvent({ type: "log", ts: now(), data: d }));
      const m = setInterval(() => onEvent({ type: "metrics", ts: now(), data: metrics() }), opts.metricsEveryMs ?? 2000);
      const l = setInterval(() => {
        if (!running) return;
        onEvent({ type: "log", ts: now(), data: LOG_SEEDS[Math.floor(rnd() * LOG_SEEDS.length)] });
      }, opts.logEveryMs ?? 1600);
      return () => { listeners.delete(onEvent); clearInterval(m); clearInterval(l); };
    },
    simulatedLevels: () => {
      const hot = running ? 0.7 : 0;
      if (!hot) return { levels: [0, 0], peaks: [0, 0] };
      return {
        levels: [0.25 + rnd() * hot, 0.2 + rnd() * (hot - 0.02)],
        peaks: [0.55 + rnd() * 0.42, 0.5 + rnd() * 0.42],
      };
    },
  };
}
