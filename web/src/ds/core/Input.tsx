import type { InputHTMLAttributes, ReactNode } from "react";
import { Icon } from "./Icon";

export type FieldSize = "sm" | "md" | "lg";
const H: Record<FieldSize, string> = { sm: "var(--ctl-sm)", md: "var(--ctl-md)", lg: "var(--ctl-lg)" };

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, "size"> {
  label?: ReactNode;
  hint?: ReactNode;
  icon?: string;
  suffix?: ReactNode;
  size?: FieldSize;
  /** Machine values (paths, URLs, IDs) render mono. */
  mono?: boolean;
  invalid?: boolean;
}

export function Input({ label, hint, icon, suffix, size = "md", mono = false, invalid = false, style, ...rest }: InputProps) {
  return (
    <label style={{ display: "flex", flexDirection: "column", gap: "var(--sp-2)", minWidth: 0 }}>
      {label ? <span className="gc-caps">{label}</span> : null}
      <span style={{ position: "relative", display: "flex", alignItems: "center" }}>
        {icon ? <Icon name={icon} size="sm" style={{ position: "absolute", left: "var(--sp-4)", color: "var(--text-faint)" }} /> : null}
        <input className="gc-field" data-invalid={invalid ? "true" : "false"} aria-invalid={invalid || undefined}
          style={{
            height: H[size], padding: "0 var(--sp-4) 0 " + (icon ? "26px" : "var(--sp-4)"),
            fontFamily: mono ? "var(--font-mono)" : "var(--font-sans)", fontSize: mono ? "var(--fs-meta)" : "var(--fs-ui)", ...style,
          }} {...rest} />
        {suffix ? (
          <span style={{ position: "absolute", right: "var(--sp-4)", font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", color: "var(--text-faint)", textTransform: "uppercase" }}>{suffix}</span>
        ) : null}
      </span>
      {hint ? <span style={{ font: "var(--type-meta)", color: invalid ? "var(--st-err)" : "var(--text-faint)" }}>{hint}</span> : null}
    </label>
  );
}
