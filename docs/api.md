# Pi-side HTTP / WebSocket API (v0)

Contract between the Go server (`cmd/gadgetcast`, `internal/server`) and the Studio web UI (`web/`).
All JSON, UTF-8. The UI treats every endpoint as optional: if the server is unreachable it falls back
to its bundled mock data and shows an "offline / demo" badge.

## REST

| Method | Path | Body / Response |
| --- | --- | --- |
| GET | `/api/health` | `{"ok":true,"version":"0.1.1"}` (the release tag without its "v"; `0.0.0-dev` in unstamped builds) |
| GET | `/api/nodes` | `NodeGroup[]` — the node catalogue |
| GET | `/api/pipelines` | `Pipeline[]` |
| GET | `/api/pipelines/{id}` | `Pipeline` or 404 |
| PUT | `/api/pipelines/{id}` | body `Pipeline` (id must match path) → saved `Pipeline` (server sets `updatedAt`) |
| DELETE | `/api/pipelines/{id}` | 204 |
| POST | `/api/pipelines/{id}/run` | 200 `Status`; marks it the running pipeline (no media yet) |
| POST | `/api/pipelines/stop` | 200 `Status` |
| GET | `/api/status` | `Status` |
| GET | `/api/devices` | `{"devices":Device[],"clients":Client[]}` |

Errors: non-2xx with `{"error":"human readable"}`.

```ts
type PortKind = "video" | "audio" | "ai" | "data";
type Port = [PortKind, string];            // [kind, format label] e.g. ["video","H.264"]
type NodeType = {
  type: string; title: string; icon: string; cls: string; category: string;
  media?: "video" | "audio" | "both" | "ai" | "data";
  gst?: string; subtitle?: string; url?: string; dev?: string;
  in?: Port[]; out?: Port[];
};
type NodeGroup = { group: string; items: NodeType[] };

type PipelineNode = { id: string; type: string; x: number; y: number;
  title?: string; enabled?: boolean; config?: Record<string, unknown> };
type PipelineEdge = { id: string; from: string; fromPort: number; to: string; toPort: number; kind: PortKind };
type Pipeline = { id: string; name: string; description?: string; isDefault?: boolean;
  nodes: PipelineNode[]; edges: PipelineEdge[]; updatedAt?: string };

type Metrics = { cpuPct: number; ramUsedMB: number; ramTotalMB: number; tempC: number | null; uptimeS: number; loadAvg: number[] };
type Status = { running: string | null; onAir: string | null; clients: number; metrics: Metrics };
type Device = { id: string; kind: "video" | "audio"; name: string; path: string };
type Client = { id: string; onAir: boolean };
```

`id` path segments match `^[a-zA-Z0-9_-]{1,64}$`.

## WebSockets

- `GET /api/events` — server → UI push, one JSON object per frame:
  `{"type":"metrics","ts":"RFC3339","data":Metrics}` every 2 s,
  `{"type":"status","ts":…,"data":Status}` on change,
  `{"type":"log","ts":…,"data":{"level":"info|warn|error","subsystem":"server|session|pipeline","msg":"…"}}`.
- `GET /ws` — WebRTC signaling. Frames are the `internal/signaling` `{type,payload}` envelope, routed
  through `session.Hub`. A PIN is submitted with the extra client→server frame
  `{"type":"submit-pin","payload":{"pin":"1234"}}` (handled by the transport, fed to the hub as `SubmitPIN`).
- Media to the Pi itself: a joined, paired client sends `{"type":"media-offer","payload":{"sdp":"…"}}`
  and gets `{"type":"media-answer","payload":{"sdp":"…"}}`. ICE is not trickled: the offer must carry all
  of the browser's candidates (wait for gathering to complete) and the answer carries all of the Pi's.
  The Pi accepts Opus audio and H.264 or VP8 video (H.264 preferred); only the on-air client's media (`switch-source`) is played into the USB mic and webcam. If the offer's audio is `sendrecv`, the answer adds an Opus track carrying the host's audio (the USB speaker), sent only while that client is on air.
  Unpaired or unjoined clients get an `error` frame.
