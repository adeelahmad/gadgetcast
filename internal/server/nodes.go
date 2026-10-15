package server

// Port shorthands for the catalogue table.
func vid(f string) Port  { return Port{string(PortVideo), f} }
func aud(f string) Port  { return Port{string(PortAudio), f} }
func ai(f string) Port   { return Port{string(PortAI), f} }
func data(f string) Port { return Port{string(PortData), f} }

func ports(p ...Port) []Port { return p }

// nodeCatalogue is the static node palette served by GET /api/nodes. Plain data:
// the UI renders it as-is; the pipeline engine (later sprint) maps "type" and
// "gst" onto real elements.
var nodeCatalogue = []NodeGroup{
	{Group: "Browser", Items: []NodeType{
		{Type: "browser_cam", Title: "Browser camera", Icon: "camera", Cls: "src", Category: "Browser", Media: "video", Subtitle: "getUserMedia video", Out: ports(vid("H.264"))},
		{Type: "browser_mic", Title: "Browser mic", Icon: "mic", Cls: "src", Category: "Browser", Media: "audio", Subtitle: "getUserMedia audio", Out: ports(aud("Opus"))},
		{Type: "browser_screen", Title: "Screen share", Icon: "monitor", Cls: "src", Category: "Browser", Media: "both", Subtitle: "getDisplayMedia", Out: ports(vid("H.264"), aud("Opus"))},
		{Type: "browser_speaker", Title: "Browser speaker", Icon: "volume-2", Cls: "sink", Category: "Browser", Media: "audio", Subtitle: "Play back to the browser", In: ports(aud("Opus"))},
	}},
	{Group: "Network", Items: []NodeType{
		{Type: "rtsp_in", Title: "RTSP source", Icon: "radio", Cls: "src", Category: "Network", Media: "both", Gst: "rtspsrc", URL: "rtsp://", Out: ports(vid("H.264"), aud("AAC"))},
		{Type: "rtmp_in", Title: "RTMP source", Icon: "radio", Cls: "src", Category: "Network", Media: "both", Gst: "rtmpsrc", URL: "rtmp://", Out: ports(vid("H.264"), aud("AAC"))},
		{Type: "rtp_in", Title: "RTP source", Icon: "radio", Cls: "src", Category: "Network", Media: "both", Gst: "udpsrc", URL: "udp://0.0.0.0:5004", Out: ports(vid("H.264"), aud("Opus"))},
		{Type: "http_in", Title: "HTTP / HLS stream", Icon: "globe", Cls: "src", Category: "Network", Media: "both", Gst: "souphttpsrc", URL: "https://", Out: ports(vid("H.264"), aud("AAC"))},
		{Type: "ndi_in", Title: "NDI source", Icon: "cast", Cls: "src", Category: "Network", Media: "both", Out: ports(vid("UYVY"), aud("PCM"))},
	}},
	{Group: "Hardware", Items: []NodeType{
		{Type: "v4l2_in", Title: "V4L2 camera", Icon: "video", Cls: "src", Category: "Hardware", Media: "video", Gst: "v4l2src", Dev: "/dev/video0", Out: ports(vid("MJPEG"))},
		{Type: "csi_cam", Title: "Pi CSI camera", Icon: "aperture", Cls: "src", Category: "Hardware", Media: "video", Gst: "libcamerasrc", Out: ports(vid("NV12"))},
		{Type: "hdmi_in", Title: "HDMI capture", Icon: "tv", Cls: "src", Category: "Hardware", Media: "both", Gst: "v4l2src", Dev: "/dev/video0", Out: ports(vid("YUYV"), aud("PCM"))},
		{Type: "alsa_in", Title: "ALSA input", Icon: "mic", Cls: "src", Category: "Hardware", Media: "audio", Gst: "alsasrc", Dev: "hw:0", Out: ports(aud("PCM"))},
		{Type: "uac2_in", Title: "Host speaker (UAC2)", Icon: "speaker", Cls: "src", Category: "Hardware", Media: "audio", Gst: "alsasrc", Subtitle: "Audio the USB host plays", Out: ports(aud("PCM"))},
	}},
	{Group: "Bluetooth", Items: []NodeType{
		{Type: "bt_audio_in", Title: "Bluetooth audio in", Icon: "bluetooth", Cls: "src", Category: "Bluetooth", Media: "audio", Subtitle: "A2DP sink", Out: ports(aud("SBC"))},
		{Type: "bt_audio_out", Title: "Bluetooth speaker", Icon: "bluetooth", Cls: "sink", Category: "Bluetooth", Media: "audio", Subtitle: "A2DP source", In: ports(aud("PCM"))},
		{Type: "bt_hfp", Title: "Bluetooth headset", Icon: "headphones", Cls: "duplex", Category: "Bluetooth", Media: "audio", Subtitle: "HFP", In: ports(aud("PCM")), Out: ports(aud("mSBC"))},
	}},
	{Group: "Mixers/FX", Items: []NodeType{
		{Type: "video_switch", Title: "Video switcher", Icon: "shuffle", Cls: "fx", Category: "Mixers/FX", Media: "video", Gst: "input-selector", In: ports(vid("raw"), vid("raw")), Out: ports(vid("raw"))},
		{Type: "compositor", Title: "Picture-in-picture", Icon: "layers", Cls: "fx", Category: "Mixers/FX", Media: "video", Gst: "compositor", In: ports(vid("raw"), vid("raw")), Out: ports(vid("raw"))},
		{Type: "audio_mixer", Title: "Audio mixer", Icon: "sliders", Cls: "fx", Category: "Mixers/FX", Media: "audio", Gst: "audiomixer", In: ports(aud("PCM"), aud("PCM")), Out: ports(aud("PCM"))},
		{Type: "scale", Title: "Scale / crop", Icon: "crop", Cls: "fx", Category: "Mixers/FX", Media: "video", Gst: "videoscale", In: ports(vid("raw")), Out: ports(vid("raw"))},
		{Type: "overlay", Title: "Text / image overlay", Icon: "type", Cls: "fx", Category: "Mixers/FX", Media: "video", Gst: "textoverlay", In: ports(vid("raw")), Out: ports(vid("raw"))},
		{Type: "resample", Title: "Resampler", Icon: "activity", Cls: "fx", Category: "Mixers/FX", Media: "audio", Gst: "audioresample", Subtitle: "48 kHz drift control", In: ports(aud("PCM")), Out: ports(aud("PCM"))},
		{Type: "aec", Title: "Echo cancel / denoise", Icon: "wind", Cls: "fx", Category: "Mixers/FX", Media: "audio", Gst: "webrtcdsp", In: ports(aud("PCM")), Out: ports(aud("PCM"))},
	}},
	{Group: "AI", Items: []NodeType{
		{Type: "ai_bg_blur", Title: "Background blur", Icon: "user", Cls: "ai", Category: "AI", Media: "video", In: ports(vid("raw")), Out: ports(vid("raw"))},
		{Type: "ai_transcribe", Title: "Live transcription", Icon: "file-text", Cls: "ai", Category: "AI", Media: "ai", In: ports(aud("PCM")), Out: ports(ai("text"))},
		{Type: "ai_captions", Title: "Burn-in captions", Icon: "subtitles", Cls: "ai", Category: "AI", Media: "video", In: ports(vid("raw"), ai("text")), Out: ports(vid("raw"))},
		{Type: "ai_detect", Title: "Object detection", Icon: "scan", Cls: "ai", Category: "AI", Media: "ai", In: ports(vid("raw")), Out: ports(ai("boxes"))},
		{Type: "ai_autoframe", Title: "Auto-framing", Icon: "maximize", Cls: "ai", Category: "AI", Media: "video", In: ports(vid("raw"), ai("boxes")), Out: ports(vid("raw"))},
	}},
	{Group: "Logic", Items: []NodeType{
		{Type: "on_air_gate", Title: "On-air gate", Icon: "toggle-right", Cls: "logic", Category: "Logic", Media: "both", Subtitle: "Pass only the on-air client", In: ports(vid("any"), aud("any")), Out: ports(vid("any"), aud("any"))},
		{Type: "fallback", Title: "Fallback / slate", Icon: "image", Cls: "logic", Category: "Logic", Media: "video", Subtitle: "Show a slate when the source drops", In: ports(vid("raw")), Out: ports(vid("raw"))},
		{Type: "schedule", Title: "Schedule", Icon: "clock", Cls: "logic", Category: "Logic", Media: "data", Out: ports(data("trigger"))},
		{Type: "webhook", Title: "Webhook", Icon: "link", Cls: "logic", Category: "Logic", Media: "data", In: ports(data("event")), Out: ports(data("trigger"))},
	}},
	{Group: "Storage", Items: []NodeType{
		{Type: "file_in", Title: "Media file", Icon: "file", Cls: "src", Category: "Storage", Media: "both", Gst: "filesrc", Out: ports(vid("H.264"), aud("AAC"))},
		{Type: "image_in", Title: "Still image", Icon: "image", Cls: "src", Category: "Storage", Media: "video", Gst: "multifilesrc", Out: ports(vid("raw"))},
		{Type: "playlist", Title: "Playlist", Icon: "list", Cls: "src", Category: "Storage", Media: "both", Out: ports(vid("H.264"), aud("AAC"))},
	}},
	{Group: "Outputs", Items: []NodeType{
		{Type: "uvc_out", Title: "USB webcam (UVC)", Icon: "webcam", Cls: "sink", Category: "Outputs", Media: "video", Subtitle: "MJPEG / YUYV to the host", Dev: "/dev/video0", In: ports(vid("MJPEG"))},
		{Type: "uac2_out", Title: "USB mic (UAC2)", Icon: "usb", Cls: "sink", Category: "Outputs", Media: "audio", Gst: "alsasink", Subtitle: "48 kHz PCM to the host", Dev: "hw:UAC2Gadget", In: ports(aud("PCM"))},
		{Type: "rtmp_out", Title: "RTMP out", Icon: "upload-cloud", Cls: "sink", Category: "Outputs", Media: "both", Gst: "rtmpsink", URL: "rtmp://", In: ports(vid("H.264"), aud("AAC"))},
		{Type: "rtp_out", Title: "RTP out", Icon: "send", Cls: "sink", Category: "Outputs", Media: "both", Gst: "udpsink", URL: "udp://", In: ports(vid("H.264"), aud("Opus"))},
		{Type: "rtsp_out", Title: "RTSP server", Icon: "server", Cls: "sink", Category: "Outputs", Media: "both", URL: "rtsp://0.0.0.0:8554/live", In: ports(vid("H.264"), aud("AAC"))},
		{Type: "recorder", Title: "Recorder", Icon: "disc", Cls: "sink", Category: "Outputs", Media: "both", Gst: "splitmuxsink", Subtitle: "MP4 / MKV to storage", In: ports(vid("H.264"), aud("AAC"))},
	}},
}
