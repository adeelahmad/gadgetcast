# WebRTC → USB AV Bridge
### Working codename: `gadgetcast`

**Status:** Concept / design draft
**Author:** Adeel Ahmad
**Date:** July 2026

---

## Install

**Easiest — flash a card.** Download `gadgetcast-*.img.xz` from the [latest release](https://github.com/adeelahmad/gadgetcast/releases/latest), open it in [Raspberry Pi Imager](https://www.raspberrypi.com/software/) ("Use custom"), set your Wi-Fi in the customisation screen, flash, boot. Then open **https://gadgetcast.local**. The pairing PIN is in `gadgetcast-pin.txt` on the card's boot partition.

**Existing Raspberry Pi OS (or any Linux box)** — one command, no prompts:

```sh
curl -fsSL https://raw.githubusercontent.com/adeelahmad/gadgetcast/master/install.sh | sh
```

It installs the latest release, creates a TLS certificate and pairing PIN (printed at the end), enables USB gadget mode on a Pi (reboot once, or add `-s -- --reboot` to do it automatically), starts the service and turns on daily automatic updates.

**Updates** happen on their own. To update right now, rerun the command above. Options: `GADGETCAST_NO_AUTOUPDATE=1` (no update timer), `GADGETCAST_VERSION=v0.1.0` (pin a release), `--uninstall`.

**Cutting a release** (maintainers): Actions → *Release* → *Run workflow* → type `v0.1.0`, or push a `v*` tag. Binaries, checksums, the installer and the Pi image are all published to the release automatically.

---

## 1. Concept

A Raspberry Pi that plugs into any computer over USB and presents itself as a **standard USB webcam and USB audio device** (microphone + speaker). The *source* of that video and audio is not a local sensor — it is a **WebRTC session** served by the Pi itself to any browser on the network.

Anyone on the LAN opens the Pi's web app, grants mic / camera / screen-share permission, and their media stream materialises on the host computer as a plug-and-play UVC webcam and UAC audio device. No drivers, no host-side software, no admin rights. The host sees exactly what it would see if you plugged in a Logitech webcam and a USB headset.

The host computer itself can *also* join the web app over the network at the same time — for example to loop its own screen (or a video playing in a browser tab) back into itself as a "webcam". This enables the killer party trick: feed any browser tab, any window, any remote participant's camera into Zoom/Teams/OBS as a native camera, on a machine where you're not allowed to install anything.

**One-line pitch:** *AirPlay for USB peripherals — your browser becomes a webcam and headset for any computer the Pi is plugged into.*

---

## 2. Why this doesn't already exist (prior art survey)

The individual pieces are all proven; the composition is the gap.

| Project | What it proves | What it lacks |
|---|---|---|
| **PiKVM** | Composite USB gadgets (HID + storage + more) on Pi 4 are rock-solid in production | Opposite direction — captures HDMI *into* the browser; doesn't emit UVC/UAC |
| **Showmewebcam** | Pi Zero + CSI camera as a zero-driver UVC webcam works beautifully | Source is a local camera, not the network |
| **uvc-gadget** (kernel + userspace) | Arbitrary frames can be pushed into a UVC function endpoint | It's a building block, not a product |
| **DroidCam / Iriun / Camo** | Huge demand for "phone as webcam" | All require host-side driver/app installs — the exact thing we're eliminating |
| **usbip** | USB over network | Needs client software and kernel modules on the host |
| **OBS Virtual Camera** | Loop anything as a webcam | Host software, admin install, per-OS quirks |

The differentiator is **zero host footprint**: the host only ever sees class-compliant USB devices. Corporate lockdown, ChromeOS, smart TVs, game consoles — anything with a USB host port and webcam support becomes a target.

---

## 3. Architecture

```
                 LAN / Wi-Fi / Tailscale
   ┌──────────────────────┬─────────────────────────┐
   │                      │                         │
┌──┴───────┐        ┌─────┴──────┐          ┌───────┴────────┐
│ Laptop A │        │  Phone B   │          │ Host computer  │
│ browser  │        │  browser   │          │ (also a client │
│ mic+cam  │        │ screenshare│          │  via browser)  │
└──────────┘        └────────────┘          └───────┬────────┘
      │  WebRTC (SRTP)      │                       │
      └──────────┬──────────┴───────────────────────┘
                 ▼
        ┌─────────────────────────────────────────┐
        │            RASPBERRY PI 4/5             │
        │                                         │
        │  ┌───────────────────────────────────┐  │
        │  │ Web app + signaling (WS over HTTPS)│ │
        │  └───────────────┬───────────────────┘  │
        │  ┌───────────────▼───────────────────┐  │
        │  │ WebRTC endpoint (Pion / GStreamer  │  │
        │  │ webrtcbin / aiortc)               │  │
        │  │  • DTLS-SRTP decrypt              │  │
        │  │  • H.264 depay → V4L2 M2M decode  │  │
        │  │  • Opus decode → PCM              │  │
        │  └───────┬───────────────┬───────────┘  │
        │          ▼               ▼              │
        │  ┌──────────────┐ ┌──────────────────┐  │
        │  │ Video convert│ │ Audio resample / │  │
        │  │ scale + MJPEG│ │ clock-drift comp │  │
        │  └──────┬───────┘ └───────┬──────────┘  │
        │         ▼                 ▼             │
        │  ┌─────────────────────────────────┐    │
        │  │  configfs composite USB gadget  │    │
        │  │   • usb_f_uvc   (webcam out)    │    │
        │  │   • usb_f_uac2  (mic + speaker) │    │
        │  │   • usb_f_ecm   (optional NIC)  │    │
        │  └───────────────┬─────────────────┘    │
        └──────────────────┼──────────────────────┘
                           │ USB-C (dwc2, USB 2.0 HS)
                           ▼
                  ┌─────────────────┐
                  │  HOST COMPUTER   │
                  │ sees: UVC webcam │
                  │  + USB mic       │
                  │  + USB speaker   │
                  └─────────────────┘
```

### 3.1 The USB gadget side (the "physical device" illusion)

The Pi 4/5 USB-C port (and Pi Zero 2 W micro-USB) runs the `dwc2` controller in peripheral mode. Via **configfs** we build one composite gadget with multiple functions — the same mechanism PiKVM uses for keyboard + mouse + mass storage, so the reliability story is already proven on this exact silicon.

**Video — `usb_f_uvc`.** The UVC function exposes a V4L2 *output* device on the Pi (`/dev/videoN`); whatever frames we write to it appear on the host as webcam frames. Descriptor advertises MJPEG (primary) and YUYV (fallback) at 1080p30 / 720p30 / 480p30. The userspace `uvc-gadget` daemon (or our own writer) handles the UVC streaming requests.

**Audio — `usb_f_uac2`.** UAC2 is inherently **bidirectional**: it exposes an ALSA card on the Pi where one PCM direction is "what the host hears as its microphone" (we write to it) and the other is "what the host plays to its speaker" (we read from it). That second direction is a free gift: host speaker audio can be looped back **into** the WebRTC session, giving remote browser participants full-duplex audio with the host. The device is an intercom, not just an injector.

**Networking bonus — `usb_f_ecm`/`ncm` (optional, off by default).** The gadget can additionally present a USB network interface, so a host with *no* LAN access can still reach the Pi's web app through the USB cable itself. Powerful, but new NICs raise flags on managed corporate machines — hence opt-in.

### 3.2 The WebRTC side

The Pi runs three cooperating services:

1. **Web app + signaling.** Static SPA plus a WebSocket signaling channel. Must be served over **HTTPS** — `getUserMedia`/`getDisplayMedia` refuse to run in insecure contexts. Options: self-signed with a first-run trust step, mkcert-style local CA, or (cleanest on your network) Tailscale HTTPS certs. mDNS advertisement (`gadgetcast.local`) for discovery.
2. **WebRTC endpoint.** Not an SFU in the classic sense — the Pi is the *terminal consumer* of the media. Candidates:
   - **GStreamer `webrtcbin`** — best fit; the same pipeline that terminates WebRTC can decode, convert, and write to the V4L2/ALSA gadget devices with zero glue code.
   - **Pion (Go)** — clean, static binary, great control over SDP; pair with GStreamer via appsrc/appsink.
   - **aiortc (Python)** — fastest to prototype, weakest on ARM CPU efficiency.
3. **Media pipeline.** Force H.264 in SDP negotiation so the Pi 4's V4L2 M2M hardware decoder (`/dev/video10`) carries the video load (Pi 5 has no H.264 hw decode block but its CPU software-decodes 1080p30 with headroom). Decoded frames are scaled and re-encoded to MJPEG for the gadget. Opus decodes to 48 kHz PCM and is written to the UAC2 ALSA device through an adaptive resampler.

### 3.3 The switcher

Multiple browser clients can be connected simultaneously; exactly one stream is "on air" to the USB gadget at a time. The web app includes a source-switcher UI (think tiny vision mixer): tiles of connected participants, tap to route one to the host. Later: picture-in-picture compositing via GStreamer `compositor`.

---

## 4. Key constraints and how we handle them

**USB 2.0 bandwidth.** `dwc2` is High-Speed only — 480 Mbps theoretical, ~300 usable for isochronous. Raw 1080p30 YUYV needs ~1.5 Gbps, so raw is capped around 480p30. **MJPEG is therefore the primary UVC format** (1080p30 MJPEG ≈ 30–60 Mbps, trivial). Every mainstream OS and app accepts MJPEG webcams; most real webcams ship MJPEG-first for the same reason.

**Transcode cost.** The awkward truth of the pipeline is H.264-in → MJPEG-out. Pi 4: hardware decode + hardware JPEG encode (`v4l2jpegenc`) keeps CPU nearly idle. Pi 5: software decode + software/ISP JPEG — measured fine at 1080p30 but worth benchmarking early. This is the project's main performance risk; prototype it first.

**Latency budget.** Browser encode (~30 ms) + network (~5 ms LAN) + jitter buffer (30–50 ms) + decode (~10 ms) + MJPEG encode (~10 ms) + UVC/host buffering (~30 ms) ⇒ **~120–180 ms glass-to-glass**. Fine for meetings and streaming; not for lip-sync-critical live performance.

**Audio clock drift.** The browser's 48 kHz and the USB host's 48 kHz are different crystals. Without compensation you get a buffer under/overrun click every few minutes. Solution: async resampler (`libsamplerate`/ALSA `plug` with drift tracking) between WebRTC output and the UAC2 device — a solved problem, but it must be in the design from day one.

**Security.** This device is, functionally, a network-controlled microphone and camera injector for the host. Baseline: HTTPS everywhere, PIN or QR pairing on first connect per client, on-air indicator in the web UI, and an option to bind the web app to the Tailscale interface only. An unauthenticated `gadgetcast` on an open LAN would be a genuinely bad time.

---

## 5. Hardware

| Option | Gadget port | Notes |
|---|---|---|
| **Pi 4B** (recommended) | USB-C (power+data via dwc2) | H.264 hw decode + hw JPEG encode = ideal pipeline. Power via GPIO or a data+power USB-C splitter, since the gadget port is also the power port |
| **Pi 5** | USB-C | Fastest CPU, but software H.264 decode; verify thermals under sustained load |
| **Pi Zero 2 W** | micro-USB OTG | The $15 dream form factor — powered *by the host port*, truly "a USB stick that is a webcam". Tight on CPU; 720p target |
| **CM4 + carrier** | Per carrier | The productised version: PoE, proper case, host-powered |

---

## 6. Use cases

1. **Locked-down corporate machine, better AV.** No install rights? Plug in, share your phone's camera or a good mic from another device. The host just sees a webcam.
2. **The long-meeting loop.** Play a video (or a pre-recorded loop of yourself) in any browser tab, screen-share that tab into the Pi, and it becomes the host's camera. Present in the meeting; free in real life.
3. **Universal OBS/streaming ingest.** Any network source becomes a native camera on the streaming rig — no NDI plugins, no capture cards for browser-based sources.
4. **Remote guest injection.** A remote participant joins the Pi's web page from anywhere (Tailscale) and appears as a local webcam in a videoconference on the host — a poor-man's remote-guest frame-sync box.
5. **Conference-room brain transplant.** Old room system with a USB codec bar? The Pi replaces its camera/mic with anything on the network.
6. **KVM-adjacent workflows.** Natural sibling to PiKVM: PiKVM takes the host's output to your browser; gadgetcast takes your browser to the host's input.

---

## 7. MVP roadmap

**Phase 0 — Feasibility spike (a weekend).** configfs composite gadget with UVC+UAC2 on a Pi 4; static test pattern via `uvc-gadget`; `speaker-test` PCM into UAC2. Confirm host enumeration on macOS + Windows + Linux simultaneously with both functions active (composite descriptor quirks are the classic trap — steal PiKVM's descriptor choices where possible).

**Phase 1 — Audio only, end to end.** Web page (HTTPS + WS signaling) → browser mic → Opus → Pi → resampler → UAC2. "Speak into your phone, host hears a USB microphone." Ships value on its own (universal wireless mic).

**Phase 2 — Video.** getDisplayMedia/getUserMedia → H.264 (SDP-pinned) → hw decode → MJPEG → UVC. 720p30 first, then 1080p30.

**Phase 3 — Duplex + multi-client.** UAC2 host-speaker capture looped back into WebRTC (echo cancellation via browser AEC); multi-client switcher UI; on-air states.

**Phase 4 — Product polish.** mDNS discovery, pairing flow, Tailscale integration, prebuilt image (pi-gen, automated by the release workflow), optional USB-NIC mode, CM4 carrier design.

---

## 8. Open questions

- GStreamer-only vs Pion+GStreamer split — prototype both in Phase 1 and pick on maintainability.
- Can we skip the MJPEG re-encode for hosts that accept H.264 UVC payloads (UVC 1.5)? Support is patchy (Windows yes-ish, macOS mostly no) — probably a later optimisation, but it would collapse the transcode cost to zero.
- Pi Zero 2 W viability at 720p — pure software path may not fit; may need MJPEG from the *browser side* (`MediaRecorder` hack) as a low-power mode.
- Multi-gadget: can one Pi present *two* webcams (two UVC functions) for A/B sources? configfs allows it; host OS support needs testing.
