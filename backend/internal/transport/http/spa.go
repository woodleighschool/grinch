package httprouter

import (
	"net/http"
	"strings"

	"github.com/go-pkgz/rest"
)

// FrontendConfig configures serving the frontend build.
type FrontendConfig struct {
	Dir        string
	EnableGzip bool
}

// NewFrontendHandler returns a handler that serves the frontend build from cfg.Dir.
func NewFrontendHandler(cfg FrontendConfig) http.Handler {
	fs, _ := rest.NewFileServer("/", cfg.Dir, rest.FsOptSPA)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fs.ServeHTTP(w, r)
	})

	if cfg.EnableGzip {
		handler = rest.Gzip(
			"text/html",
			"text/css",
			"application/javascript",
			"application/json",
			"image/svg+xml",
		)(handler)
	}

	return handler
}
