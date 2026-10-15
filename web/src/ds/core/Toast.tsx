import type { CSSProperties, ReactNode } from "react";
import { Icon } from "./Icon";

export type ToastTone = "neutral" | "ok" | "warn" | "err" | "live" | "video" | "audio" | "ai";
const TONES: Record<ToastTone, [string, string, string]> = {
  neutral: ["var(--text-muted)", "var(--line-strong)", "info"],
  ok: ["var(--st-ok)", "var(--st-ok-dim)", "check_circle"],
  warn: ["var(--st-warn)", "var(--st-warn-dim)", "warning"],
  err: ["var(--st-err)", "var(--st-err-dim)", "error"],
  live: ["var(--sig-live)", "var(--sig-live-dim)", "fiber_manual_record"],
  video: ["var(--sig-video)", "var(--sig-video-dim)", "videocam"],
  audio: ["var(--sig-audio)", "var(--sig-audio-dim)", "graphic_eq"],
  ai: ["var(--sig-ai)", "var(--sig-ai-dim)", "auto_awesome"],
};

export interface ToastProps {
  children: ReactNode;
  tone?: ToastTone;
  icon?: string;
  detail?: ReactNode;
  actionLabel?: string;
  onAction?: () => void;
  onDismiss?: () => void;
  style?: CSSProperties;
}

/** One line, pill, app-positioned and app-timed. Anything needing a decision is a Modal. */
export function Toast({ children, tone = "neutral", icon, detail, actionLabel, onAction, onDismiss, style }: ToastProps) {
  const t = TONES[tone];
  return (
    <div role="status" className="gc-toast" style={{
      display: "inline-flex", alignItems: "center", gap: "var(--sp-4)", height: "var(--ctl-lg)", padding: "0 var(--sp-3) 0 var(--sp-5)",
      background: "var(--surface-raised)", border: "1px solid " + t[1], borderRadius: "var(--r-pill)", maxWidth: 520, boxShadow: "var(--el-3)", ...style,
    }}>
      <Icon name={icon || t[2]} size="sm" color={t[0]} />
      <span style={{ font: "var(--type-ui)", color: "var(--text-primary)", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis", paddingRight: onDismiss || actionLabel ? 0 : "var(--sp-3)" }}>{children}</span>
      {detail ? <span style={{ font: "var(--type-meta)", letterSpacing: "var(--ls-mono)", color: "var(--text-muted)", whiteSpace: "nowrap", flex: "0 0 auto" }}>{detail}</span> : null}
      {actionLabel ? (
        <button type="button" className="gc-toast-action" onClick={onAction}
          style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: t[0], background: "transparent", border: "1px solid " + t[1], borderRadius: "var(--r-pill)", height: "var(--ctl-xs)", padding: "0 var(--sp-4)", cursor: "pointer", flex: "0 0 auto" }}>{actionLabel}</button>
      ) : null}
      {onDismiss ? (
        <button type="button" aria-label="Dismiss" className="gc-iconbtn" onClick={onDismiss}
          style={{ width: "var(--ctl-sm)", height: "var(--ctl-sm)", borderRadius: "var(--r-pill)", flex: "0 0 auto" }}><Icon name="close" size="sm" /></button>
      ) : null}
    </div>
  );
}
