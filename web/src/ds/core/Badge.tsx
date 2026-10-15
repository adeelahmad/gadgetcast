import type { HTMLAttributes, ReactNode } from "react";
import { Icon } from "./Icon";

export type BadgeTone = "neutral" | "video" | "audio" | "ai" | "data" | "live" | "ok" | "warn" | "err";
/* [hue, dim border]; the fill is the hue at 12% so every tint resolves through tokens in both themes. */
const TONES: Record<BadgeTone, [string, string]> = {
  neutral: ["var(--text-muted)", "var(--line-border)"],
  video: ["var(--sig-video)", "var(--sig-video-dim)"],
  audio: ["var(--sig-audio)", "var(--sig-audio-dim)"],
  ai: ["var(--sig-ai)", "var(--sig-ai-dim)"],
  data: ["var(--sig-data)", "var(--sig-data-dim)"],
  live: ["var(--sig-live)", "var(--sig-live-dim)"],
  ok: ["var(--st-ok)", "var(--st-ok-dim)"],
  warn: ["var(--st-warn)", "var(--st-warn-dim)"],
  err: ["var(--st-err)", "var(--st-err-dim)"],
};

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: BadgeTone;
  icon?: string;
  dot?: boolean;
  solid?: boolean;
  children?: ReactNode;
}

export function Badge({ children, tone = "neutral", icon, dot = false, solid = false, style, ...rest }: BadgeProps) {
  const [hue, dim] = TONES[tone];
  const fill = tone === "neutral" ? "var(--surface-hover)" : "color-mix(in srgb, " + hue + " 12%, transparent)";
  return (
    <span style={{
      display: "inline-flex", alignItems: "center", gap: "var(--sp-2)", height: "var(--ctl-xs)", padding: "0 var(--sp-3)",
      borderRadius: "var(--r-xs)", background: solid ? hue : fill, border: "1px solid " + (solid ? "transparent" : dim),
      color: solid ? "var(--ink-1000)" : hue, font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase",
      whiteSpace: "nowrap", flex: "0 0 auto", ...style,
    }} {...rest}>
      {dot ? <span style={{ width: 5, height: 5, borderRadius: "var(--r-pill)", background: "currentColor", flex: "0 0 auto" }} /> : null}
      {icon ? <Icon name={icon} size="xs" /> : null}
      {children}
    </span>
  );
}
