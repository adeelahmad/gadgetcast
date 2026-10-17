import { Fragment, type CSSProperties, type MouseEvent, type PointerEvent, type ReactNode } from "react";
import { Icon } from "../core/Icon";
import { PortHandle, type SignalKind } from "./PortHandle";

export type NodeClass = "source" | "audio" | "video" | "ai" | "logic" | "sink" | "storage" | "bluetooth";
export interface NodePort { kind: SignalKind; label?: string; connected?: boolean; snapped?: boolean }

export const NODE_CLASS_HUE: Record<NodeClass, string> = {
  source: "var(--sig-video)", audio: "var(--sig-audio)", video: "var(--sig-video)", ai: "var(--sig-ai)",
  logic: "var(--sig-data)", sink: "var(--st-ok)", storage: "var(--ink-300)", bluetooth: "var(--sig-bt)",
};

export interface PipelineNodeProps {
  title: string;
  /** Detail line: device path, URL, model name, format. */
  subtitle?: string;
  icon?: string;
  nodeClass?: NodeClass | string;
  inputs?: NodePort[];
  outputs?: NodePort[];
  state?: "idle" | "armed" | "live" | "error";
  latencyMs?: number;
  cpuPct?: number;
  selected?: boolean;
  /** Bypassed node: greys out but keeps wires. */
  disabled?: boolean;
  badge?: ReactNode;
  children?: ReactNode;
  /** Kind of the wire being dragged: matching inputs light up, everything else fades. */
  hotKind?: SignalKind | null;
  onPortDown?: (e: PointerEvent<HTMLSpanElement>) => void;
  /** Per-port overlay inside each hit area: return a React Flow <Handle> here to make the node a flow node. */
  renderHandle?: (direction: "in" | "out", index: number, port: NodePort) => ReactNode;
  onContextMenu?: (e: MouseEvent<HTMLDivElement>) => void;
  onClick?: (e: MouseEvent<HTMLDivElement>) => void;
  style?: CSSProperties;
}

/* Flat node card: tinted icon tile, sans title, ports in aligned in/out columns with the dots on the card edge.
   Selection is a brand ring; live is a pulsing dot. No shadows anywhere. */
export function PipelineNode({ title, subtitle, icon, nodeClass = "source", inputs = [], outputs = [], state = "idle", latencyMs, cpuPct, selected = false, disabled = false, badge, children, hotKind = null, onPortDown, renderHandle, onContextMenu, onClick, style }: PipelineNodeProps) {
  const hue = NODE_CLASS_HUE[nodeClass as NodeClass] || NODE_CLASS_HUE.source;
  const tint = "color-mix(in srgb, " + hue + " 13%, transparent)";
  const rows = Math.max(inputs.length, outputs.length);
  const hot = (p: NodePort) => (hotKind == null ? null : p.kind === hotKind);
  const chip = (v: string, warn: boolean) => (
    <span style={{ font: "var(--fw-medium) var(--fs-meta)/1 var(--font-sans)", color: warn ? "var(--st-warn)" : "var(--text-muted)", background: "var(--surface-input)", border: "1px solid var(--line-hairline)", borderRadius: "var(--r-pill)", padding: "4px 8px" }}>{v}</span>
  );
  return (
    <div className="gc-node" onContextMenu={onContextMenu} onClick={onClick} data-node-class={nodeClass}
      style={{
        width: "var(--w-node)", position: "relative", background: "var(--surface-card)", borderRadius: "var(--r-md)",
        border: "1px solid " + (selected ? "var(--brand)" : "var(--line-border)"), boxShadow: selected ? "0 0 0 1px var(--brand)" : "none",
        opacity: disabled ? 0.5 : 1, filter: disabled ? "grayscale(.7)" : undefined, userSelect: "none", ...style,
      }}>
      <div style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px 8px" }}>
        <span style={{ width: 28, height: 28, flex: "0 0 auto", borderRadius: "var(--r-sm)", background: tint, display: "grid", placeItems: "center" }}>
          <Icon name={icon || "widgets"} size="sm" style={{ color: hue }} />
        </span>
        <span style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column", gap: 2 }}>
          <span style={{ font: "var(--fw-semibold) var(--fs-body)/1.2 var(--font-sans)", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{title}</span>
          {subtitle ? <span style={{ font: "var(--fw-regular) var(--fs-meta)/1.2 var(--font-sans)", color: "var(--text-faint)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{subtitle}</span> : null}
        </span>
        {state === "live" ? <span className="gc-pulse" style={{ width: 7, height: 7, flex: "0 0 auto", borderRadius: "50%", background: "var(--sig-live)" }} /> : null}
        {state === "error" ? <Icon name="error" size="xs" style={{ color: "var(--st-err)" }} /> : null}
        {disabled ? <Icon name="pause_circle" size="xs" style={{ color: "var(--text-faint)" }} /> : null}
      </div>
      {badge ? <div style={{ padding: "0 12px 8px" }}>{badge}</div> : null}
      {children ? <div style={{ padding: "0 12px 10px", minWidth: 0 }}>{children}</div> : null}
      {rows > 0 ? (
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", rowGap: 2, padding: "2px 0 8px", borderTop: badge || children ? "1px solid var(--line-hairline)" : "none", paddingTop: badge || children ? 8 : 2 }}>
          {Array.from({ length: rows }).map((_, i) => {
            const a = inputs[i], b = outputs[i];
            return (
              <Fragment key={i}>
                <div style={{ marginLeft: -10, display: "flex", alignItems: "center", minHeight: 22 }}>
                  {a ? <PortHandle direction="in" index={i} kind={a.kind} label={a.label} connected={a.connected} snapped={a.snapped} hot={hot(a)} onPortDown={onPortDown}>{renderHandle ? renderHandle("in", i, a) : null}</PortHandle> : null}
                </div>
                <div style={{ marginRight: -10, display: "flex", alignItems: "center", justifyContent: "flex-end", minHeight: 22 }}>
                  {b ? <PortHandle direction="out" index={i} kind={b.kind} label={b.label} connected={b.connected} hot={hotKind == null ? null : false} onPortDown={onPortDown}>{renderHandle ? renderHandle("out", i, b) : null}</PortHandle> : null}
                </div>
              </Fragment>
            );
          })}
        </div>
      ) : null}
      {latencyMs != null || cpuPct != null ? (
        <div style={{ display: "flex", gap: 6, padding: "0 12px 10px" }}>
          {latencyMs != null ? chip(latencyMs + " ms", latencyMs > 200) : null}
          {cpuPct != null ? chip("CPU " + cpuPct + "%", cpuPct > 75) : null}
        </div>
      ) : null}
    </div>
  );
}
