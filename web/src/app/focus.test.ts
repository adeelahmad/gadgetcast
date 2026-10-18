import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

/* studio.css turns React Flow's focus styles off, and React Flow's own stylesheet loads after it
   with .react-flow__node.selectable:focus-visible{outline:none}. A keyboard walkthrough found canvas
   nodes and cables took focus with nothing drawn; these rules keep the focus visible (WCAG 2.4.7). */
const css = readFileSync(new URL("./studio.css", import.meta.url), "utf8");

function rule(selector: string): string {
  const i = css.indexOf(selector + "{");
  return i < 0 ? "" : css.slice(i + selector.length + 1, css.indexOf("}", i));
}

describe("canvas keyboard focus", () => {
  it("outlines a focused node, out-ranking React Flow's own rule", () => {
    expect(rule(".react-flow .react-flow__node.react-flow__node-gc:focus-visible")).toMatch(/outline:\s*2px solid var\(--brand\)/);
  });
  it("paints a halo on a focused cable", () => {
    expect(rule(".react-flow__edge:focus-visible .gc-wire>path:first-child")).toMatch(/stroke:\s*var\(--brand\)/);
  });
});
