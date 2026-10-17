import type { KeyboardEvent } from "react";

/** Props that make a non-button element (row, table row) reachable and
 *  operable from the keyboard and announced as a button. Empty when the
 *  element has no action, so static rows stay out of the tab order. */
export function clickable(onClick: (() => void) | undefined, selected?: boolean) {
  if (!onClick) return {};
  return {
    role: "button" as const,
    tabIndex: 0,
    "aria-pressed": selected === undefined ? undefined : selected,
    onKeyDown: (e: Pick<KeyboardEvent, "key" | "target" | "currentTarget" | "preventDefault">) => {
      if (e.target !== e.currentTarget) return; // let nested buttons handle their own keys
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onClick();
      }
    },
  };
}
