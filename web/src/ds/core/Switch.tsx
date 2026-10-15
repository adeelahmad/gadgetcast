import type { CSSProperties, ReactNode } from "react";

export interface SwitchProps {
  checked?: boolean;
  onChange?: (checked: boolean) => void;
  label?: ReactNode;
  description?: ReactNode;
  size?: "sm" | "md";
  disabled?: boolean;
  style?: CSSProperties;
}

export function Switch({ checked = false, onChange, label, description, size = "md", disabled = false, style }: SwitchProps) {
  const w = size === "sm" ? 26 : 32, h = size === "sm" ? 14 : 18, k = h - 4;
  return (
    <label style={{ display: "flex", alignItems: description ? "flex-start" : "center", gap: "var(--sp-4)", cursor: disabled ? "not-allowed" : "pointer", opacity: disabled ? 0.5 : 1, ...style }}>
      <button type="button" role="switch" aria-checked={checked} disabled={disabled}
        onClick={() => onChange?.(!checked)}
        style={{
          width: w, height: h, flex: "0 0 auto", borderRadius: "var(--r-pill)", border: "1px solid " + (checked ? "transparent" : "var(--line-strong)"),
          background: checked ? "var(--brand)" : "var(--surface-input)", boxShadow: checked ? "none" : "var(--el-inset)", padding: 0, cursor: "inherit", position: "relative",
          transition: "background-color var(--dur-fast) var(--ease-out)",
        }}>
        <span style={{
          position: "absolute", top: 1, left: checked ? w - k - 3 : 2, width: k, height: k, borderRadius: "var(--r-pill)",
          background: checked ? "var(--brand-ink)" : "var(--ink-300)", transition: "left var(--dur-fast) var(--ease-snap)",
        }} />
      </button>
      {label || description ? (
        <span style={{ display: "flex", flexDirection: "column", gap: 1, minWidth: 0 }}>
          {label ? <span style={{ font: "var(--type-ui)", color: "var(--text-primary)" }}>{label}</span> : null}
          {description ? <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>{description}</span> : null}
        </span>
      ) : null}
    </label>
  );
}
