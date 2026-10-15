/* Offline demo data, ported from the design system's ui_kits/studio/data.js and reshaped to the v0 contract
   (NodeGroup / Pipeline / DevicesResponse). Only the mock backend reads this file. */
import type { DevicesResponse, LogEventData, NodeGroup, NodeType, Pipeline, PipelineEdge, PipelineNode, PortKind } from "./types";

export const CATALOG: NodeGroup[] = [
  { group: "Browser", items: [
    { type: "webrtc_in", title: "WebRTC Client", icon: "cast_connected", cls: "source", category: "source", media: "both", gst: "webrtcbin bundle-policy=max-bundle", subtitle: "any browser on the LAN", out: [["video", "H.264"], ["audio", "Opus"]] },
    { type: "browser_cam", title: "Browser Camera", icon: "videocam", cls: "source", category: "source", media: "video", gst: "webrtcbin ! rtph264depay", subtitle: "getUserMedia · client cam", out: [["video", "H.264"]] },
    { type: "browser_mic", title: "Browser Mic", icon: "mic", cls: "source", category: "source", media: "audio", gst: "webrtcbin ! rtpopusdepay", subtitle: "getUserMedia · Opus 48 kHz", out: [["audio", "Opus"]] },
    { type: "browser_screen", title: "Screen Share", icon: "screen_share", cls: "source", category: "source", media: "both", gst: "webrtcbin (getDisplayMedia)", subtitle: "tab · window · display", out: [["video", "H.264"], ["audio", "Opus"]] },
    { type: "remote_guest", title: "Remote Guest", icon: "person_add", cls: "source", category: "source", media: "both", gst: "webrtcbin (tailscale)", subtitle: "off-LAN client via tailnet", out: [["video", "H.264"], ["audio", "Opus"]] },
  ] },
  { group: "Network", items: [
    { type: "rtsp", title: "RTSP Source", icon: "sensors", cls: "source", category: "source", media: "both", gst: "rtspsrc latency=100", url: "rtsp://10.0.0.14:554/cam1", out: [["video", "H.264"], ["audio", "AAC"]] },
    { type: "rtmp", title: "RTMP Ingest", icon: "podcasts", cls: "source", category: "source", media: "both", gst: "rtmp2src", url: "rtmp://gadgetcast.local/live/studio", out: [["video", "H.264"], ["audio", "AAC"]] },
    { type: "rtp", title: "RTP Source", icon: "swap_calls", cls: "source", category: "source", media: "video", gst: "udpsrc ! rtph264depay", url: "udp://0.0.0.0:5004", out: [["video", "H.264"]] },
    { type: "srt", title: "SRT Source", icon: "lan", cls: "source", category: "source", media: "both", gst: "srtsrc mode=listener", url: "srt://0.0.0.0:9000?mode=listener", out: [["video", "H.264"], ["audio", "AAC"]] },
    { type: "hls", title: "HLS / HTTP", icon: "public", cls: "source", category: "source", media: "both", gst: "souphttpsrc ! hlsdemux", url: "https://cdn.example.com/live/index.m3u8", out: [["video", "H.264"], ["audio", "AAC"]] },
    { type: "url", title: "Network Stream", icon: "link", cls: "source", category: "source", media: "both", gst: "uridecodebin", url: "https://cdn.example.com/live/stream.mp4", subtitle: "validated on drop", out: [["video", "Raw"], ["audio", "PCM"]] },
    { type: "file", title: "File / Loop", icon: "movie", cls: "source", category: "source", media: "both", gst: "filesrc ! decodebin", url: "/media/loops/desk-loop.mp4", out: [["video", "Raw"], ["audio", "PCM"]] },
    { type: "testpattern", title: "Test Pattern", icon: "grid_on", cls: "source", category: "source", media: "video", gst: "videotestsrc pattern=smpte", subtitle: "SMPTE bars · 1080p30", out: [["video", "Raw"]] },
    { type: "tonegen", title: "Tone Generator", icon: "waves", cls: "source", category: "source", media: "audio", gst: "audiotestsrc wave=sine", subtitle: "sine 1 kHz · −12 dBFS", out: [["audio", "PCM"]] },
  ] },
  { group: "Hardware", items: [
    { type: "csi", title: "CSI Camera", icon: "photo_camera", cls: "source", category: "source", media: "video", gst: "libcamerasrc", dev: "/dev/video0", subtitle: "/dev/video0 · libcamera", out: [["video", "NV12"]] },
    { type: "usb_cam", title: "USB Camera (UVC in)", icon: "camera", cls: "source", category: "source", media: "video", gst: "v4l2src device=/dev/video4", dev: "/dev/video4", subtitle: "/dev/video4 · MJPEG in", out: [["video", "MJPEG"]] },
    { type: "i2s_mic", title: "I²S / USB Mic", icon: "mic_external_on", cls: "source", category: "source", media: "audio", gst: "alsasrc device=hw:2,0", dev: "hw:2,0", subtitle: "hw:2,0 · 48 kHz S16_LE", out: [["audio", "PCM"]] },
  ] },
  { group: "Bluetooth", items: [
    { type: "bt_src", title: "BT A2DP Source", icon: "bluetooth_audio", cls: "bluetooth", category: "bluetooth", media: "audio", gst: "bluezsrc profile=a2dp", subtitle: "paired headset / phone", out: [["audio", "PCM"]] },
    { type: "bt_sink", title: "BT A2DP Sink", icon: "bluetooth", cls: "bluetooth", category: "bluetooth", media: "audio", gst: "bluezsink profile=a2dp", subtitle: "monitor on headphones", in: [["audio", "PCM"]] },
    { type: "bt_hfp", title: "BT Handsfree (HFP)", icon: "headset_mic", cls: "bluetooth", category: "bluetooth", media: "both", gst: "bluezsrc profile=hfp", subtitle: "16 kHz duplex SCO", in: [["audio", "PCM"]], out: [["audio", "PCM"]] },
    { type: "bt_remote", title: "BT Camera Remote", icon: "settings_remote", cls: "bluetooth", category: "bluetooth", media: "data", gst: "n/a (control plane)", subtitle: "shutter → take / cut", out: [["data", "Event"]] },
  ] },
  { group: "Decode", items: [
    { type: "h264_dec", title: "H.264 Decode", icon: "memory", cls: "video", category: "decoder", media: "video", gst: "v4l2h264dec", subtitle: "hw · /dev/video10", in: [["video", "H.264"]], out: [["video", "NV12"]] },
    { type: "jpeg_dec", title: "MJPEG Decode", icon: "image", cls: "video", category: "decoder", media: "video", gst: "v4l2jpegdec", subtitle: "hw JPEG → NV12", in: [["video", "MJPEG"]], out: [["video", "NV12"]] },
    { type: "opus_dec", title: "Opus Decode", icon: "graphic_eq", cls: "audio", category: "decoder", media: "audio", gst: "opusdec plc=true", subtitle: "→ 48 kHz PCM", in: [["audio", "Opus"]], out: [["audio", "PCM"]] },
  ] },
  { group: "Audio", items: [
    { type: "mixer_a", title: "Audio Mixer", icon: "tune", cls: "audio", category: "filter", media: "audio", gst: "audiomixer", subtitle: "4 bus · per-strip gain", in: [["audio", "A"], ["audio", "B"]], out: [["audio", "Mix"]] },
    { type: "aec", title: "AEC / Gate", icon: "noise_control_off", cls: "audio", category: "filter", media: "audio", gst: "webrtcdsp", subtitle: "AEC3 + noise gate", in: [["audio", "PCM"]], out: [["audio", "PCM"]] },
    { type: "resampler", title: "Drift Resampler", icon: "sync_alt", cls: "audio", category: "filter", media: "audio", gst: "audioresample", subtitle: "clock-drift compensation", in: [["audio", "PCM"]], out: [["audio", "PCM"]] },
    { type: "opus_enc", title: "Opus Encode", icon: "equalizer", cls: "audio", category: "encoder", media: "audio", gst: "opusenc bitrate=64000", subtitle: "back to browser clients", in: [["audio", "PCM"]], out: [["audio", "Opus"]] },
  ] },
  { group: "Video", items: [
    { type: "mixer_v", title: "Vision Mixer", icon: "grid_view", cls: "video", category: "filter", media: "video", gst: "compositor", subtitle: "cut · fade · PiP · split", in: [["video", "A"], ["video", "B"]], out: [["video", "Program"]] },
    { type: "scale", title: "Scale / Convert", icon: "aspect_ratio", cls: "video", category: "filter", media: "video", gst: "v4l2convert", subtitle: "1080p30 → 720p30", in: [["video", "Raw"]], out: [["video", "Raw"]] },
    { type: "overlay", title: "Text / Tally Overlay", icon: "title", cls: "video", category: "filter", media: "video", gst: "textoverlay", subtitle: "burn-in caption", in: [["video", "Raw"], ["data", "Text"]], out: [["video", "Raw"]] },
    { type: "mjpeg_enc", title: "MJPEG Encode", icon: "compress", cls: "video", category: "encoder", media: "video", gst: "v4l2jpegenc", subtitle: "primary UVC format", in: [["video", "Raw"]], out: [["video", "MJPEG"]] },
    { type: "h264_enc", title: "H.264 Encode", icon: "movie_filter", cls: "video", category: "encoder", media: "video", gst: "v4l2h264enc", subtitle: "for RTMP / recording", in: [["video", "Raw"]], out: [["video", "H.264"]] },
  ] },
  { group: "AI", items: [
    { type: "gemini_live", title: "Gemini Live", icon: "audio_spark", cls: "ai", category: "ai_processor", media: "both", gst: "appsink ↔ live API", subtitle: "voice conversation · realtime", in: [["audio", "Voice"], ["video", "Frames"]], out: [["audio", "Reply"], ["data", "Text"]] },
    { type: "vision_llm", title: "Gemini Vision", icon: "image_search", cls: "ai", category: "ai_processor", media: "video", gst: "appsink (jpeg snapshots)", subtitle: "frame → description, 1 fps", in: [["video", "Frame"]], out: [["data", "Text"]] },
    { type: "llm", title: "LLM (text)", icon: "bolt", cls: "ai", category: "ai_processor", media: "data", gst: "n/a (control plane)", subtitle: "low-latency text step", in: [["data", "Text"]], out: [["data", "Text"]] },
    { type: "stt", title: "Speech → Text", icon: "speech_to_text", cls: "ai", category: "ai_processor", media: "audio", gst: "appsink → STT stream", subtitle: "streaming transcription", in: [["audio", "PCM"]], out: [["data", "Text"]] },
    { type: "tts", title: "Text → Speech", icon: "record_voice_over", cls: "ai", category: "ai_processor", media: "audio", gst: "appsrc ← TTS PCM", subtitle: "48 kHz synth voice", in: [["data", "Text"]], out: [["audio", "PCM"]] },
  ] },
  { group: "Logic", items: [
    { type: "cond", title: "Conditional", icon: "call_split", cls: "logic", category: "logic", media: "data", gst: "n/a (control plane)", subtitle: "if drift_ppm > 40", in: [["data", "In"]], out: [["data", "True"], ["data", "False"]] },
    { type: "setvar", title: "Set Variable", icon: "data_object", cls: "logic", category: "logic", media: "data", gst: "n/a (control plane)", subtitle: "assign pipeline var", in: [["data", "In"]], out: [["data", "Out"]] },
    { type: "exec", title: "Exec (Python)", icon: "terminal", cls: "logic", category: "logic", media: "data", gst: "n/a (control plane)", subtitle: "run script on the Pi", in: [["data", "Trigger"]], out: [["data", "Result"]] },
  ] },
  { group: "Storage", items: [
    { type: "s3", title: "AWS S3", icon: "cloud_upload", cls: "storage", category: "storage", media: "both", gst: "appsink → boto3", url: "s3://casts/2026-08", in: [["video", "Frame"], ["audio", "PCM"]] },
    { type: "nfs", title: "NFS Mount", icon: "dns", cls: "storage", category: "storage", media: "video", gst: "filesink (nfs)", url: "nfs://nas/casts", in: [["video", "Frame"]] },
    { type: "cifs", title: "CIFS / SMB", icon: "folder_shared", cls: "storage", category: "storage", media: "video", gst: "filesink (cifs)", url: "smb://nas/media", in: [["video", "Frame"]] },
    { type: "nvme", title: "Local NVMe", icon: "storage", cls: "storage", category: "storage", media: "both", gst: "splitmuxsink", url: "/mnt/nvme/casts", in: [["video", "Frame"], ["audio", "PCM"]] },
  ] },
  { group: "USB gadget", items: [
    { type: "uvc_out", title: "UVC Video Out", icon: "usb", cls: "sink", category: "endpoint", media: "video", gst: "v4l2sink device=/dev/video3", dev: "/dev/video3", subtitle: "/dev/video3 · host sees a webcam", in: [["video", "MJPEG"]] },
    { type: "uac_out", title: "UAC2 Mic Out", icon: "mic_external_on", cls: "sink", category: "endpoint", media: "audio", gst: "alsasink device=hw:1,0", dev: "hw:1,0", subtitle: "hw:1,0 · host hears this", in: [["audio", "PCM"]] },
    { type: "uac_in", title: "UAC2 Host Spk", icon: "speaker", cls: "source", category: "source", media: "audio", gst: "alsasrc device=hw:1,0", dev: "hw:1,0", subtitle: "what the host plays", out: [["audio", "PCM"]] },
    { type: "rtmp_out", title: "RTMP Publish", icon: "cell_tower", cls: "sink", category: "endpoint", media: "both", gst: "flvmux ! rtmp2sink", url: "rtmp://a.rtmp.youtube.com/live2/…", in: [["video", "H.264"], ["audio", "AAC"]] },
    { type: "record", title: "Local Recorder", icon: "fiber_manual_record", cls: "sink", category: "endpoint", media: "both", gst: "splitmuxsink max-size-time=60000000000", subtitle: "/media/recordings", in: [["video", "H.264"], ["audio", "PCM"]] },
  ] },
];

const BY_TYPE: Record<string, NodeType> = {};
CATALOG.forEach(g => g.items.forEach(i => { BY_TYPE[i.type] = i; }));

const AI_PROMPT = "You are a meeting copilot. Watch the video frames and listen to the call. Interrupt only when a decision or an action item is stated.";

/** Compact preset spec: nodes as [type, x, y], edges as [fromIdx, fromPort, toIdx, toPort]. */
interface PresetSpec { id: string; name: string; description: string; isDefault?: boolean; nodes: [string, number, number][]; edges: [number, number, number, number][] }

const PRESETS: PresetSpec[] = [
  { id: "usb-webcam", name: "USB Webcam Streaming", description: "Browser screen → MJPEG UVC + UAC2 mic out", isDefault: true,
    nodes: [["browser_screen", 32, 48], ["h264_dec", 272, 16], ["mjpeg_enc", 512, 16], ["uvc_out", 752, 16], ["opus_dec", 272, 240], ["aec", 512, 240], ["uac_out", 752, 240]],
    edges: [[0, 0, 1, 0], [1, 0, 2, 0], [2, 0, 3, 0], [0, 1, 4, 0], [4, 0, 5, 0], [5, 0, 6, 0]] },
  { id: "meeting-loop", name: "Meeting Loop", description: "Looping file into the UVC webcam",
    nodes: [["file", 32, 64], ["mixer_v", 288, 48], ["mjpeg_enc", 528, 48], ["uvc_out", 752, 48], ["uac_in", 32, 272]],
    edges: [[0, 0, 1, 0], [1, 0, 2, 0], [2, 0, 3, 0]] },
  { id: "gemini-audio", name: "Gemini Audio Analysis", description: "Browser mic + CSI frames into Gemini Live",
    nodes: [["browser_mic", 32, 48], ["opus_dec", 256, 48], ["gemini_live", 480, 32], ["uac_out", 736, 32], ["csi", 32, 272], ["vision_llm", 480, 272]],
    edges: [[0, 0, 1, 0], [1, 0, 2, 0], [4, 0, 2, 1], [2, 0, 3, 0], [4, 0, 5, 0]] },
  { id: "conference-room", name: "Conference Room Brain", description: "RTSP camera + Bluetooth audio mixed to USB",
    nodes: [["rtsp", 32, 32], ["h264_dec", 256, 16], ["mjpeg_enc", 480, 16], ["uvc_out", 720, 16], ["bt_src", 32, 256], ["mixer_a", 320, 240], ["uac_out", 720, 240]],
    edges: [[0, 0, 1, 0], [1, 0, 2, 0], [2, 0, 3, 0], [0, 1, 5, 0], [4, 0, 5, 1], [5, 0, 6, 0]] },
  { id: "storage-vision", name: "Storage + Vision Log", description: "CSI frames to S3 every 30 s with a vision log",
    nodes: [["csi", 32, 64], ["scale", 272, 48], ["vision_llm", 512, 32], ["s3", 736, 192], ["exec", 512, 272]],
    edges: [[0, 0, 1, 0], [1, 0, 2, 0], [1, 0, 3, 0], [2, 0, 4, 0]] },
];

function buildPreset(p: PresetSpec): Pipeline {
  const nodes: PipelineNode[] = p.nodes.map(([type, x, y], i) => {
    const nt = BY_TYPE[type];
    const config: Record<string, unknown> = {};
    if (nt.url) config.url = nt.url;
    if (nt.cls === "ai") { config.prompt = AI_PROMPT; config.model = "gemini-live-2.5-flash"; }
    return { id: type + "_" + (i + 1), type, x, y, title: nt.title, enabled: true, config };
  });
  const edges: PipelineEdge[] = p.edges.map(([a, ap, b, bp], i) => {
    const out = BY_TYPE[p.nodes[a][0]].out ?? [];
    const kind: PortKind = out[ap]?.[0] ?? "data";
    return { id: "e" + (i + 1), from: nodes[a].id, fromPort: ap, to: nodes[b].id, toPort: bp, kind };
  });
  return { id: p.id, name: p.name, description: p.description, isDefault: !!p.isDefault, nodes, edges, updatedAt: "2026-10-01T09:00:00Z" };
}

export function seedPipelines(): Pipeline[] { return PRESETS.map(buildPreset); }

export const DEVICES: DevicesResponse = {
  devices: [
    { id: "video0", kind: "video", name: "Pi Camera Module 3", path: "/dev/video0" },
    { id: "video4", kind: "video", name: "Logitech BRIO", path: "/dev/video4" },
    { id: "video3", kind: "video", name: "UVC gadget (host webcam)", path: "/dev/video3" },
    { id: "hw2", kind: "audio", name: "I²S MEMS mic", path: "hw:2,0" },
    { id: "hw1", kind: "audio", name: "UAC2 gadget", path: "hw:1,0" },
  ],
  clients: [
    { id: "10.0.0.31 · iPhone", onAir: false },
    { id: "10.0.0.18 · MacBook", onAir: false },
    { id: "100.88.4.2 · tailnet guest", onAir: false },
  ],
};

/** Log seeds from data.js, re-homed onto the contract's three subsystems. The original source tag stays in the message. */
export const LOG_SEEDS: LogEventData[] = [
  { subsystem: "server", level: "info", msg: "dwc2: gadget bound 0x1d6b:0x0104 composite (uvc+uac2)" },
  { subsystem: "pipeline", level: "info", msg: "uvc: streamon MJPEG 1920x1080@30 · 34.2 Mbps" },
  { subsystem: "pipeline", level: "info", msg: "gst: webrtcbin0 → v4l2h264dec → v4l2convert → jpegenc → v4l2sink" },
  { subsystem: "pipeline", level: "warn", msg: "uac2: resampler drift +0.7 ppm, correcting" },
  { subsystem: "session", level: "info", msg: "signal: client 10.0.0.31 joined · offer sdp h264/opus" },
  { subsystem: "session", level: "info", msg: "gemini: live session open · audio + frames @ 1 fps" },
  { subsystem: "pipeline", level: "info", msg: "uvc: host SET_CUR dwFrameInterval 333333" },
  { subsystem: "pipeline", level: "info", msg: "gst: pipeline state PLAYING (2.1 ms)" },
  { subsystem: "pipeline", level: "info", msg: "uac2: host opened capture hw:1,0 · 48000 Hz S16_LE" },
  { subsystem: "server", level: "info", msg: "bluez: a2dp source connected AC:80:0A:41:9B:2E · ldac" },
  { subsystem: "server", level: "error", msg: "dwc2: ep2in xfer stall, retrying" },
];
