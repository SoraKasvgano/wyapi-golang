package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type spaHandler struct {
	fs        fs.FS
	indexFile string
}

func NewSPAHandler(assetFS fs.FS, indexFile string) http.Handler {
	if indexFile == "" {
		indexFile = "index.html"
	}
	return &spaHandler{fs: assetFS, indexFile: indexFile}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requested := strings.TrimPrefix(r.URL.Path, "/")
	if requested == "" {
		requested = h.indexFile
	}

	if !fileExists(h.fs, requested) {
		requested = h.indexFile
	}

	if requested == h.indexFile {
		data, err := fs.ReadFile(h.fs, h.indexFile)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(data)
		return
	}

	fileServer := http.FileServer(http.FS(h.fs))
	r2 := *r
	r2.URL.Path = "/" + path.Clean(requested)
	fileServer.ServeHTTP(w, &r2)
}

func fileExists(assetFS fs.FS, name string) bool {
	if assetFS == nil {
		return false
	}
	if _, err := fs.Stat(assetFS, name); err == nil {
		return true
	}
	return false
}
