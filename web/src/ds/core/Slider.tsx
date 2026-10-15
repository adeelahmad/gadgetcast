import type { CSSProperties, ReactNode } from "react";

export interface SliderProps {
  value?: number;
  min?: number;
  max?: number;
  step?: number;
  onChange?: (value: number) => void;
  label?: ReactNode;
  unit?: string;
  /** Fill colour, a token reference. */
  tone?: string;
  style?: CSSProperties;
  "aria-label"?: string;
}

export function Slider({ value = 0, min = 0, max = 100, step = 1, onChange, label, unit, tone = "var(--brand)", style, ...aria }: SliderProps) {
  const pct = ((value - min) / (max - min)) * 100;
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--sp-3)", minWidth: 0, ...style }}>
      {label || unit ? (
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
          <span className="gc-caps">{label}</span>
          <span style={{ font: "var(--type-meta)", color: "var(--text-body)" }}>{value}{unit ? " " + unit : ""}</span>
        </div>
      ) : null}
      <div style={{ position: "relative", height: 16, display: "flex", alignItems: "center" }}>
        <div style={{ position: "absolute", left: 0, right: 0, height: 4, borderRadius: "var(--r-pill)", background: "var(--surface-input)", border: "1px solid var(--line-hairline)" }} />
        <div style={{ position: "absolute", left: 0, width: pct + "%", height: 4, borderRadius: "var(--r-pill)", background: tone }} />
        <div style={{ position: "absolute", left: "calc(" + pct + "% - 6px)", width: 12, height: 12, borderRadius: "var(--r-pill)", background: "var(--surface-raised)", border: "1.5px solid " + tone }} />
        <input type="range" value={value} min={min} max={max} step={step} aria-label={aria["aria-label"] || (typeof label === "string" ? label : undefined)}
          onChange={e => onChange?.(Number(e.target.value))}
          style={{ position: "absolute", inset: 0, width: "100%", opacity: 0, cursor: "pointer", margin: 0 }} />
      </div>
    </div>
  );
}
