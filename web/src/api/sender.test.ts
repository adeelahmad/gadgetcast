import { describe, expect, it } from "vitest";
import { newClientId, startMicSender, type SendState, type SenderDeps, type SenderPeer, type SenderSocket } from "./sender";

class FakeSocket implements SenderSocket {
  sent: { type: string; payload: Record<string, unknown> }[] = [];
  closed = false;
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  send(d: string) { this.sent.push(JSON.parse(d)); }
  close() { this.closed = true; }
  push(type: string, payload: unknown) { this.onmessage?.({ data: JSON.stringify({ type, payload }) }); }
}

class FakePeer implements SenderPeer {
  local: { sdp: string } | null = null;
  remote: string | null = null;
  closed = false;
  iceGatheringState = "new";
  onicegatheringstatechange: (() => void) | null = null;
  transceivers = 0;
  get localDescription() { return this.local; }
  addTransceiver() { this.transceivers++; return {}; }
  async createOffer() { return { type: "offer", sdp: "v=0 offer" }; }
  async setLocalDescription(d: { sdp?: string }) {
    this.local = { sdp: (d.sdp ?? "") + " +candidates" };
    setTimeout(() => { this.iceGatheringState = "complete"; this.onicegatheringstatechange?.(); }, 1);
  }
  async setRemoteDescription(d: { sdp: string }) { this.remote = d.sdp; }
  close() { this.closed = true; }
}

function fakeStream() {
  const track = { stopped: false, stop() { this.stopped = true; } };
  const video = { stopped: false, stop() { this.stopped = true; }, ended: null as null | (() => void), addEventListener(_: string, f: () => void) { this.ended = f; } };
  return { track, video, stream: { getTracks: () => [track, video] } as unknown as MediaStream };
}

function setup(opts: { micError?: Error } = {}) {
  const sock = new FakeSocket();
  const peer = new FakePeer();
  const mic = fakeStream();
  const states: [SendState, string | undefined][] = [];
  const deps: SenderDeps = {
    url: "wss://pi/ws",
    socket: () => sock,
    peer: () => peer,
    getMedia: () => (opts.micError ? Promise.reject(opts.micError) : Promise.resolve(mic.stream)),
    gatherTimeoutMs: 50,
  };
  const stop = startMicSender(deps, "1234", { onState: (s, d) => states.push([s, d]) }, "mic-abc");
  return { sock, peer, mic, states, stop };
}

const tick = () => new Promise(r => setTimeout(r, 10));
const last = (s: [SendState, string | undefined][]) => s[s.length - 1];

describe("startMicSender", () => {
  it("joins, pairs, offers, and goes on air", async () => {
    const { sock, peer, states, stop, mic } = setup();
    await tick(); await tick();
    sock.onopen?.();
    expect(sock.sent.map(f => f.type)).toEqual(["join", "submit-pin", "media-offer"]);
    expect(sock.sent[0].payload).toEqual({ client: "mic-abc" });
    expect(sock.sent[1].payload).toEqual({ pin: "1234" });
    expect(sock.sent[2].payload.sdp).toBe("v=0 offer +candidates");
    expect(peer.transceivers).toBe(2);

    sock.push("media-answer", { sdp: "v=0 answer" });
    await tick();
    expect(peer.remote).toBe("v=0 answer");
    expect(sock.sent[3]).toEqual({ type: "switch-source", payload: { target: "mic-abc" } });

    sock.push("on-air-state", { onair: "mic-abc", live: true });
    expect(last(states)[0]).toBe("on-air");
    sock.push("on-air-state", { onair: "mic-xyz", live: true });
    expect(last(states)).toEqual(["connecting", "Another device is on air."]);

    stop();
    expect(last(states)[0]).toBe("stopped");
    expect(sock.closed && peer.closed && mic.track.stopped && mic.video.stopped).toBe(true);
    stop(); // idempotent
    expect(states.filter(s => s[0] === "stopped").length).toBe(1);
  });

  it("reports a wrong PIN once and cleans up", async () => {
    const { sock, peer, states } = setup();
    await tick(); await tick();
    sock.onopen?.();
    sock.push("error", { reason: "pairing: incorrect pin" });
    sock.push("error", { reason: "pair with the PIN before sending media" });
    expect(states.filter(s => s[0] === "error")).toEqual([["error", "Wrong PIN."]]);
    expect(peer.closed).toBe(true);
  });

  it("explains lockout and lost connections", async () => {
    const a = setup();
    await tick(); await tick();
    a.sock.push("error", { reason: "pairing: client rejected" });
    expect(last(a.states)[1]).toMatch(/locked out/);

    const b = setup();
    await tick(); await tick();
    b.sock.onclose?.();
    expect(last(b.states)).toEqual(["error", "Lost the connection to the Pi."]);
  });

  it("reports a denied microphone", async () => {
    const e = new Error("denied"); e.name = "NotAllowedError";
    const { states } = setup({ micError: e });
    await tick();
    expect(last(states)[1]).toMatch(/denied/);
  });

  it("makes distinct client ids", () => {
    expect(newClientId(() => 0)).toBe("mic-000000");
    expect(newClientId(() => 0.5)).toMatch(/^mic-[0-9a-f]{6}$/);
  });

  it("stops when the shared screen ends", async () => {
    const { sock, mic, states } = setup();
    await tick(); await tick();
    sock.onopen?.();
    mic.video.ended?.();
    expect(last(states)).toEqual(["error", "The shared screen or camera was stopped."]);
  });
});
