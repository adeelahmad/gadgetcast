import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

/* Text tokens must hold WCAG AA (4.5:1) as small text on the surfaces they sit on, in both themes.
   axe-core found these failing on the real pages; this keeps them fixed. */
const css = readFileSync(new URL("./colors.css", import.meta.url), "utf8");

function tokens(block: string): Record<string, string> {
  const out: Record<string, string> = {};
  for (const m of block.matchAll(/--([\w-]+):\s*([^;]+);/g)) out[m[1]] = m[2].trim();
  return out;
}
const light = tokens(css.slice(css.indexOf(":root{"), css.indexOf('[data-theme="dark"]{')));
const dark = { ...light, ...tokens(css.slice(css.indexOf('[data-theme="dark"]{'))) };

function resolve(t: Record<string, string>, v: string): string {
  const m = /^var\(--([\w-]+)\)$/.exec(v);
  return m ? resolve(t, t[m[1]]) : v;
}
function lum(hex: string): number {
  const [r, g, b] = [1, 3, 5].map(i => parseInt(hex.slice(i, i + 2), 16) / 255)
    .map(c => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}
function contrast(a: string, b: string): number {
  const [x, y] = [lum(a), lum(b)].sort((p, q) => q - p);
  return (x + 0.05) / (y + 0.05);
}
/** a over b at alpha p, as color-mix(in srgb, a p%, b) */
function mix(a: string, p: number, b: string): string {
  return "#" + [1, 3, 5].map(i => Math.round(parseInt(a.slice(i, i + 2), 16) * p + parseInt(b.slice(i, i + 2), 16) * (1 - p)).toString(16).padStart(2, "0")).join("");
}

const surfaces = ["surface-canvas", "surface-app", "surface-panel", "surface-card", "surface-raised", "surface-input"];
const text = ["text-primary", "text-body", "text-muted", "text-faint", "text-link"];
const accents = ["sig-video", "sig-audio", "sig-live", "st-ok", "st-warn", "st-err", "brand"];

describe.each([["light", light], ["dark", dark]] as const)("%s theme", (_, t) => {
  const col = (k: string) => resolve(t, t[k]).slice(0, 7);
  it.each(text)("%s holds 4.5:1 on every surface", k => {
    for (const s of surfaces) expect(contrast(col(k), col(s)), `${k} on ${s}`).toBeGreaterThanOrEqual(4.5);
  });
  it.each(accents)("%s holds 4.5:1 on its own 12%% tint", k => {
    const bg = col("surface-raised");
    expect(contrast(col(k), mix(col(k), 0.12, bg))).toBeGreaterThanOrEqual(4.5);
  });
});
