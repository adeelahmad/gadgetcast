import { useEffect, useRef, type CSSProperties, type ReactNode } from "react";
import { Icon } from "../core/Icon";
import { Badge, type BadgeTone } from "../core/Badge";

/* ── LogStream ── mono tail; follows the end while `follow` is on. */
export type LogLevel = "info" | "ok" | "warn" | "err" | "dbg";
export interface LogLine { id?: string | number; t: string; src: string; level: LogLevel; msg: string }
const LV: Record<LogLevel, string> = { info: "var(--text-muted)", ok: "var(--st-ok)", warn: "var(--st-warn)", err: "var(--st-err)", dbg: "var(--text-faint)" };

export function LogStream({ lines, height = 180, follow = true, empty = "No log lines yet.", style }: { lines: LogLine[]; height?: number | string; follow?: boolean; empty?: ReactNode; style?: CSSProperties }) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => { const el = ref.current; if (el && follow) el.scrollTop = el.scrollHeight; }, [lines, follow]);
  return (
    <div ref={ref} className="gc-scroll" role="log" aria-live="off"
      style={{ height, overflow: "auto", background: "var(--surface-canvas)", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)", padding: "var(--sp-4)", display: "flex", flexDirection: "column", gap: 1, ...style }}>
      {lines.length === 0 ? <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>{empty}</span> : null}
      {lines.map((l, i) => (
        <div key={l.id ?? i} style={{ display: "flex", gap: "var(--sp-4)", font: "var(--type-meta)", fontSize: "var(--fs-tick)", whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
          <span style={{ color: "var(--text-faint)", flex: "0 0 auto" }}>{l.t}</span>
          <span style={{ color: "var(--text-faint)", flex: "0 0 68px", textTransform: "uppercase" }}>{l.src}</span>
          <span style={{ color: LV[l.level] || LV.info, flex: 1, minWidth: 0 }}>{l.msg}</span>
        </div>
      ))}
    </div>
  );
}

/* ── PresetRow ── 2px left rail: red = running, cyan = selected. */
export interface PresetRowProps { name: string; meta?: ReactNode; icon?: string; selected?: boolean; running?: boolean; isDefault?: boolean; actions?: ReactNode; onClick?: () => void; style?: CSSProperties }

export function PresetRow({ name, meta, icon = "account_tree", selected = false, running = false, isDefault = false, actions, onClick, style }: PresetRowProps) {
  return (
    <div className="gc-row" data-selected={selected ? "true" : "false"} onClick={onClick} role="button" tabIndex={0}
      onKeyDown={e => { if (e.key === "Enter" && e.target === e.currentTarget) onClick?.(); }}
      style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-4) var(--sp-5)", borderRadius: "var(--r-sm)", cursor: "pointer", borderLeft: "2px solid " + (running ? "var(--sig-live)" : selected ? "var(--brand)" : "transparent"), ...style }}>
      <Icon name={icon} size="sm" style={{ color: running ? "var(--sig-live)" : "var(--text-faint)" }} />
      <div style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", gap: 1 }}>
        <span style={{ font: "var(--type-ui)", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{name}</span>
        {meta ? <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)" }}>{meta}</span> : null}
      </div>
      {running ? <Badge tone="live" dot>Live</Badge> : null}
      {isDefault && !running ? <Icon name="push_pin" size="xs" style={{ color: "var(--text-faint)" }} /> : null}
      {actions}
    </div>
  );
}

/* ── DeviceRow ── */
export type DeviceStatus = "online" | "offline" | "pairing" | "error";
const ST: Record<DeviceStatus, [BadgeTone, string]> = { online: ["ok", "Online"], offline: ["neutral", "Offline"], pairing: ["warn", "Pairing"], error: ["err", "Fault"] };
export interface DeviceRowProps { name: string; detail?: ReactNode; icon?: string; transport?: string; status?: DeviceStatus; statusLabel?: string; trailing?: ReactNode; onClick?: () => void; selected?: boolean; style?: CSSProperties }

export function DeviceRow({ name, detail, icon = "usb", transport, status = "online", statusLabel, trailing, onClick, selected = false, style }: DeviceRowProps) {
  const s = ST[status];
  return (
    <div className="gc-row" data-selected={selected ? "true" : "false"} onClick={onClick}
      style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", padding: "var(--sp-4) var(--sp-5)", borderBottom: "1px solid var(--line-hairline)", cursor: onClick ? "pointer" : "default", ...style }}>
      <span style={{ display: "flex", alignItems: "center", justifyContent: "center", width: 24, height: 24, flex: "0 0 auto", background: "var(--surface-raised)", border: "1px solid var(--line-border)", borderRadius: "var(--r-sm)" }}>
        <Icon name={icon} size="sm" style={{ color: "var(--text-muted)" }} />
      </span>
      <div style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", gap: 1 }}>
        <span style={{ font: "var(--type-ui)", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{name}</span>
        {detail ? <span style={{ font: "var(--type-meta)", fontSize: "var(--fs-tick)", color: "var(--text-faint)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{detail}</span> : null}
      </div>
      {transport ? <Badge tone="neutral">{transport}</Badge> : null}
      <Badge tone={s[0]} dot>{statusLabel || s[1]}</Badge>
      {trailing}
    </div>
  );
}

/* ── Table ── dense list surface: sticky mono-caps header, hairline rows, mono machine values, 2px state rail. */
export interface TableColumn<R> { key?: keyof R & string; label: string; width?: string | number; align?: "left" | "right" | "center"; mono?: boolean; render?: (row: R) => ReactNode; icon?: (row: R) => string }
export interface TableProps<R> { columns: TableColumn<R>[]; rows: R[]; rowKey: (row: R, index: number) => string; selectedKey?: string; onSelect?: (key: string, row: R) => void; rail?: (row: R) => string | null; dense?: boolean; empty?: ReactNode; footer?: ReactNode; style?: CSSProperties }

export function Table<R>({ columns, rows, rowKey, selectedKey, onSelect, rail, dense = false, empty = "No rows.", footer, style }: TableProps<R>) {
  const h = dense ? "var(--ctl-sm)" : "var(--ctl-md)";
  const cell = (c: TableColumn<R>): CSSProperties => ({ padding: "0 var(--sp-5)", textAlign: c.align || "left", font: c.mono ? "var(--type-meta)" : "var(--type-ui)", letterSpacing: c.mono ? "var(--ls-mono)" : "normal", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis", maxWidth: 0 });
  return (
    <div style={{ border: "1px solid var(--line-border)", borderRadius: "var(--r-md)", background: "var(--surface-card)", overflow: "hidden", ...style }}>
      <div style={{ overflow: "auto" }} className="gc-scroll">
        <table style={{ width: "100%", borderCollapse: "collapse", tableLayout: "fixed" }}>
          <colgroup>{columns.map((c, i) => <col key={i} style={{ width: c.width || "auto" }} />)}</colgroup>
          <thead>
            <tr style={{ height: "var(--ctl-sm)", background: "var(--surface-panel)" }}>
              {columns.map((c, i) => (
                <th key={i} scope="col" style={{ ...cell(c), position: "sticky", top: 0, zIndex: 2, background: "var(--surface-panel)", borderBottom: "1px solid var(--line-border)", font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: "var(--text-faint)" }}>{c.label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr><td colSpan={columns.length} style={{ padding: "var(--sp-8) var(--sp-5)", textAlign: "center", font: "var(--type-body)", color: "var(--text-faint)" }}>{empty}</td></tr>
            ) : rows.map((r, ri) => {
              const key = rowKey(r, ri);
              const sel = selectedKey !== undefined && selectedKey === key;
              const railTone = rail ? rail(r) : null;
              return (
                <tr key={key} className="gc-row" data-selected={sel ? "true" : "false"} onClick={onSelect ? () => onSelect(key, r) : undefined}
                  style={{ height: h, borderBottom: "1px solid var(--line-hairline)", cursor: onSelect ? "pointer" : "default" }}>
                  {columns.map((c, ci) => {
                    const v = c.render ? c.render(r) : c.key ? (r[c.key] as ReactNode) : null;
                    return (
                      <td key={ci} style={{ ...cell(c), color: c.mono ? "var(--text-muted)" : "var(--text-body)", boxShadow: ci === 0 ? "inset 2px 0 0 " + (sel ? "var(--brand)" : railTone || "transparent") : undefined }}>
                        {c.icon && ci === 0 ? <Icon name={c.icon(r)} size="xs" color="var(--text-faint)" style={{ marginRight: "var(--sp-3)", verticalAlign: "-2px", display: "inline-block" }} /> : null}
                        {v}
                      </td>
                    );
                  })}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      {footer ? (
        <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-4)", height: "var(--ctl-sm)", padding: "0 var(--sp-5)", borderTop: "1px solid var(--line-border)", background: "var(--surface-panel)", font: "var(--type-caps)", letterSpacing: "var(--ls-caps)", textTransform: "uppercase", color: "var(--text-faint)" }}>{footer}</div>
      ) : null}
    </div>
  );
}
