package santa

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// TODO: check if this actuall works for the 3 options
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
		return r.Body, nil
	}
}
