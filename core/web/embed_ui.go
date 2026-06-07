//go:build embed_ui

// Package web serves the embedded web UI (single-page app) from the Gin server.
//
// This file is only compiled when building with the "embed_ui" build tag, e.g.:
//
//	go build -tags embed_ui ./...
//
// The built SPA must be present in core/web/dist at build time (copy the output of
// `pnpm run build` from frontend/crawlab-ui into core/web/dist before building).
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFiles embed.FS

// DistFS returns the embedded web UI filesystem rooted at the dist directory.
func DistFS() (fs.FS, bool) {
	sub, err := fs.Sub(distFiles, "dist")
	if err != nil {
		return nil, false
	}
	return sub, true
}
