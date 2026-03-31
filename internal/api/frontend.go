package api

import (
	"io/fs"
	"net/http"
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
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
