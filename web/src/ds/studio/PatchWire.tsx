import type { MouseEvent, PointerEvent } from "react";
import { SIGNAL_HUE, type SignalKind } from "./PortHandle";

export interface Pt { x: number; y: number }

export function wirePath(from: Pt, to: Pt): string {
  const dist = Math.abs(to.x - from.x);
  const dx = Math.max(56, dist * 0.5) + (to.x < from.x ? Math.min(120, (from.x - to.x) * 0.3) : 0);
  return "M " + from.x + " " + from.y + " C " + (from.x + dx) + " " + from.y + ", " + (to.x - dx) + " " + to.y + ", " + to.x + " " + to.y;
}

export interface PatchWireProps {
  from: Pt;
  to: Pt;
  kind?: SignalKind;
  flowing?: boolean;
  invalid?: boolean;
  dimmed?: boolean;
  selected?: boolean;
  hovered?: boolean;
  onClick?: (e: MouseEvent<SVGGElement>) => void;
  onContextMenu?: (e: MouseEvent<SVGGElement>) => void;
  onPointerEnter?: (e: PointerEvent<SVGGElement>) => void;
  onPointerLeave?: (e: PointerEvent<SVGGElement>) => void;
}

/* Cubic cable between two port centres. A wide transparent stroke underneath is the hit target. */
export function PatchWire({ from, to, kind = "video", flowing = false, invalid = false, dimmed = false, selected = false, hovered = false, onClick, onContextMenu, onPointerEnter, onPointerLeave }: PatchWireProps) {
  const d = wirePath(from, to);
  const hue = invalid ? "var(--st-err)" : SIGNAL_HUE[kind] || SIGNAL_HUE.data;
  const w = selected ? 3 : hovered ? 2.75 : 2;
  return (
    <g className="gc-wire" data-kind={kind} onClick={onClick} onContextMenu={onContextMenu} onPointerEnter={onPointerEnter} onPointerLeave={onPointerLeave}
      style={{ cursor: onClick ? "pointer" : "default", opacity: dimmed ? 0.22 : 1, transition: "opacity 160ms" }}>
      <path d={d} fill="none" stroke="transparent" strokeWidth="16" pointerEvents="stroke" />
      {selected ? <path d={d} fill="none" stroke="var(--brand)" strokeOpacity=".22" strokeWidth={w + 6} strokeLinecap="round" pointerEvents="none" /> : null}
      <path d={d} fill="none" stroke={selected ? "var(--brand)" : hue} strokeWidth={w} strokeLinecap="round" pointerEvents="none"
        strokeDasharray={invalid ? "4 5" : undefined} style={{ transition: "stroke-width 120ms" }} />
      {flowing && !invalid ? <path className="gc-wire-flow" d={d} fill="none" stroke="var(--surface-panel)" strokeOpacity=".95" strokeWidth="1.5" strokeLinecap="round" pointerEvents="none" /> : null}
    </g>
  );
}
