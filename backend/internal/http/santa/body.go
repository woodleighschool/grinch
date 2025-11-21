package santa

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// decodeBody wraps the request body based on Content-Encoding.
func decodeBody(r *http.Request) (io.ReadCloser, error) {
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Encoding"))) {
	case "", "identity":
		return r.Body, nil
	case "gzip":
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		return reader, nil
	case "deflate":
		return flate.NewReader(r.Body), nil
	default:
		return nil, fmt.Errorf("unsupported content encoding %q", r.Header.Get("Content-Encoding"))
	}
}
