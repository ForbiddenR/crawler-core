package web

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func NewSPAHandler(staticFS fs.FS) (http.Handler, error) {
	sub, err := fs.Sub(staticFS, "dist")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(r.URL.Path, "/")
		if requestPath == "" {
			requestPath = "index.html"
		}

		f, err := sub.Open(requestPath)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		if path.Ext(requestPath) == "" {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	}), nil
}
