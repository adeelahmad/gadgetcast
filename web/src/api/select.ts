import type { Backend } from "./backend";
import { createHttpBackend, probeHealth, type FetchLike } from "./client";
import { createMockBackend, type MockOptions } from "./mock";

export interface SelectOptions {
  fetch?: FetchLike;
  base?: string;
  timeoutMs?: number;
  /** Force the offline demo (e.g. ?demo=1). */
  forceDemo?: boolean;
  mock?: MockOptions;
}

export interface Selection {
  backend: Backend;
  /** Why the demo was chosen; undefined when live. */
  reason?: string;
}

/**
 * The single mock/real seam: probe GET /api/health and hand back the HTTP client when the Pi answers
 * {"ok":true}, the bundled demo otherwise.
 */
export async function selectBackend(opts: SelectOptions = {}): Promise<Selection> {
  const f: FetchLike = opts.fetch ?? ((i, init) => fetch(i, init));
  if (opts.forceDemo) return { backend: createMockBackend(opts.mock), reason: "Demo forced by ?demo=1" };
  const health = await probeHealth(f, opts.base ?? "", opts.timeoutMs ?? 2500);
  if (health) return { backend: createHttpBackend({ base: opts.base, fetch: opts.fetch, version: health.version }) };
  return { backend: createMockBackend(opts.mock), reason: "The Pi did not answer /api/health" };
}
