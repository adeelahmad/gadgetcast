import { describe, expect, it } from "vitest";
import { clickable } from "./clickable";

const key = (k: string, nested = false) => {
  const el = {};
  let prevented = false;
  return { e: { key: k, target: nested ? {} : el, currentTarget: el, preventDefault: () => { prevented = true; } } as never, prevented: () => prevented };
};

describe("clickable", () => {
  it("is empty without an action", () => {
    expect(clickable(undefined)).toEqual({});
  });

  it("exposes a focusable button and reports selection", () => {
    const p = clickable(() => {}, true);
    expect(p.role).toBe("button");
    expect(p.tabIndex).toBe(0);
    expect(p["aria-pressed"]).toBe(true);
    expect(clickable(() => {})["aria-pressed"]).toBeUndefined();
  });

  it("activates on Enter and Space only", () => {
    let n = 0;
    const p = clickable(() => { n++; });
    const enter = key("Enter");
    p.onKeyDown!(enter.e);
    expect(enter.prevented()).toBe(true);
    p.onKeyDown!(key(" ").e);
    p.onKeyDown!(key("a").e);
    expect(n).toBe(2);
  });

  it("ignores keys from nested controls", () => {
    let n = 0;
    clickable(() => { n++; }).onKeyDown!(key("Enter", true).e);
    expect(n).toBe(0);
  });
});
