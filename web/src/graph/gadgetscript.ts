import type { Catalog, GraphDoc, GraphNode } from "./model";

/* A readable sketch of what the graph asks the Pi to do: the configfs gadget setup for the USB sinks in use
   and one gst-launch chain per media path. The Pi builds the real pipeline; this is for humans and diffs. */

const URL_PROP: Record<string, string> = {
  rtspsrc: "location", souphttpsrc: "location", filesrc: "location", srtsrc: "uri", rtmp2src: "location",
  udpsrc: "uri", uridecodebin: "uri", rtmp2sink: "location", filesink: "location", splitmuxsink: "location",
};

function isControlPlane(gst: string | undefined): boolean { return !gst || /^n\/a/.test(gst); }

/** The node's gst fragment with its configured URL / path applied where the element has such a property. */
export function elementFor(n: GraphNode, catalog: Catalog): string {
  const t = catalog.get(n.type);
  /* Catalogue gst strings may carry a parenthetical note, e.g. "filesink (nfs)"; drop it from the chain. */
  let gst = (t?.gst ?? n.type).replace(/\s*\([^)]*\)/g, "").trim();
  const url = typeof n.config.url === "string" ? n.config.url.trim() : "";
  if (url) {
    const parts = gst.split(" ! ");
    const idx = parts.findIndex(p => URL_PROP[p.split(/[\s(]/)[0]] !== undefined);
    if (idx >= 0) {
      const el = parts[idx].split(/[\s(]/)[0];
      const prop = URL_PROP[el];
      if (!new RegExp("\\b" + prop + "=").test(parts[idx])) parts[idx] = parts[idx] + " " + prop + "=\"" + url.replace(/"/g, "\\\"") + "\"";
      gst = parts.join(" ! ");
    }
  }
  return n.enabled ? gst : "identity drop-probability=1.0 /* bypassed: " + gst + " */";
}

export function generateGadgetScript(doc: GraphDoc, catalog: Catalog): string {
  const byId = new Map(doc.nodes.map(n => [n.id, n]));
  const media = (n: GraphNode | undefined) => !!n && !isControlPlane(catalog.get(n.type)?.gst);
  const out = new Map<string, string[]>();
  const hasIn = new Set<string>();
  doc.edges.forEach(e => {
    if (e.kind !== "video" && e.kind !== "audio") return;
    if (!media(byId.get(e.from.node)) || !media(byId.get(e.to.node))) return;
    out.set(e.from.node, [...(out.get(e.from.node) ?? []), e.to.node]);
    hasIn.add(e.to.node);
  });

  const chains: string[][] = [];
  const walk = (id: string, path: string[]) => {
    if (chains.length >= 24) return;
    const next = (out.get(id) ?? []).filter(x => !path.includes(x));
    if (!next.length) { chains.push(path); return; }
    next.forEach(x => walk(x, [...path, x]));
  };
  doc.nodes.filter(n => media(n) && !hasIn.has(n.id) && out.has(n.id)).forEach(n => walk(n.id, [n.id]));

  const types = new Set(doc.nodes.map(n => n.type));
  const uvc = types.has("uvc_out"), uac = types.has("uac_out") || types.has("uac_in");
  const lines: string[] = ["# generated from “" + doc.name + "” · " + doc.nodes.length + " nodes · " + doc.edges.length + " cables", ""];
  if (uvc || uac) {
    lines.push(
      "modprobe libcomposite",
      "cd /sys/kernel/config/usb_gadget/gadgetcast",
      "echo 0x1d6b > idVendor; echo 0x0104 > idProduct",
    );
    if (uvc) lines.push("mkdir -p functions/uvc.0 && ln -s functions/uvc.0 configs/c.1/");
    if (uac) lines.push("mkdir -p functions/uac2.0 && ln -s functions/uac2.0 configs/c.1/");
    lines.push("ls /sys/class/udc > UDC", "");
  } else {
    lines.push("# no USB gadget sink in this graph · configfs untouched", "");
  }
  if (!chains.length) {
    lines.push("# no complete media path yet · patch a source through to a sink");
    return lines.join("\n");
  }
  lines.push("# one line per media path; where a source fans out, the Pi inserts a tee", "gst-launch-1.0 -e \\");
  chains.forEach((c, i) => {
    const body = c.map(id => elementFor(byId.get(id)!, catalog)).join(" ! ");
    lines.push("  " + body + (i < chains.length - 1 ? " \\" : ""));
  });
  const control = doc.nodes.filter(n => !media(n));
  if (control.length) lines.push("", "# control plane (run by the Pi daemon, not GStreamer): " + control.map(n => n.title || n.type).join(", "));
  return lines.join("\n");
}
