import { describe, expect, it } from "vitest";
import type { NodeGroup } from "../api/types";
import { indexCatalog, type GraphDoc } from "./model";
import { connectionError, graphProblems, lookupFor } from "./validate";

const groups: NodeGroup[] = [{
  group: "Test", items: [
    { type: "cam", title: "Camera", icon: "videocam", cls: "source", category: "source", out: [["video", "Raw"], ["audio", "PCM"]] },
    { type: "enc", title: "Encoder", icon: "compress", cls: "video", category: "encoder", in: [["video", "Raw"], ["data", "Text"]], out: [["video", "MJPEG"]] },
    { type: "uvc", title: "UVC", icon: "usb", cls: "sink", category: "endpoint", in: [["video", "MJPEG"]] },
    { type: "mic", title: "UAC2", icon: "mic", cls: "sink", category: "endpoint", in: [["audio", "PCM"]] },
  ],
}];
const catalog = indexCatalog(groups);
const doc: GraphDoc = {
  id: "t", name: "t",
  nodes: [
    { id: "a", type: "cam", x: 0, y: 0, enabled: true, config: {} },
    { id: "b", type: "enc", x: 0, y: 0, enabled: true, config: {} },
    { id: "c", type: "uvc", x: 0, y: 0, enabled: true, config: {} },
    { id: "d", type: "mic", x: 0, y: 0, enabled: true, config: {} },
  ],
  edges: [{ id: "e1", from: { node: "b", port: 0 }, to: { node: "c", port: 0 }, kind: "video" }],
};
const lookup = lookupFor(doc, catalog);
const conn = (source: string, sh: string, target: string, th: string) => ({ source, sourceHandle: sh, target, targetHandle: th });

describe("connectionError: port compatibility", () => {
  it("accepts an output → input of the same kind", () => {
    expect(connectionError(conn("a", "out-0", "b", "in-0"), lookup, doc.edges)).toBeNull();
    expect(connectionError(conn("a", "out-1", "d", "in-0"), lookup, doc.edges)).toBeNull();
  });

  it("refuses mismatched kinds with a reason that names both kinds", () => {
    expect(connectionError(conn("a", "out-1", "b", "in-0"), lookup, doc.edges)).toBe("Cannot link AUDIO output to VIDEO input");
    expect(connectionError(conn("a", "out-0", "b", "in-1"), lookup, doc.edges)).toBe("Cannot link VIDEO output to DATA input");
  });

  it("refuses a node patched into itself", () => {
    expect(connectionError(conn("b", "out-0", "b", "in-0"), lookup, doc.edges)).toBe("A node cannot patch into itself");
  });

  it("allows one cable per input", () => {
    expect(connectionError(conn("b", "out-0", "c", "in-0"), lookup, doc.edges)).toBe("That input is already patched");
  });

  it("lets the cable being re-routed vacate its own input", () => {
    expect(connectionError(conn("b", "out-0", "c", "in-0"), lookup, doc.edges, "e1")).toBeNull();
  });

  it("allows fan-out from one output to many inputs", () => {
    const edges = [...doc.edges, { id: "e2", from: { node: "a", port: 0 }, to: { node: "b", port: 0 }, kind: "video" as const }];
    expect(connectionError(conn("a", "out-0", "b", "in-0"), lookup, edges)).toBe("That input is already patched");
    expect(connectionError(conn("b", "out-0", "c", "in-0"), lookup, [])).toBeNull();
  });

  it("refuses missing ports and wrong directions", () => {
    expect(connectionError(conn("a", "out-5", "b", "in-0"), lookup, doc.edges)).toBe("That port does not exist");
    expect(connectionError(conn("a", "in-0", "b", "out-0"), lookup, doc.edges)).toBe("Patch from an output to an input");
    expect(connectionError({ source: "a", sourceHandle: "out-0", target: null, targetHandle: null }, lookup, doc.edges)).toBe("Drop the cable on a port");
  });
});

describe("graphProblems", () => {
  it("is empty for a valid graph and lists every broken cable otherwise", () => {
    expect(graphProblems(doc, catalog)).toEqual([]);
    const bad: GraphDoc = { ...doc, edges: [...doc.edges,
      { id: "x1", from: { node: "a", port: 1 }, to: { node: "c", port: 0 }, kind: "audio" },
      { id: "x2", from: { node: "a", port: 0 }, to: { node: "a", port: 0 }, kind: "video" }] };
    const p = graphProblems(bad, catalog);
    expect(p).toHaveLength(2);
    expect(p[0]).toMatch(/^x1: /);
    expect(p[1]).toBe("x2: A node cannot patch into itself");
  });
});
