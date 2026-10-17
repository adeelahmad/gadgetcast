import type { Status } from "../api/types";

export type Tally = "live" | "armed" | "off";

/** The on-air indicator is a privacy control (README §4): it must read "live" whenever any client's media
 *  reaches the USB host, whether or not a Studio pipeline is running. A running pipeline with nobody on
 *  air is "armed". */
export function tallyOf(st: Pick<Status, "running" | "onAir"> | null | undefined): Tally {
  if (st?.onAir) return "live";
  return st?.running ? "armed" : "off";
}
