import { StrictMode, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import "./ds/styles.css";
import "./app/studio.css";
import logoDark from "./ds/assets/logo-dark.svg";
import logo from "./ds/assets/logo.svg";
import { selectBackend, type Selection } from "./api";
import { App } from "./app/App";

function Root() {
  const [sel, setSel] = useState<Selection | null>(null);
  useEffect(() => {
    const forceDemo = new URLSearchParams(window.location.search).has("demo");
    let alive = true;
    void selectBackend({ forceDemo }).then(s => { if (alive) setSel(s); });
    return () => { alive = false; };
  }, []);
  if (!sel) {
    const dark = document.documentElement.getAttribute("data-theme") === "dark";
    return (
      <div className="gc-boot" role="status">
        <img src={dark ? logoDark : logo} alt="GadgetCast" style={{ height: 30 }} />
        <span style={{ font: "var(--type-meta)", color: "var(--text-faint)" }}>Connecting to the Pi…</span>
      </div>
    );
  }
  return <App backend={sel.backend} reason={sel.reason} />;
}

createRoot(document.getElementById("root")!).render(<StrictMode><Root /></StrictMode>);
