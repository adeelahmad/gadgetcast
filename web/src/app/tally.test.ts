import { describe, expect, it } from "vitest";
import { tallyOf } from "./tally";

describe("tallyOf", () => {
  it("is live whenever a client is on air, even with no pipeline running", () => {
    expect(tallyOf({ running: null, onAir: "mic-1" })).toBe("live");
    expect(tallyOf({ running: "p", onAir: "mic-1" })).toBe("live");
  });
  it("is armed with a pipeline and nobody on air, off otherwise", () => {
    expect(tallyOf({ running: "p", onAir: null })).toBe("armed");
    expect(tallyOf({ running: null, onAir: null })).toBe("off");
    expect(tallyOf(null)).toBe("off");
  });
});
