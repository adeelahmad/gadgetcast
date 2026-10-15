import { describe, expect, it, vi } from "vitest";
import { ApiError } from "./backend";
import { createHttpBackend, type FetchLike } from "./client";
import { backoffMs, openEventStream, parseEvent, type SocketLike } from "./events";
import { createMockBackend } from "./mock";
import { selectBackend } from "./select";

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

describe("selectBackend: mock fallback selection", () => {
  it("uses the live client when /api/health answers ok", async () => {
    const f = vi.fn<FetchLike>(async () => json({ ok: true, version: "0.1.0" }));
    const { backend, reason } = await selectBackend({ fetch: f });
    expect(backend.mode).toBe("live");
    expect(backend.version).toBe("0.1.0");
    expect(reason).toBeUndefined();
    expect(f.mock.calls[0][0]).toBe("/api/health");
  });

  it("falls back to the demo when the request fails", async () => {
    const { backend, reason } = await selectBackend({ fetch: async () => { throw new TypeError("Failed to fetch"); } });
    expect(backend.mode).toBe("demo");
    expect(reason).toMatch(/health/);
  });

  it("falls back on a non-2xx, a non-JSON body, or ok:false", async () => {
    for (const res of [json({ error: "boom" }, 502), new Response("<html>", { status: 200 }), json({ ok: false, version: "0.1.0" })]) {
      const { backend } = await selectBackend({ fetch: async () => res });
      expect(backend.mode).toBe("demo");
    }
  });

  it("falls back when the Pi hangs past the timeout", async () => {
    const hang: FetchLike = (_u, init) => new Promise((_, rej) => init?.signal?.addEventListener("abort", () => rej(new DOMException("aborted", "AbortError"))));
    const { backend } = await selectBackend({ fetch: hang, timeoutMs: 20 });
    expect(backend.mode).toBe("demo");
  });

  it("honours a forced demo without probing", async () => {
    const f = vi.fn<FetchLike>();
    const { backend } = await selectBackend({ fetch: f, forceDemo: true });
    expect(backend.mode).toBe("demo");
    expect(f).not.toHaveBeenCalled();
  });
});

describe("HTTP client", () => {
  it("maps every endpoint to the contract's method + path", async () => {
    const calls: string[] = [];
    const f: FetchLike = async (url, init) => { calls.push((init?.method ?? "GET") + " " + url); return url.endsWith("/api/pipelines/p1") && init?.method === "DELETE" ? new Response(null, { status: 204 }) : json({}); };
    const b = createHttpBackend({ fetch: f });
    await b.listNodes(); await b.listPipelines(); await b.getPipeline("p1");
    await b.savePipeline({ id: "p1", name: "P", nodes: [], edges: [] });
    await b.deletePipeline("p1"); await b.runPipeline("p1"); await b.stopPipeline(); await b.getStatus(); await b.getDevices();
    expect(calls).toEqual([
      "GET /api/nodes", "GET /api/pipelines", "GET /api/pipelines/p1", "PUT /api/pipelines/p1", "DELETE /api/pipelines/p1",
      "POST /api/pipelines/p1/run", "POST /api/pipelines/stop", "GET /api/status", "GET /api/devices",
    ]);
  });

  it("surfaces the server's {error} text and refuses ids outside the contract pattern", async () => {
    const b = createHttpBackend({ fetch: async () => json({ error: "id must match path" }, 400) });
    await expect(b.getStatus()).rejects.toMatchObject({ status: 400, message: "id must match path" });
    await expect(b.getPipeline("../etc")).rejects.toBeInstanceOf(ApiError);
  });
});

describe("events stream", () => {
  it("parses only well-formed contract frames", () => {
    expect(parseEvent('{"type":"log","ts":"t","data":{"level":"info","subsystem":"server","msg":"x"}}')?.type).toBe("log");
    expect(parseEvent('{"type":"nope","ts":"t","data":{}}')).toBeNull();
    expect(parseEvent("not json")).toBeNull();
  });

  it("reconnects with backoff and stops when closed", () => {
    const sockets: SocketLike[] = [];
    const pending: (() => void)[] = [];
    const states: string[] = [];
    const factory = () => { const s: SocketLike = { onopen: null, onclose: null, onerror: null, onmessage: null, close: vi.fn() }; sockets.push(s); return s; };
    const got: string[] = [];
    const stop = openEventStream("ws://pi/api/events", ev => got.push(ev.type), s => states.push(s), factory,
      { set: fn => { pending.push(fn); return pending.length; }, clear: () => {} });
    sockets[0].onopen?.({});
    sockets[0].onmessage?.({ data: '{"type":"metrics","ts":"t","data":{"cpuPct":1}}' });
    sockets[0].onclose?.({});
    expect(pending).toHaveLength(1);
    pending[0]();
    expect(sockets).toHaveLength(2);
    expect(got).toEqual(["metrics"]);
    expect(states).toEqual(["connecting", "open", "reconnecting", "reconnecting"]);
    stop();
    expect(sockets[1].close).toHaveBeenCalled();
    expect(backoffMs(0)).toBe(500);
    expect(backoffMs(10)).toBe(15000);
  });
});

describe("mock backend honours the contract", () => {
  it("saves, runs, stops and deletes", async () => {
    const b = createMockBackend({ latencyMs: 0 });
    const list = await b.listPipelines();
    expect(list.length).toBeGreaterThan(0);
    const saved = await b.savePipeline({ ...list[0], id: "copy", name: "Copy", isDefault: true });
    expect(saved.updatedAt).toBeTruthy();
    expect((await b.listPipelines()).filter(p => p.isDefault).map(p => p.id)).toEqual(["copy"]);
    expect((await b.runPipeline("copy")).running).toBe("copy");
    await expect(b.deletePipeline("copy")).rejects.toMatchObject({ status: 409 });
    expect((await b.stopPipeline()).running).toBeNull();
    await b.deletePipeline("copy");
    await expect(b.getPipeline("copy")).rejects.toMatchObject({ status: 404 });
    await expect(b.savePipeline({ id: "bad id!", name: "x", nodes: [], edges: [] })).rejects.toMatchObject({ status: 400 });
  });
});
