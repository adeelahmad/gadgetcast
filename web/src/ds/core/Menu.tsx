import { useEffect, useRef, useState, type CSSProperties, type ReactNode } from "react";
import { Icon } from "./Icon";

export type MenuItem =
  | { divider: true }
  | { header: string }
  | { value: string; label: string; icon?: string; shortcut?: string; selected?: boolean; danger?: boolean; disabled?: boolean; tone?: string };

export interface MenuProps {
  items: MenuItem[];
  trigger: ReactNode;
  align?: "start" | "end";
  width?: number;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSelect?: (value: string) => void;
  style?: CSSProperties;
}

export function Menu({ items, trigger, align = "start", width = 224, open, onOpenChange, onSelect, style }: MenuProps) {
  const [self, setSelf] = useState(false);
  const isOpen = open === undefined ? self : open;
  const ref = useRef<HTMLSpanElement>(null);
  const setRef = useRef<(v: boolean) => void>(() => {});
  setRef.current = (v: boolean) => { if (open === undefined) setSelf(v); onOpenChange?.(v); };
  useEffect(() => {
    if (!isOpen) return undefined;
    const away = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) setRef.current(false); };
    const key = (e: KeyboardEvent) => { if (e.key === "Escape") setRef.current(false); };
    document.addEventListener("mousedown", away); document.addEventListener("keydown", key);
    return () => { document.removeEventListener("mousedown", away); document.removeEventListener("keydown", key); };
  }, [isOpen]);
  return (
    <span ref={ref} style={{ position: "relative", display: "inline-flex", ...style }}>
      <span onClick={() => setRef.current(!isOpen)} style={{ display: "inline-flex" }}>{trigger}</span>
      {isOpen ? (
        <div role="menu" style={{
          position: "absolute", top: "calc(100% + 4px)", zIndex: 90, width, left: align === "start" ? 0 : "auto", right: align === "end" ? 0 : "auto",
          background: "var(--surface-raised)", border: "1px solid var(--line-strong)", borderRadius: "var(--r-lg)", boxShadow: "var(--el-3)",
          padding: "var(--sp-2)", display: "flex", flexDirection: "column", gap: "var(--sp-1)",
        }}>
          {items.map((it, i) => {
            if ("divider" in it) return <span key={i} style={{ height: 1, background: "var(--line-hairline)", margin: "var(--sp-2) var(--sp-3)" }} />;
            if ("header" in it) return <span key={i} style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: "var(--text-faint)", padding: "var(--sp-3) var(--sp-4) var(--sp-2)" }}>{it.header}</span>;
            const tone = it.danger ? "var(--st-err)" : (it.tone || "var(--text-body)");
            return (
              <button key={it.value} type="button" role="menuitem" className="gc-menuitem" disabled={it.disabled}
                data-selected={it.selected ? "true" : "false"} data-danger={it.danger ? "true" : "false"}
                onClick={() => { onSelect?.(it.value); setRef.current(false); }}
                style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", height: "var(--ctl-md)", padding: "0 var(--sp-4)", background: "transparent", border: 0, borderRadius: "var(--r-sm)", cursor: "pointer", textAlign: "left", font: "var(--type-ui)", color: tone, width: "100%" }}>
                {it.icon ? <Icon name={it.icon} size="sm" color={it.danger ? "var(--st-err)" : "var(--text-muted)"} /> : null}
                <span style={{ flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{it.label}</span>
                {it.shortcut ? <kbd style={{ font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", color: "var(--text-faint)" }}>{it.shortcut}</kbd> : null}
                {it.selected ? <Icon name="check" size="sm" color="var(--brand)" /> : null}
              </button>
            );
          })}
        </div>
      ) : null}
    </span>
  );
}
