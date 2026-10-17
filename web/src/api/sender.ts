/* Sends this browser's microphone to the Pi, which plays it into the USB mic the host sees.
   Flow over GET /ws (docs/api.md): join → submit-pin → media-offer (non-trickle: the offer carries every
   ICE candidate) → media-answer → switch-source to ourselves → on-air-state. Browser APIs are injected so
   the flow is unit-tested without a DOM. */

export type SendState = "idle" | "connecting" | "on-air" | "error" | "stopped";

export interface SenderSocket {
  send(data: string): void;
  close(): void;
  onopen: (() => void) | null;
  onmessage: ((ev: { data: unknown }) => void) | null;
  onclose: (() => void) | null;
}

/** The subset of RTCPeerConnection the sender uses. */
export interface SenderPeer {
  addTransceiver(track: MediaStreamTrack, init: { direction: "sendonly"; streams?: MediaStream[] }): unknown;
  createOffer(): Promise<{ type: string; sdp?: string }>;
  setLocalDescription(d: { type: string; sdp?: string }): Promise<void>;
  setRemoteDescription(d: { type: "answer"; sdp: string }): Promise<void>;
  readonly localDescription: { sdp: string } | null;
  readonly iceGatheringState: string;
  onicegatheringstatechange: (() => void) | null;
  close(): void;
}

export interface SenderDeps {
  socket(url: string): SenderSocket;
  peer(): SenderPeer;
  getMic(): Promise<MediaStream>;
  url: string;
  /** Max wait for ICE gathering before sending what we have. */
  gatherTimeoutMs?: number;
}

export interface SenderEvents {
  onState(state: SendState, detail?: string): void;
}

export const newClientId = (rand: () => number = Math.random) =>
  "mic-" + Math.floor(rand() * 0xffffff).toString(16).padStart(6, "0");

function waitForGathering(pc: SenderPeer, ms: number): Promise<void> {
  if (pc.iceGatheringState === "complete") return Promise.resolve();
  return new Promise(resolve => {
    const t = setTimeout(resolve, ms);
    pc.onicegatheringstatechange = () => {
      if (pc.iceGatheringState === "complete") { clearTimeout(t); resolve(); }
    };
  });
}

/** Starts sending; returns a stop function. Errors are reported through onState("error"). */
export function startMicSender(deps: SenderDeps, pin: string, ev: SenderEvents, clientId = newClientId()): () => void {
  let stopped = false;
  let ws: SenderSocket | null = null;
  let pc: SenderPeer | null = null;
  let stream: MediaStream | null = null;

  const cleanup = () => {
    stream?.getTracks().forEach(t => t.stop());
    pc?.close();
    if (ws) { ws.onclose = null; ws.close(); }
    stream = null; pc = null; ws = null;
  };
  const fail = (msg: string) => {
    if (stopped) return;
    stopped = true;
    cleanup();
    ev.onState("error", msg);
  };
  const send = (type: string, payload: unknown) => ws?.send(JSON.stringify({ type, payload }));

  ev.onState("connecting");
  void (async () => {
    try {
      stream = await deps.getMic();
    } catch (e) {
      const name = e instanceof Error ? e.name : "";
      fail(name === "NotAllowedError" ? "Microphone permission was denied." : "Could not open the microphone: " + String(e instanceof Error ? e.message : e));
      return;
    }
    if (stopped) { cleanup(); return; }
    const track = stream.getAudioTracks()[0];
    if (!track) { fail("No microphone track."); return; }

    pc = deps.peer();
    pc.addTransceiver(track, { direction: "sendonly", streams: [stream] });
    try {
      await pc.setLocalDescription(await pc.createOffer());
      await waitForGathering(pc, deps.gatherTimeoutMs ?? 3000);
    } catch (e) {
      fail("WebRTC setup failed: " + String(e instanceof Error ? e.message : e));
      return;
    }
    if (stopped || !pc?.localDescription) return;
    const offer = pc.localDescription.sdp;

    ws = deps.socket(deps.url);
    ws.onopen = () => {
      send("join", { client: clientId });
      send("submit-pin", { pin });
      send("media-offer", { sdp: offer });
    };
    ws.onclose = () => fail("Lost the connection to the Pi.");
    ws.onmessage = msg => {
      let f: { type?: string; payload?: Record<string, unknown> };
      try { f = JSON.parse(String(msg.data)); } catch { return; }
      const p = f.payload ?? {};
      if (f.type === "error") {
        const reason = String(p.reason ?? "unknown error");
        // A wrong PIN is followed by "pair with the PIN"; report the first, clearer one.
        fail(/incorrect pin/i.test(reason) ? "Wrong PIN." : /rejected/i.test(reason) ? "Too many wrong PINs: this browser is locked out." : reason);
      } else if (f.type === "media-answer" && typeof p.sdp === "string") {
        pc?.setRemoteDescription({ type: "answer", sdp: p.sdp })
          .then(() => send("switch-source", { target: clientId }))
          .catch(e => fail("Bad answer from the Pi: " + String(e instanceof Error ? e.message : e)));
      } else if (f.type === "on-air-state") {
        if (p.live === true && p.onair === clientId) ev.onState("on-air");
        else if (!stopped) ev.onState("connecting", "Another device is on air.");
      }
    };
  })();

  return () => {
    if (stopped) return;
    stopped = true;
    cleanup();
    ev.onState("stopped");
  };
}
