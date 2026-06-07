//go:build !embed_ui

package web

import "io/fs"

// DistFS returns no filesystem because the web UI was not embedded (the binary was
// built without the "embed_ui" build tag). In this mode the server runs API-only and
// Register is a no-op, so plain `go build ./...` works without a frontend build.
func DistFS() (fs.FS, bool) {
	return nil, false
}
