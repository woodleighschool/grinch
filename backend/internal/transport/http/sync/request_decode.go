package synchttp

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
)

const maxRequestBodyBytes = 16 << 20

func (h *Handler) decodeRequest(r *http.Request, msg proto.Message) error {
	gr, err := gzip.NewReader(r.Body)
	if err != nil {
		return fmt.Errorf("%w: new gzip reader: %w", appsanta.ErrInvalidSyncRequest, err)
	}
	defer gr.Close()

	payload, err := io.ReadAll(io.LimitReader(gr, maxRequestBodyBytes))
	if err != nil {
		return fmt.Errorf("%w: read request body: %w", appsanta.ErrInvalidSyncRequest, err)
	}

	if err = proto.Unmarshal(payload, msg); err != nil {
		return fmt.Errorf("%w: unmarshal proto: %w", appsanta.ErrInvalidSyncRequest, err)
	}

	return nil
}
