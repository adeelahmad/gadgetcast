import type { CSSProperties } from "react";
import { Icon } from "./Icon";

export interface TabItem<V extends string = string> { value: V; label: string; icon?: string; count?: number }

export interface TabsProps<V extends string> {
  items: TabItem<V>[];
  value: V;
  onChange?: (value: V) => void;
  size?: "sm" | "md";
  style?: CSSProperties;
}

export function Tabs<V extends string>({ items, value, onChange, size = "md", style }: TabsProps<V>) {
  const h = size === "sm" ? "var(--ctl-sm)" : "var(--ctl-lg)";
  return (
    <div role="tablist" style={{ display: "flex", alignItems: "stretch", gap: "var(--sp-1)", borderBottom: "1px solid var(--line-hairline)", ...style }}>
      {items.map(it => {
        const sel = it.value === value;
        return (
          <button key={it.value} type="button" role="tab" aria-selected={sel} className="gc-tab" data-selected={sel ? "true" : "false"}
            onClick={() => onChange?.(it.value)}
            style={{ height: h, display: "flex", alignItems: "center", gap: "var(--sp-3)", padding: "0 var(--sp-5)", font: "var(--type-ui)", whiteSpace: "nowrap" }}>
            {it.icon ? <Icon name={it.icon} size="sm" /> : null}
            {it.label}
            {it.count != null ? <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)" }}>{it.count}</span> : null}
          </button>
        );
      })}
    </div>
  );
}
