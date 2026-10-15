import type { ButtonHTMLAttributes } from "react";
import { Icon, type IconSize } from "./Icon";

export type IconButtonSize = "xs" | "sm" | "md" | "lg";
const H: Record<IconButtonSize, string> = { xs: "var(--ctl-xs)", sm: "var(--ctl-sm)", md: "var(--ctl-md)", lg: "var(--ctl-lg)" };
const I: Record<IconButtonSize, IconSize> = { xs: "xs", sm: "sm", md: "sm", lg: "md" };

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: string;
  /** Accessible name; also the native title. */
  label: string;
  size?: IconButtonSize;
  active?: boolean;
  outlined?: boolean;
  /** Glyph colour override, a token reference. */
  tone?: string;
}

export function IconButton({ icon, size = "md", active = false, outlined = false, label, tone, style, type = "button", ...rest }: IconButtonProps) {
  return (
    <button type={type} className={"gc-iconbtn" + (outlined ? " gc-iconbtn--outlined" : "")} data-active={active ? "true" : "false"}
      title={label} aria-label={label} aria-pressed={active || undefined}
      style={{ width: H[size], height: H[size], color: tone || undefined, padding: 0, ...style }} {...rest}>
      <Icon name={icon} size={I[size]} filled={active} />
    </button>
  );
}
