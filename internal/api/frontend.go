package api

import (
	"io/fs"
	"net/http"
	"strings"
)

func frontendHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if f, err := distFS.Open(path[1:]); err == nil {
			f.Close()

			// Hashed assets (e.g. /assets/index-ImR1482X.js) can be cached forever.
			// index.html must always be revalidated so deploys take effect immediately.
			if strings.HasPrefix(path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}

			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		w.Header().Set("Cache-Control", "no-cache")
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
