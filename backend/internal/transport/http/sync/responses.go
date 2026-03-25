package synchttp

import (
	"bytes"
	"compress/gzip"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/protobuf/proto"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
)

const protobufContentType = "application/x-protobuf"

func (h *Handler) writeProtoResponse(w http.ResponseWriter, r *http.Request, msg proto.Message) {
	payload, err := marshalCompressedProto(msg)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"sync response marshal failed",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"machine_id", chi.URLParam(r, "machine_id"),
			"error", err,
		)
		writeStatusOnly(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", protobufContentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(payload)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := statusCodeForError(err)
	args := []any{
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"machine_id", chi.URLParam(r, "machine_id"),
		"content_type", r.Header.Get("Content-Type"),
		"content_encoding", r.Header.Get("Content-Encoding"),
		"content_length", r.ContentLength,
		"user_agent", r.UserAgent(),
		"error", err,
	}

	switch {
	case statusCode >= http.StatusInternalServerError:
		h.logger.ErrorContext(r.Context(), "sync request failed", args...)
	case statusCode >= http.StatusBadRequest:
		h.logger.WarnContext(r.Context(), "sync request rejected", args...)
	default:
		h.logger.InfoContext(r.Context(), "sync request completed", args...)
	}

	writeStatusOnly(w, statusCode)
}

func statusCodeForError(err error) int {
	switch {
	case errors.Is(err, appsanta.ErrInvalidSyncRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func writeStatusOnly(w http.ResponseWriter, statusCode int) {
	w.Header().Del("Content-Type")
	w.Header().Del("Content-Encoding")
	w.WriteHeader(statusCode)
}

func marshalCompressedProto(msg proto.Message) ([]byte, error) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err = zw.Write(payload); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
