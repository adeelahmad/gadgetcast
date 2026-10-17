import { useEffect, useRef, type CSSProperties, type ReactNode } from "react";
import { Icon } from "../core/Icon";

/* ── LevelMeter ── green under -6 dBFS, amber approaching 0, red clipping. Flat fills, no gradient. */
export interface LevelMeterProps {
  channels?: number[];
  peaks?: number[];
  orientation?: "horizontal" | "vertical";
  showScale?: boolean;
  height?: number;
  width?: number;
  label?: ReactNode;
  style?: CSSProperties;
}

export function LevelMeter({ channels = [0, 0], peaks, orientation = "horizontal", showScale = true, height, width, label, style }: LevelMeterProps) {
  const vertical = orientation === "vertical";
  const seg = (v: number) => (v > 0.92 ? "var(--st-err)" : v > 0.75 ? "var(--st-warn)" : "var(--st-ok)");
  const barLen = vertical ? height || 64 : width || 120;
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-2)", ...style }}>
      {label ? <span className="gc-caps">{label}</span> : null}
      <div style={{ display: "flex", flexDirection: vertical ? "row" : "column", gap: "var(--sp-2)" }}>
        {channels.map((v, i) => {
          const pk = peaks && peaks[i] != null ? peaks[i] : v;
          return (
            <div key={i} style={{ position: "relative", background: "var(--surface-input)", border: "1px solid var(--line-hairline)", borderRadius: "var(--r-xs)", width: vertical ? 6 : barLen, height: vertical ? barLen : 5, overflow: "hidden" }}>
              <div style={{ position: "absolute", left: 0, bottom: 0, width: vertical ? "100%" : v * 100 + "%", height: vertical ? v * 100 + "%" : "100%", background: seg(v), transition: "all 90ms linear" }} />
              {pk > 0 ? <div style={{ position: "absolute", background: "var(--text-muted)", left: vertical ? 0 : pk * 100 + "%", bottom: vertical ? pk * 100 + "%" : 0, width: vertical ? "100%" : 2, height: vertical ? 2 : "100%" }} /> : null}
            </div>
          );
        })}
      </div>
      {showScale ? (
        <div style={{ display: "flex", justifyContent: "space-between", width: vertical ? undefined : barLen, font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)" }}>
          <span>-60</span><span>-24</span><span>-6</span><span>0</span>
        </div>
      ) : null}
    </div>
  );
}

/* ── WaveScope ── canvas oscilloscope / spectrum. `live` animates; otherwise draws one frame (flat line when gain is 0). */
export interface WaveScopeProps {
  mode?: "wave" | "spectrum";
  tone?: string;
  height?: number;
  live?: boolean;
  gain?: number;
  grid?: boolean;
  label?: ReactNode;
  style?: CSSProperties;
}

export function WaveScope({ mode = "wave", tone = "var(--sig-audio)", height = 64, live = true, gain = 1, grid = true, label, style }: WaveScopeProps) {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const cv = ref.current; if (!cv) return undefined;
    const ctx = cv.getContext("2d"); if (!ctx) return undefined;
    let raf = 0, t = 0, stop = false;
    const bars = 48;
    const draw = () => {
      const w = (cv.width = cv.clientWidth * 2), h = (cv.height = cv.clientHeight * 2);
      ctx.clearRect(0, 0, w, h);
      const cs = getComputedStyle(cv);
      if (grid) {
        ctx.strokeStyle = cs.getPropertyValue("--line-hairline").trim() || cs.color; ctx.lineWidth = 2;
        for (let gx = 0; gx < 5; gx++) { const x = (w / 4) * gx; ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, h); ctx.stroke(); }
        ctx.beginPath(); ctx.moveTo(0, h / 2); ctx.lineTo(w, h / 2); ctx.stroke();
      }
      ctx.strokeStyle = cs.color; ctx.fillStyle = cs.color; ctx.lineWidth = 3; ctx.lineJoin = "round";
      if (mode === "spectrum") {
        for (let i = 0; i < bars; i++) {
          let amp = (Math.sin(t * 0.05 + i * 0.35) * 0.4 + Math.sin(t * 0.11 + i * 0.9) * 0.3 + 0.5) * gain;
          amp = Math.max(gain ? 0.04 : 0, Math.min(1, Math.abs(amp)));
          const bw = (w / bars) * 0.62, bx = i * (w / bars);
          ctx.globalAlpha = 0.35 + amp * 0.65;
          ctx.fillRect(bx, h - amp * h, bw, amp * h);
        }
        ctx.globalAlpha = 1;
      } else {
        ctx.beginPath();
        for (let px = 0; px < w; px += 3) {
          const ph = px * 0.012 + t * 0.06;
          const y = h / 2 + Math.sin(ph) * h * 0.18 * gain * (0.6 + 0.4 * Math.sin(ph * 0.21 + t * 0.02)) + Math.sin(ph * 3.1) * h * 0.05 * gain;
          if (px === 0) ctx.moveTo(px, y); else ctx.lineTo(px, y);
        }
        ctx.stroke();
      }
      if (live && !stop) { t += 1; raf = requestAnimationFrame(draw); }
    };
    draw();
    return () => { stop = true; cancelAnimationFrame(raf); };
  }, [mode, live, gain, grid]);
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-2)", minWidth: 0, ...style }}>
      {label ? <span className="gc-caps">{label}</span> : null}
      <div style={{ position: "relative", background: "var(--surface-canvas)", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)", overflow: "hidden", height }}>
        <canvas ref={ref} style={{ width: "100%", height: "100%", display: "block", color: tone }} />
      </div>
    </div>
  );
}

/* ── OnAirIndicator ── the tally. Never hide that media is flowing. */
export interface OnAirIndicatorProps { state?: "off" | "armed" | "live"; label?: string; source?: string; compact?: boolean; style?: CSSProperties }

export function OnAirIndicator({ state = "off", label, source, compact = false, style }: OnAirIndicatorProps) {
  const on = state === "live", arm = state === "armed";
  const hue = on ? "var(--sig-live)" : arm ? "var(--st-warn)" : "var(--st-idle)";
  const text = label || (on ? "On air" : arm ? "Armed" : "Off air");
  return (
    <div className={on ? "gc-onair" : undefined} data-testid="tally" data-state={state}
      style={{
        display: "inline-flex", alignItems: "center", gap: "var(--sp-3)", height: compact ? "var(--ctl-sm)" : "var(--ctl-md)", padding: "0 var(--sp-4)", borderRadius: "var(--r-sm)", flex: "0 0 auto",
        background: on ? "color-mix(in srgb, var(--sig-live) 8%, transparent)" : arm ? "color-mix(in srgb, var(--st-warn) 8%, transparent)" : "var(--surface-raised)",
        border: "1px solid " + (on ? "var(--sig-live-dim)" : arm ? "var(--st-warn-dim)" : "var(--line-border)"), ...style,
      }}>
      <span className={on ? "gc-pulse" : undefined} style={{ width: 7, height: 7, borderRadius: "var(--r-pill)", background: hue, flex: "0 0 auto" }} />
      <span style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: on || arm ? hue : "var(--text-muted)", whiteSpace: "nowrap" }}>{text}</span>
      {source && !compact ? <span style={{ font: "var(--type-meta)", color: "var(--text-muted)", borderLeft: "1px solid var(--line-hairline)", paddingLeft: "var(--sp-3)", whiteSpace: "nowrap", maxWidth: 160, overflow: "hidden", textOverflow: "ellipsis" }}>{source}</span> : null}
      {on ? <Icon name="usb" size="xs" style={{ color: hue }} /> : null}
    </div>
  );
}

/* ── MetricPill ── a mono readout with its unit attached. */
export interface MetricPillProps { icon?: string; label?: string; value: ReactNode; unit?: string; tone?: string; trend?: "up" | "down" | "flat"; style?: CSSProperties }

export function MetricPill({ icon, label, value, unit, tone, trend, style }: MetricPillProps) {
  return (
    <div style={{ display: "inline-flex", alignItems: "center", gap: "var(--sp-3)", height: "var(--ctl-sm)", padding: "0 var(--sp-4)", background: "var(--surface-raised)", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)", flex: "0 0 auto", whiteSpace: "nowrap", ...style }}>
      {icon ? <Icon name={icon} size="xs" style={{ color: tone || "var(--text-faint)" }} /> : null}
      {label ? <span style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: "var(--text-faint)" }}>{label}</span> : null}
      <span style={{ font: "var(--type-meta)", color: tone || "var(--text-primary)" }}>{value}{unit ? <span style={{ color: "var(--text-faint)" }}> {unit}</span> : null}</span>
      {trend ? <Icon name={trend === "up" ? "trending_up" : trend === "down" ? "trending_down" : "remove"} size="xs" style={{ color: trend === "up" ? "var(--st-warn)" : trend === "down" ? "var(--st-ok)" : "var(--text-faint)" }} /> : null}
    </div>
  );
}
