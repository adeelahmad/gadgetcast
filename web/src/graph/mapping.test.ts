import { describe, expect, it } from "vitest";
import { CATALOG, seedPipelines } from "../api/mockData";
import type { Pipeline } from "../api/types";
import { generateGadgetScript } from "./gadgetscript";
import { graphToPipeline, parseImportedPipeline, pipelineToGraph } from "./mapping";
import { indexCatalog, portsOf, slugId } from "./model";
import { graphProblems } from "./validate";

const full: Pipeline = {
  id: "studio-a", name: "Studio A", description: "desk rig", isDefault: true, updatedAt: "2026-10-01T09:00:00Z",
  nodes: [
    { id: "cam_1", type: "csi", x: 32, y: 48, title: "Desk cam", enabled: true, config: { url: "" } },
    { id: "enc_1", type: "mjpeg_enc", x: 272, y: 48, title: "Encode", enabled: false, config: { format: "MJPEG 720p30", gain: -3 } },
    { id: "uvc_1", type: "uvc_out", x: 512, y: 48, title: "UVC", enabled: true, config: {} },
  ],
  edges: [
    { id: "e1", from: "cam_1", fromPort: 0, to: "enc_1", toPort: 0, kind: "video" },
    { id: "e2", from: "enc_1", fromPort: 0, to: "uvc_1", toPort: 0, kind: "video" },
  ],
};

describe("graph ↔ Pipeline mapping", () => {
  it("round-trips a fully specified Pipeline unchanged", () => {
    const { doc, dropped } = pipelineToGraph(full);
    expect(dropped).toBe(0);
    expect(doc.edges[0]).toEqual({ id: "e1", from: { node: "cam_1", port: 0 }, to: { node: "enc_1", port: 0 }, kind: "video" });
    expect(graphToPipeline(doc)).toEqual(full);
  });

  it("round-trips every bundled preset", () => {
    for (const p of seedPipelines()) expect(graphToPipeline(pipelineToGraph(p).doc)).toEqual(p);
  });

  it("fills contract defaults: enabled defaults to true, config to {}", () => {
    const sparse: Pipeline = { id: "s", name: "S", nodes: [{ id: "n", type: "csi", x: 1.4, y: 2.6 }], edges: [] };
    const { doc } = pipelineToGraph(sparse);
    expect(doc.nodes[0]).toEqual({ id: "n", type: "csi", x: 1.4, y: 2.6, enabled: true, config: {} });
    expect(graphToPipeline(doc).nodes[0]).toEqual({ id: "n", type: "csi", x: 1, y: 3, enabled: true, config: {} });
  });

  it("drops cables that point at missing nodes", () => {
    const broken: Pipeline = { ...full, edges: [...full.edges, { id: "e9", from: "ghost", fromPort: 0, to: "uvc_1", toPort: 0, kind: "video" }] };
    const { doc, dropped } = pipelineToGraph(broken);
    expect(dropped).toBe(1);
    expect(doc.edges.map(e => e.id)).toEqual(["e1", "e2"]);
  });

  it("does not share config objects between the wire and the graph", () => {
    const { doc } = pipelineToGraph(full);
    doc.nodes[1].config.gain = 6;
    expect(full.nodes[1].config).toEqual({ format: "MJPEG 720p30", gain: -3 });
  });

  it("imports the design-system prototype export shape", () => {
    const p = parseImportedPipeline({ name: "Old", nodes: [{ id: "a", type: "csi", x: 0, y: 0, title: "Cam" }, { id: "b", type: "uvc_out", x: 1, y: 1 }],
      edges: [{ id: "e1", from: { node: "a", port: 0 }, to: { node: "b", port: 0 }, kind: "video" }] });
    expect(p.edges[0]).toEqual({ id: "e1", from: "a", fromPort: 0, to: "b", toPort: 0, kind: "video" });
    expect(p.nodes[0].title).toBe("Cam");
    expect(() => parseImportedPipeline({ nodes: [] })).toThrow(/edges/);
  });

  it("keeps unknown node types renderable by deriving ports from their cables", () => {
    const catalog = indexCatalog(CATALOG);
    const { doc } = pipelineToGraph({ ...full, nodes: [...full.nodes, { id: "z", type: "future_node", x: 0, y: 0 }],
      edges: [...full.edges, { id: "e3", from: "z", fromPort: 1, to: "uvc_1", toPort: 0, kind: "video" }] });
    const ports = portsOf(doc.nodes[3], catalog, doc.edges);
    expect(ports.inputs).toEqual([]);
    expect(ports.outputs.map(p => p.kind)).toEqual(["data", "video"]);
  });
});

describe("bundled presets", () => {
  it("are all valid graphs against the bundled catalogue", () => {
    const catalog = indexCatalog(CATALOG);
    for (const p of seedPipelines()) expect(graphProblems(pipelineToGraph(p).doc, catalog)).toEqual([]);
  });
});

describe("slugId", () => {
  it("makes contract-valid unique ids", () => {
    expect(slugId("USB Webcam Streaming", [])).toBe("usb-webcam-streaming");
    expect(slugId("USB Webcam Streaming", ["usb-webcam-streaming"])).toBe("usb-webcam-streaming-2");
    expect(slugId("  ··· ", [])).toBe("pipeline");
    expect(slugId("x".repeat(200), [])).toMatch(/^[a-zA-Z0-9_-]{1,64}$/);
  });
});

describe("generateGadgetScript", () => {
  it("emits configfs for the gadget sinks and one gst chain per media path", () => {
    const catalog = indexCatalog(CATALOG);
    const s = generateGadgetScript(pipelineToGraph(seedPipelines()[0]).doc, catalog);
    expect(s).toContain("functions/uvc.0");
    expect(s).toContain("functions/uac2.0");
    expect(s).toContain("v4l2h264dec ! v4l2jpegenc ! v4l2sink device=/dev/video3");
    expect(s).toContain("opusdec plc=true ! webrtcdsp ! alsasink device=hw:1,0");
  });
});
