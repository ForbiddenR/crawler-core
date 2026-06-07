package web

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
)

// Register installs handlers that serve the embedded web UI (SPA) for any request not
// handled by a registered API route. It replaces the previous nginx setup, where nginx
// served the SPA and proxied /api to the backend.
//
// It is a no-op when the UI was not embedded (built without the "embed_ui" build tag),
// leaving the server API-only.
//
// The API is mounted under the API base path (default "/api"); everything else is
// treated as a UI request. The SPA uses hash-mode routing, so the browser only ever
// requests "/", index.html, and static assets — but we still fall back to index.html
// for any unknown non-API path to be safe.
func Register(engine *gin.Engine) {
	distFS, ok := DistFS()
	if !ok {
		return
	}

	httpFS := http.FS(distFS)
	apiPath := utils.GetApiPath()

	engine.NoRoute(func(c *gin.Context) {
		// Only GET/HEAD can map to UI assets; anything else is a genuine API miss.
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}

		// Never serve the SPA for unmatched API paths — return a real 404 instead.
		reqPath := c.Request.URL.Path
		if reqPath == apiPath || strings.HasPrefix(reqPath, apiPath+"/") {
			c.Status(http.StatusNotFound)
			return
		}

		// Serve the requested file from the embedded FS if it exists, otherwise fall
		// back to index.html so client-side routing can take over.
		name := strings.TrimPrefix(reqPath, "/")
		if name != "" {
			if f, err := distFS.Open(name); err == nil {
				_ = f.Close()
				c.FileFromFS(reqPath, httpFS)
				return
			}
		}
		serveIndex(c, distFS)
	})
}

func serveIndex(c *gin.Context, distFS fs.FS) {
	data, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
