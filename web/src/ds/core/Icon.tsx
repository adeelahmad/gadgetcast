import type { CSSProperties, HTMLAttributes } from "react";

export type IconSize = "xs" | "sm" | "md" | "lg" | "xl";
const SIZES: Record<IconSize, number> = { xs: 12, sm: 14, md: 16, lg: 20, xl: 24 };

export interface IconProps extends Omit<HTMLAttributes<HTMLSpanElement>, "color"> {
  /** Material Symbols Rounded ligature name, e.g. "videocam". */
  name: string;
  size?: IconSize | number;
  /** Filled glyph: reserved for active / selected state. */
  filled?: boolean;
  color?: string;
  style?: CSSProperties;
}

export function Icon({ name, size = "md", filled = false, color, style, ...rest }: IconProps) {
  const px = typeof size === "number" ? size : SIZES[size];
  return (
    <span className="gc-icon" data-filled={filled ? "true" : "false"} aria-hidden="true"
      style={{ fontSize: px, width: px, height: px, color: color || "inherit", flex: "0 0 auto", ...style }} {...rest}>{name}</span>
  );
}
