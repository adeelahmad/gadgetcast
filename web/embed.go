// Package web embeds the built Studio UI (web/dist) into the server binary.
//
// dist/ is produced by the UI build; a .gitkeep keeps the directory present so
// the Go build succeeds without a UI build (the server then serves a notice).
package web

import "embed"

// Dist holds the built UI under the "dist" prefix.
//
//go:embed all:dist
var Dist embed.FS
