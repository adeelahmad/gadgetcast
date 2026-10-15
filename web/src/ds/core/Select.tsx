import type { ReactNode, SelectHTMLAttributes } from "react";
import { Icon } from "./Icon";
import type { FieldSize } from "./Input";

const H: Record<FieldSize, string> = { sm: "var(--ctl-sm)", md: "var(--ctl-md)", lg: "var(--ctl-lg)" };
export type SelectOption = string | { value: string; label: string };

export interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, "size"> {
  label?: ReactNode;
  options: SelectOption[];
  size?: FieldSize;
  hint?: ReactNode;
}

export function Select({ label, options, size = "md", hint, style, ...rest }: SelectProps) {
  return (
    <label style={{ display: "flex", flexDirection: "column", gap: "var(--sp-2)", minWidth: 0 }}>
      {label ? <span className="gc-caps">{label}</span> : null}
      <span style={{ position: "relative", display: "flex", alignItems: "center" }}>
        <select className="gc-field" style={{ height: H[size], padding: "0 24px 0 var(--sp-4)", fontSize: "var(--fs-ui)", appearance: "none", cursor: "pointer", ...style }} {...rest}>
          {options.map(o => {
            const v = typeof o === "string" ? o : o.value;
            const l = typeof o === "string" ? o : o.label;
            return <option key={v} value={v}>{l}</option>;
          })}
        </select>
        <Icon name="expand_more" size="sm" style={{ position: "absolute", right: "var(--sp-3)", color: "var(--text-faint)", pointerEvents: "none" }} />
      </span>
      {hint ? <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>{hint}</span> : null}
    </label>
  );
}
