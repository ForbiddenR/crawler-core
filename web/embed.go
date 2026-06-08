package web

import "embed"

// StaticAssets contains the embedded web frontend assets.
//
//go:embed dist
var StaticAssets embed.FS
