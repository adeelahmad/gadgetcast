import type { ButtonHTMLAttributes, ReactNode } from "react";
import { Icon, type IconSize } from "./Icon";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "live";
export type ButtonSize = "xs" | "sm" | "md" | "lg";
const S: Record<ButtonSize, { h: string; px: string; fs: string; ic: IconSize }> = {
  xs: { h: "var(--ctl-xs)", px: "var(--sp-3)", fs: "var(--fs-caps)", ic: "xs" },
  sm: { h: "var(--ctl-sm)", px: "var(--sp-4)", fs: "var(--fs-meta)", ic: "xs" },
  md: { h: "var(--ctl-md)", px: "var(--sp-5)", fs: "var(--fs-ui)", ic: "sm" },
  lg: { h: "var(--ctl-lg)", px: "var(--sp-7)", fs: "var(--fs-body)", ic: "md" },
};

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: string;
  iconEnd?: string;
  fullWidth?: boolean;
  children?: ReactNode;
}

export function Button({ variant = "secondary", size = "md", icon, iconEnd, fullWidth = false, children, style, type = "button", ...rest }: ButtonProps) {
  const s = S[size];
  return (
    <button type={type} className={"gc-btn gc-btn--" + variant}
      style={{ height: s.h, padding: "0 " + s.px, fontSize: s.fs, width: fullWidth ? "100%" : undefined, ...style }} {...rest}>
      {icon ? <Icon name={icon} size={s.ic} /> : null}
      {children != null ? <span>{children}</span> : null}
      {iconEnd ? <Icon name={iconEnd} size={s.ic} /> : null}
    </button>
  );
}
