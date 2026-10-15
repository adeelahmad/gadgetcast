import { useEffect, useRef, useState, type CSSProperties, type ReactNode } from "react";
import { Icon } from "./Icon";

export type TooltipTone = "neutral" | "video" | "audio" | "ai" | "live" | "warn" | "err";
const TONES: Record<TooltipTone, [string, string, string]> = {
  neutral: ["var(--surface-raised)", "var(--line-strong)", "var(--text-primary)"],
  video: ["var(--surface-raised)", "var(--sig-video-dim)", "var(--sig-video)"],
  audio: ["var(--surface-raised)", "var(--sig-audio-dim)", "var(--sig-audio)"],
  ai: ["var(--surface-raised)", "var(--sig-ai-dim)", "var(--sig-ai)"],
  live: ["var(--surface-raised)", "var(--sig-live-dim)", "var(--sig-live)"],
  warn: ["var(--surface-raised)", "var(--st-warn-dim)", "var(--st-warn)"],
  err: ["var(--surface-raised)", "var(--st-err-dim)", "var(--st-err)"],
};

export interface TooltipProps {
  label: ReactNode;
  detail?: ReactNode;
  shortcut?: string;
  icon?: string;
  side?: "top" | "bottom" | "left" | "right";
  tone?: TooltipTone;
  mono?: boolean;
  delay?: number;
  wrap?: boolean;
  children: ReactNode;
  style?: CSSProperties;
}

export function Tooltip({ label, detail, shortcut, icon, side = "top", tone = "neutral", mono = false, delay = 180, wrap = false, children, style }: TooltipProps) {
  const [on, setOn] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const t = TONES[tone];
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current); }, []);
  const enter = () => { if (timer.current) clearTimeout(timer.current); timer.current = setTimeout(() => setOn(true), delay); };
  const leave = () => { if (timer.current) clearTimeout(timer.current); setOn(false); };
  const pos: CSSProperties = side === "top" ? { bottom: "calc(100% + 6px)", left: "50%", transform: "translateX(-50%)" }
    : side === "bottom" ? { top: "calc(100% + 6px)", left: "50%", transform: "translateX(-50%)" }
      : side === "left" ? { right: "calc(100% + 6px)", top: "50%", transform: "translateY(-50%)" }
        : { left: "calc(100% + 6px)", top: "50%", transform: "translateY(-50%)" };
  return (
    <span style={{ position: "relative", display: "inline-flex", ...style }} onMouseEnter={enter} onMouseLeave={leave} onFocus={enter} onBlur={leave} onMouseDown={leave}>
      {children}
      {on ? (
        <span role="tooltip" style={{
          position: "absolute", zIndex: 80, display: "flex", alignItems: detail ? "flex-start" : "center", gap: "var(--sp-3)",
          padding: detail ? "var(--sp-3) var(--sp-4)" : "3px var(--sp-4)", background: t[0], border: "1px solid " + t[1], borderRadius: "var(--r-xs)",
          boxShadow: "var(--el-3)", whiteSpace: wrap ? "normal" : "nowrap", width: wrap ? 220 : "auto", flexDirection: detail ? "column" : "row", pointerEvents: "none", ...pos,
        }}>
          <span style={{ display: "flex", alignItems: "center", gap: "var(--sp-3)" }}>
            {icon ? <Icon name={icon} size="xs" color={t[2]} /> : null}
            <span style={{ font: mono ? "var(--type-meta)" : "var(--type-ui)", letterSpacing: mono ? "var(--ls-mono)" : "normal", color: t[2] }}>{label}</span>
            {shortcut && !detail ? <kbd style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", color: "var(--text-faint)" }}>{shortcut}</kbd> : null}
          </span>
          {detail ? <span style={{ font: "var(--type-meta)", letterSpacing: "var(--ls-mono)", color: "var(--text-muted)", whiteSpace: wrap ? "normal" : "nowrap" }}>{detail}</span> : null}
          {shortcut && detail ? <kbd style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", color: "var(--text-faint)" }}>{shortcut}</kbd> : null}
        </span>
      ) : null}
    </span>
  );
}
