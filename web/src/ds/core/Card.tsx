import type { HTMLAttributes, ReactNode } from "react";
import { Icon } from "./Icon";

export interface CardProps extends Omit<HTMLAttributes<HTMLElement>, "title"> {
  title?: ReactNode;
  eyebrow?: ReactNode;
  icon?: string;
  actions?: ReactNode;
  /** Always mono meta. */
  footer?: ReactNode;
  padding?: string | number;
  /** Border / icon tone, a token reference. */
  tone?: string;
}

export function Card({ title, eyebrow, icon, actions, footer, padding = "var(--sp-6)", tone, children, style, ...rest }: CardProps) {
  return (
    <section style={{
      background: "var(--surface-card)", border: "1px solid " + (tone || "var(--line-border)"), borderRadius: "var(--r-md)",
      boxShadow: "var(--el-2)", display: "flex", flexDirection: "column", minWidth: 0, overflow: "hidden", ...style,
    }} {...rest}>
      {title || eyebrow || actions ? (
        <header style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-5) var(--sp-6)", borderBottom: "1px solid var(--line-hairline)" }}>
          {icon ? <Icon name={icon} size="sm" style={{ color: tone || "var(--text-faint)" }} /> : null}
          <div style={{ display: "flex", flexDirection: "column", gap: 1, minWidth: 0, flex: 1 }}>
            {eyebrow ? <span className="gc-caps">{eyebrow}</span> : null}
            {title ? <h3 style={{ font: "var(--type-heading)" }}>{title}</h3> : null}
          </div>
          {actions ? <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-2)" }}>{actions}</div> : null}
        </header>
      ) : null}
      <div style={{ padding, minWidth: 0 }}>{children}</div>
      {footer ? (
        <footer style={{ padding: "var(--sp-4) var(--sp-6)", borderTop: "1px solid var(--line-hairline)", background: "var(--surface-app)", font: "var(--type-meta)", color: "var(--text-faint)" }}>{footer}</footer>
      ) : null}
    </section>
  );
}
