/* Hardware view. The top half is live (GET /api/devices); the bottom half is reference material from the
   design system's data.js (Pi models, USB endpoints, host OS support, latency budget), carried over from the
   Gemini prototype's "HardwareMatrix" idea. Reference figures are labelled as such: they are not measured. */
import { Badge, Card, Icon, Table, type TableColumn } from "../ds";
import type { Client, Device } from "../api/types";
import type { Studio } from "./store";

interface PiModel { id: string; name: string; chipset: string; gadgetPort: string; h264Dec: boolean; mjpegEnc: boolean; maxRes: string; cpu: number; temp: number; power: string }
const PI_MODELS: PiModel[] = [
  { id: "pi4b", name: "Pi 4B", chipset: "BCM2711", gadgetPort: "USB-C (dwc2)", h264Dec: true, mjpegEnc: true, maxRes: "1080p30", cpu: 22, temp: 54, power: "GPIO 5 V or data+power splitter" },
  { id: "pi5", name: "Pi 5", chipset: "BCM2712", gadgetPort: "USB-C (dwc2)", h264Dec: false, mjpegEnc: false, maxRes: "1080p30", cpu: 48, temp: 66, power: "27 W PSU · active cooling" },
  { id: "pizero2w", name: "Pi Zero 2 W", chipset: "RP3A0", gadgetPort: "micro-USB OTG", h264Dec: true, mjpegEnc: false, maxRes: "720p30", cpu: 82, temp: 61, power: "Host-powered" },
  { id: "cm4", name: "CM4 + carrier", chipset: "BCM2711", gadgetPort: "Per carrier", h264Dec: true, mjpegEnc: true, maxRes: "1080p30", cpu: 24, temp: 52, power: "PoE / host-powered" },
];
interface Endpoint { name: string; fn: string; dir: string; format: string; xfer: string; mbps: number; dev: string }
const ENDPOINTS: Endpoint[] = [
  { name: "UVC streaming", fn: "usb_f_uvc", dir: "IN", format: "MJPEG 1920×1080@30", xfer: "Isochronous", mbps: 34.2, dev: "/dev/video3" },
  { name: "UAC2 capture", fn: "usb_f_uac2", dir: "IN", format: "48 kHz S16_LE stereo", xfer: "Isochronous", mbps: 1.5, dev: "hw:1,0" },
  { name: "UAC2 playback", fn: "usb_f_uac2", dir: "OUT", format: "48 kHz S16_LE stereo", xfer: "Isochronous", mbps: 1.5, dev: "hw:1,0" },
  { name: "ECM network", fn: "usb_f_ecm", dir: "BIDIR", format: "CDC Ethernet", xfer: "Bulk", mbps: 0, dev: "usb0 · off by default" },
];
interface HostOs { os: string; icon: string; mjpeg: boolean; yuyv: boolean; uac2: boolean; status: string; apps: string }
const HOSTS: HostOs[] = [
  { os: "macOS 15", icon: "laptop_mac", mjpeg: true, yuyv: true, uac2: true, status: "Native", apps: "Zoom · Teams · QuickTime · OBS" },
  { os: "Windows 11", icon: "desktop_windows", mjpeg: true, yuyv: true, uac2: true, status: "Native (UVC 1.1 / UAC2)", apps: "Teams · Zoom · Camera · OBS" },
  { os: "Linux (v4l2)", icon: "terminal", mjpeg: true, yuyv: true, uac2: true, status: "Native", apps: "Chrome · OBS · ffmpeg" },
  { os: "ChromeOS", icon: "laptop_chromebook", mjpeg: true, yuyv: false, uac2: true, status: "Native", apps: "Meet · Chrome" },
];
const STAGES: { name: string; ms: number; tone: string; desc: string }[] = [
  { name: "Browser encode", ms: 30, tone: "var(--sig-data)", desc: "H.264 encode in the client" },
  { name: "Network", ms: 5, tone: "var(--st-ok)", desc: "LAN / tailnet transit" },
  { name: "Jitter buffer", ms: 40, tone: "var(--sig-video)", desc: "webrtcbin depay + buffer" },
  { name: "Decode", ms: 10, tone: "var(--sig-ai)", desc: "v4l2h264dec" },
  { name: "MJPEG encode", ms: 12, tone: "var(--sig-audio)", desc: "v4l2jpegenc" },
  { name: "UVC / host", ms: 30, tone: "var(--sig-live)", desc: "isochronous + host buffering" },
];

const yes = (v: boolean) => <span style={{ color: v ? "var(--st-ok)" : "var(--text-faint)" }}>{v ? "yes" : "no"}</span>;

export function HardwareView({ studio }: { studio: Studio }) {
  const { s, act } = studio;
  const total = STAGES.reduce((a, b) => a + b.ms, 0);
  const devCols: TableColumn<Device>[] = [
    { label: "Device", key: "name", icon: d => (d.kind === "video" ? "videocam" : "mic") },
    { label: "Kind", key: "kind", width: 90, mono: true },
    { label: "Path", key: "path", width: 160, mono: true },
    { label: "Id", key: "id", width: 120, mono: true },
  ];
  const clientCols: TableColumn<Client>[] = [
    { label: "Client", key: "id", icon: () => "smartphone" },
    { label: "State", width: 120, render: c => <Badge tone={c.onAir ? "live" : "neutral"} dot>{c.onAir ? "On air" : "Idle"}</Badge> },
  ];
  const piCols: TableColumn<PiModel>[] = [
    { label: "Board", key: "name", width: 130 },
    { label: "SoC", key: "chipset", width: 90, mono: true },
    { label: "Gadget port", key: "gadgetPort", width: 130, mono: true },
    { label: "H.264 hw dec", width: 110, render: p => yes(p.h264Dec) },
    { label: "MJPEG hw enc", width: 110, render: p => yes(p.mjpegEnc) },
    { label: "Max", key: "maxRes", width: 80, mono: true },
    { label: "Typ. CPU", width: 80, align: "right", mono: true, render: p => p.cpu + " %" },
    { label: "Typ. temp", width: 90, align: "right", mono: true, render: p => p.temp + " °C" },
    { label: "Power", key: "power" },
  ];
  const epCols: TableColumn<Endpoint>[] = [
    { label: "Function", key: "name", width: 140 },
    { label: "Driver", key: "fn", width: 110, mono: true },
    { label: "Dir", key: "dir", width: 64, mono: true },
    { label: "Format", key: "format", mono: true },
    { label: "Transfer", key: "xfer", width: 110 },
    { label: "Mbps", width: 70, align: "right", mono: true, render: e => e.mbps.toFixed(1) },
    { label: "Node", key: "dev", width: 160, mono: true },
  ];
  const osCols: TableColumn<HostOs>[] = [
    { label: "Host", key: "os", width: 140, icon: h => h.icon },
    { label: "MJPEG", width: 70, render: h => yes(h.mjpeg) },
    { label: "YUYV", width: 70, render: h => yes(h.yuyv) },
    { label: "UAC2", width: 70, render: h => yes(h.uac2) },
    { label: "Driver", key: "status", width: 180 },
    { label: "Tested with", key: "apps" },
  ];
  return (
    <div className="gc-scroll gc-stack" data-testid="hardware-view" style={{ flex: 1, minWidth: 0, overflow: "auto", padding: "var(--sp-7)", display: "flex", flexDirection: "column", gap: "var(--sp-7)", background: "var(--surface-app)" }}>
      <div style={{ display: "grid", gridTemplateColumns: "minmax(0,2fr) minmax(0,1fr)", gap: "var(--sp-7)" }}>
        <Card eyebrow="Live · GET /api/devices" title="Capture devices on the Pi" icon="memory" padding="0" actions={<Badge tone="neutral">{s.devices?.devices.length ?? 0}</Badge>}>
          <Table columns={devCols} rows={s.devices?.devices ?? []} rowKey={d => d.id} dense empty="No V4L2 or ALSA devices reported." style={{ border: 0, borderRadius: 0 }} />
        </Card>
        <Card eyebrow="Live" title="Browser clients" icon="groups" padding="0" actions={<Badge tone="neutral">{s.devices?.clients.length ?? 0}</Badge>}>
          <Table columns={clientCols} rows={s.devices?.clients ?? []} rowKey={c => c.id} dense empty="No browser clients connected." rail={c => (c.onAir ? "var(--sig-live)" : null)} style={{ border: 0, borderRadius: 0 }} />
        </Card>
      </div>
      <button type="button" className="gc-row" onClick={act.refreshDevices} style={{ alignSelf: "flex-start", display: "flex", alignItems: "center", gap: "var(--sp-3)", padding: "var(--sp-3) var(--sp-4)", background: "transparent", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)", cursor: "pointer", font: "var(--type-ui)", color: "var(--text-muted)" }}>
        <Icon name="refresh" size="sm" />Rescan devices
      </button>

      <Card eyebrow="Reference · not measured" title={"Glass-to-glass latency budget · ~" + total + " ms"} icon="timer" padding="var(--sp-6)">
        <div style={{ display: "flex", height: 14, borderRadius: "var(--r-xs)", overflow: "hidden", border: "1px solid var(--line-border)" }}>
          {STAGES.map(st => <div key={st.name} title={st.name + " · " + st.ms + " ms"} style={{ width: (st.ms / total) * 100 + "%", background: st.tone }} />)}
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))", gap: "var(--sp-4)", marginTop: "var(--sp-6)" }}>
          {STAGES.map(st => (
            <div key={st.name} style={{ display: "flex", gap: "var(--sp-3)", alignItems: "flex-start" }}>
              <span style={{ width: 8, height: 8, marginTop: 4, borderRadius: "var(--r-pill)", background: st.tone, flex: "0 0 auto" }} />
              <span style={{ display: "flex", flexDirection: "column", gap: 1, minWidth: 0 }}>
                <span style={{ font: "var(--type-ui)", color: "var(--text-primary)" }}>{st.name} <span style={{ font: "var(--type-meta)", color: "var(--text-muted)" }}>{st.ms} ms</span></span>
                <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>{st.desc}</span>
              </span>
            </div>
          ))}
        </div>
      </Card>

      <Card eyebrow="Reference" title="Raspberry Pi models" icon="developer_board" padding="0" footer="Typical figures for MJPEG 1080p30 out · Pi 5 has no H.264 hardware decoder, so it transcodes in software">
        <Table columns={piCols} rows={PI_MODELS} rowKey={p => p.id} dense style={{ border: 0, borderRadius: 0 }} />
      </Card>
      <Card eyebrow="Reference" title="USB gadget endpoints" icon="usb" padding="0" footer="USB 2.0 high-speed · 480 Mbps bus">
        <Table columns={epCols} rows={ENDPOINTS} rowKey={e => e.name} dense style={{ border: 0, borderRadius: 0 }} />
      </Card>
      <Card eyebrow="Reference" title="Host OS support · no drivers, no host software" icon="laptop_mac" padding="0">
        <Table columns={osCols} rows={HOSTS} rowKey={h => h.os} dense style={{ border: 0, borderRadius: 0 }} />
      </Card>
    </div>
  );
}
