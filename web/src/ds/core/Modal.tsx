import type { CSSProperties, ReactNode } from "react";
import { Icon } from "./Icon";
import { IconButton } from "./IconButton";

export interface ModalProps {
  open?: boolean;
  title?: ReactNode;
  eyebrow?: ReactNode;
  icon?: string;
  width?: number;
  onClose?: () => void;
  footer?: ReactNode;
  children?: ReactNode;
  style?: CSSProperties;
}

export function Modal({ open = true, title, eyebrow, icon, width = 520, onClose, footer, children, style }: ModalProps) {
  if (!open) return null;
  return (
    <div style={{ position: "absolute", inset: 0, background: "var(--scrim)", display: "flex", alignItems: "center", justifyContent: "center", zIndex: 60, padding: "var(--sp-9)" }}
      onClick={e => { if (e.target === e.currentTarget) onClose?.(); }}>
      <div role="dialog" aria-modal="true" aria-label={typeof title === "string" ? title : undefined}
        style={{
          width, maxWidth: "100%", maxHeight: "100%", display: "flex", flexDirection: "column", background: "var(--surface-panel)",
          border: "1px solid var(--line-border)", borderRadius: "var(--r-lg)", boxShadow: "var(--el-4)", overflow: "hidden", ...style,
        }}>
        <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-6) var(--sp-7)", borderBottom: "1px solid var(--line-hairline)" }}>
          {icon ? <Icon name={icon} size="md" style={{ color: "var(--brand)" }} /> : null}
          <div style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", gap: 1 }}>
            {eyebrow ? <span className="gc-caps">{eyebrow}</span> : null}
            {title ? <h3 style={{ font: "var(--type-heading)" }}>{title}</h3> : null}
          </div>
          <IconButton icon="close" label="Close" size="sm" onClick={onClose} />
        </div>
        <div className="gc-scroll" style={{ padding: "var(--sp-7)", overflow: "auto", display: "flex", flexDirection: "column", gap: "var(--sp-6)" }}>{children}</div>
        {footer ? (
          <div style={{ display: "flex", justifyContent: "flex-end", gap: "var(--sp-4)", padding: "var(--sp-6) var(--sp-7)", borderTop: "1px solid var(--line-hairline)", background: "var(--surface-app)" }}>{footer}</div>
        ) : null}
      </div>
    </div>
  );
}
