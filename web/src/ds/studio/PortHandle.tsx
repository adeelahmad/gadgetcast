import type { CSSProperties, PointerEvent, ReactNode } from "react";

export type SignalKind = "video" | "audio" | "ai" | "data";
export const SIGNAL_HUE: Record<SignalKind, string> = { video: "var(--sig-video)", audio: "var(--sig-audio)", ai: "var(--sig-ai)", data: "var(--sig-data)" };

export interface PortHandleProps {
  kind?: SignalKind;
  direction?: "in" | "out";
  label?: string;
  index?: number;
  connected?: boolean;
  snapped?: boolean;
  /** true while a compatible wire is dragged (grows + halo), false while an incompatible one is (fades). */
  hot?: boolean | null;
  invalid?: boolean;
  onPortDown?: (e: PointerEvent<HTMLSpanElement>) => void;
  /** Rendered inside the 20px hit area: the seam where the canvas drops a React Flow <Handle>. */
  children?: ReactNode;
  style?: CSSProperties;
}

/* 20px hit area, 10px dot. Data ports are square. The dot's centre sits on the node edge via -10px margins. */
export function PortHandle({ kind = "video", direction = "in", label, index = 0, connected = false, snapped = false, hot = null, invalid = false, onPortDown, children, style }: PortHandleProps) {
  const hue = invalid ? "var(--st-err)" : SIGNAL_HUE[kind] || SIGNAL_HUE.data;
  const tint = "color-mix(in srgb, " + hue + " 18%, transparent)";
  const dim = hot === false;
  const big = snapped || hot === true;
  return (
    <span style={{ display: "flex", alignItems: "center", gap: 6, flexDirection: direction === "out" ? "row-reverse" : "row", opacity: dim ? 0.3 : 1, transition: "opacity 120ms", ...style }}>
      <span className="gc-port" data-snapped={snapped ? "true" : "false"} data-kind={kind} data-direction={direction} data-port={index}
        onPointerDown={onPortDown}
        style={{ position: "relative", width: 20, height: 20, flex: "0 0 auto", display: "grid", placeItems: "center", cursor: "crosshair" }}>
        {big ? <span style={{ position: "absolute", inset: 0, borderRadius: "50%", background: tint }} /> : null}
        <span style={{
          width: big ? 12 : 10, height: big ? 12 : 10, borderRadius: kind === "data" ? 3 : "50%", position: "relative",
          background: connected || big ? hue : "var(--surface-card)", border: "2px solid " + hue, transition: "width 100ms,height 100ms,background-color 100ms",
        }} />
        {children}
      </span>
      {label ? <span style={{ font: "var(--fw-medium) var(--fs-meta)/1 var(--font-sans)", letterSpacing: ".01em", color: connected ? "var(--text-muted)" : "var(--text-faint)", whiteSpace: "nowrap" }}>{label}</span> : null}
    </span>
  );
}
