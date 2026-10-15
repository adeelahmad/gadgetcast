# GadgetCast Studio (web/)

The single-page app the Pi serves: a patch-bay pipeline editor, mixer, diagnostics and hardware view for
GadgetCast. Vite + React 18.3.1 + TypeScript (strict) + React Flow 11.11.4, all from npm with exact pins.
No Tailwind, no CDN loads, no AI SDKs or API keys in the browser.

## Run

```sh
cd web
npm ci                # if NODE_ENV=production is set in your shell, use: npm ci --include=dev
npm run dev           # http://localhost:5173, proxies /api and /ws (WebSocket too) to http://localhost:8080
npm run lint          # tsc --noEmit
npm test              # vitest
npm run build         # tsc --noEmit && vite build → web/dist
npx vite preview      # serve the built dist on :4173 (same proxy)
```

Point the proxy at another Pi with `GADGETCAST_API=http://gadgetcast.local:8080 npm run dev`.
Append `?demo=1` to the URL to force the offline demo even when a Pi answers.

`web/dist/` is build output embedded by the Go server (`web/embed.go`, owned elsewhere). Only
`dist/.gitkeep` is tracked; the build puts it back if Vite's outDir wipe removed it.

## Layout

```
src/
  api/        the data layer; the only place that knows mock vs real
    types.ts      wire types, verbatim from docs/api.md
    backend.ts    the Backend interface every screen talks to, ApiError
    client.ts     HTTP client for every endpoint + /api/health probe
    events.ts     /api/events WebSocket with exponential-backoff reconnect (0.5 s → 15 s)
    mock.ts       offline demo backend: same contract, in memory, simulated events
    mockData.ts   catalogue, presets, devices and log seeds ported from the design system's data.js
    select.ts     selectBackend(): probe /api/health → HTTP client, else demo
  graph/      UI graph model, independent of React
    model.ts      GraphDoc / GraphNode / GraphEdge, catalogue index, port resolution, id helpers
    mapping.ts    GraphDoc ⇄ Pipeline (wire), plus JSON import of both shapes
    validate.ts   port-compatibility rule, used by the canvas and by import
    gadgetscript.ts  configfs + gst-launch sketch generated from the graph
  ds/         the design system, ported to typed TSX (pixels only)
    tokens/ styles.css components.css assets/  copied from the system; fonts.css rewritten for local fonts
    core/     Icon Button IconButton Badge Card Input Select Switch Slider Tabs Modal Tooltip Menu Toast
    studio/   PortHandle PipelineNode PatchWire Meters (LevelMeter WaveScope OnAirIndicator MetricPill)
              Lists (LogStream PresetRow DeviceRow Table)
  app/        the Studio: store.ts (state + actions), App, Header + StatusBar, Sidebar + Rail, Canvas,
              Inspector, Panels (Mixer, Diagnostics), HardwareView, Modals (⌘K palette, settings,
              shortcuts, save-as, node config, delete, discard), QuickCapture
```

### Canvas

React Flow owns pan, zoom, node drag, marquee selection, minimap, connection radius and edge reconnection.
The design system owns the pixels: `nodeTypes.gc` renders `PipelineNode`, `edgeTypes.gc` renders `PatchWire`,
and `connectionLineComponent` renders `PatchWire` for the cable in flight. `PipelineNode.renderHandle` drops a
React Flow `<Handle>` (ids `in-<n>` / `out-<n>`) inside each `PortHandle` hit area. Port typing is enforced in
`isValidConnection` via `graph/validate.ts` (same kind only, no self-loop, one cable per input, output → input),
and a refused cable shows a bubble with the reason. Releasing a cable on empty canvas opens quick-add filtered
to nodes with a matching input and patches the new node in. Zoom-to-fit is computed from node bounds so the
graph clears the floating sidebar and inspector. Only thin overrides of React Flow internals live in
`app/studio.css`.

### Theme

Light is the default. The toggle sets `data-theme="dark"` on `<html>`, which remaps the semantic tokens in
`ds/tokens/colors.css`; the choice is remembered in `localStorage` and applied before first paint. Components
reference tokens only; the few hard-coded tints in the system's JSX (Badge, Card footer, OnAirIndicator, button
hover/press) were replaced with `color-mix()` over tokens. LevelMeter uses flat fills instead of a gradient,
per the system's "no gradients" rule. Fonts (Space Grotesk, IBM Plex Sans, JetBrains Mono, Material Symbols
Rounded with the FILL axis) are bundled from `@fontsource` so the Studio works on an offline LAN; the icon
font is the largest asset (about 1.5 MB).

## The mock / real seam

Everything above `src/api` talks to one `Backend` interface. At boot `selectBackend()` calls
`GET /api/health` (2.5 s timeout). `{"ok":true}` → `createHttpBackend()`; anything else (network error,
non-2xx, non-JSON, `ok:false`, timeout) → `createMockBackend()` and the header shows an **Offline demo**
badge with the reason. Screens learn only `backend.mode` (for the badge) and may call the optional
`backend.simulatedLevels()`, which only the demo implements: the v0 API reports no audio levels, so on a real
Pi the meters stay at rest and say so.

Through that interface the app lists the node catalogue, lists / opens / saves (PUT) / deletes pipelines,
runs and stops them, reads status and devices, and subscribes to `/api/events` for metrics (header,
inspector, diagnostics, status bar), status (tally, Go on air / Stop) and the log tail (diagnostics,
last 500 lines). "Go on air" saves first when the graph has unsaved edits. "Load on boot" PUTs
`isDefault:true` on the chosen pipeline and `false` on any other default. Autosave (on by default) PUTs an
already-saved pipeline every 30 s while it has edits.

Graph ↔ wire mapping (`graph/mapping.ts`): node `title`, `enabled` (default true) and `config` carry over;
inspector fields (URL, format, gain, model, prompt, sample rate, Bluetooth profile, logic expression…) are all
stored in `config`; edges map `{from:{node,port}}` ⇄ `from`/`fromPort`; positions round to whole pixels;
cables that reference missing nodes are dropped with a warning. Node types the catalogue does not know still
render, with ports reconstructed from their cables, so a newer server's pipelines round-trip.

## Tests

`npm test` runs `src/**/*.test.ts` (28 tests):

- `graph/validate.test.ts`: same-kind only with the exact reason text, self-loop, one cable per input, re-routing
  a cable vacates its own input, fan-out allowed, missing ports and wrong direction; whole-graph problems.
- `graph/mapping.test.ts`: Pipeline → graph → Pipeline round trip (full and sparse), every bundled preset
  round-trips and validates, dangling cables dropped, unknown types keep ports, import of the prototype export
  shape, id slugs, GadgetScript output.
- `api/select.test.ts`: mock fallback selection (healthy, network error, 5xx, non-JSON, `ok:false`, timeout,
  forced demo), HTTP method + path for every endpoint, `{error}` surfacing and id validation, event parsing,
  WebSocket reconnect/backoff, and the mock honouring the contract (404, 409 while running, 400 bad id, single
  default).

## Feature provenance

**From the Claude design system (theme and the Studio itself):** tokens, components and every visual;
Header + StatusBar, floating Sidebar (palette, library, hardware & clients, shortcuts), React Flow canvas with
typed ports, rejection bubble, quick-add, context menus, multi-select bar, minimap, Inspector, Mixer,
Diagnostics with GadgetScript, ⌘K palette, settings, save preset, node config / prompt editor, JSON
import/export, mock catalogue and presets (data.js).

**Taken from the Gemini AI Studio prototype (features only, restyled with the system):**

- Hardware view (its `HardwareMatrix` + `HostOSBench`): live devices and clients tables, plus clearly
  labelled reference tables for Pi models, USB gadget endpoints, host OS support and a glass-to-glass latency
  budget. The figures come from the design system's data.js, not from the prototype's numbers.
- GadgetScript generated from the actual graph (its `GadgetScriptViewer` showed static scripts): configfs for
  the gadget sinks present, one gst-launch chain per media path, with copy and download.
- Diagnostics log filters: level (warn+, errors) and text search, download and clear.
- Quick capture (its `QuickCaptureModal`): reimplemented for real with `getUserMedia` / `getDisplayMedia`,
  canvas frame grab to PNG and `MediaRecorder` clips, downloaded locally. Requires https or localhost.
- Theme remembered across visits (its theme/auto-save settings).

**Skipped as hallucinated or not supported by the v0 contract:**

- Gemini calls from the browser (`@google/genai`, `/api/gemini/*` analyze-frame, speech, live-chat) and
  any API key in the client. AI nodes are configuration only; model calls belong on the Pi.
- The prototype's Express backend paths (`/api/pi/status`, `/api/pi/pipeline/apply` with a raw gst command,
  `/api/pi/pipeline/probe`, `/api/pi/switcher`, `/api/pi/configfs/reload`, `/api/pi/config/default`,
  `/api/pi/logs`). None exist in docs/api.md; equivalents are the contract's run/stop/status/events.
- Vision switcher cut/auto/PiP/fade transitions, preview/program video, per-node latency/CPU, dropped
  frames and fps, audio levels: no API for them. The mixer shows sources from the graph, gain (saved in node
  config) and Take = run; meters are simulated only in the demo and labelled.
- Self-test, "USB bandwidth & power calculator" numbers, resolution-target switcher, simulated capture that
  returned stock photos, keyboard shortcut rebinding, editable device defaults (PIN, bind interface, default
  UVC format): either invented figures or nothing on the Pi to apply them to.
- data.js claimed Windows 11 needs a "WinUSB fallback"; Windows supports UVC and UAC2 class drivers natively,
  so the reference table says Native.

## Known gaps

- `/ws` (WebRTC signaling) is proxied but the Studio does not use it; pairing and media are out of scope.
- No per-node runtime status from the API, so nodes show "live" only as "the running pipeline".
- Video monitors are placeholders; the Studio never receives media.
- Icon font is the full Material Symbols set (1.5 MB); subsetting would cut that substantially.
