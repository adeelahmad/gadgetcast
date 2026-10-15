import { existsSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig, type Plugin } from "vitest/config";

const PI = process.env.GADGETCAST_API ?? "http://localhost:8080";
const outDir = resolve(__dirname, "dist");

/* dist/.gitkeep is tracked so the Go embed directive always has a directory to embed. Vite empties outDir
   on build, so put the placeholder back if it was there before. */
function keepGitkeep(): Plugin {
  const keep = resolve(outDir, ".gitkeep");
  let had = false;
  return {
    name: "gadgetcast-keep-gitkeep",
    apply: "build",
    buildStart() { had = existsSync(keep); },
    closeBundle() { if (had && !existsSync(keep)) writeFileSync(keep, ""); },
  };
}

export default defineConfig({
  plugins: [react(), keepGitkeep()],
  base: "/",
  build: {
    outDir,
    emptyOutDir: true,
    target: "es2020",
    sourcemap: false,
    chunkSizeWarningLimit: 900,
  },
  server: {
    port: 5173,
    proxy: {
      "/api": { target: PI, changeOrigin: true, ws: true },
      "/ws": { target: PI, changeOrigin: true, ws: true },
    },
  },
  preview: { port: 4173 },
  test: {
    environment: "node",
    include: ["src/**/*.test.ts"],
  },
});
