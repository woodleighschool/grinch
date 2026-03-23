package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-pkgz/rest"
)

const (
	frontendIndexFile = "index.html"
	frontendAssetsURL = "/assets/"

	cacheControlImmutable = "public, max-age=31536000, immutable"
	cacheControlNoCache   = "no-cache"
)

func mountFrontend(r chi.Router, dir string) {
	if !frontendExists(dir) {
		return
	}

	fileServer, err := rest.NewFileServer("/", dir, rest.FsOptSPA)
	if err != nil {
		return
	}

	frontend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setFrontendCacheHeaders(w, r)
		fileServer.ServeHTTP(w, r)
	})

	r.NotFound(
		rest.Gzip(
			"text/html",
			"text/css",
			"application/javascript",
			"application/json",
			"image/svg+xml",
		)(frontend).ServeHTTP,
	)
}

func frontendExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, frontendIndexFile))
	return err == nil
}

func setFrontendCacheHeaders(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, frontendAssetsURL) {
		w.Header().Set("Cache-Control", cacheControlImmutable)
		return
	}

	w.Header().Set("Cache-Control", cacheControlNoCache)
}
