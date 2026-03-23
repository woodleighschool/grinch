package synchttp

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
)

const (
	protobufContentType = "application/x-protobuf"
	maxRequestBodyBytes = 16 << 20
)

type Service interface {
	HandlePreflight(context.Context, uuid.UUID, *syncv1.PreflightRequest) (*syncv1.PreflightResponse, error)
	HandleEventUpload(context.Context, uuid.UUID, *syncv1.EventUploadRequest) (*syncv1.EventUploadResponse, error)
	HandleRuleDownload(context.Context, uuid.UUID, *syncv1.RuleDownloadRequest) (*syncv1.RuleDownloadResponse, error)
	HandlePostflight(context.Context, uuid.UUID, *syncv1.PostflightRequest) (*syncv1.PostflightResponse, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/preflight/{machine_id}", h.preflight)
	r.Post("/eventupload/{machine_id}", h.eventUpload)
	r.Post("/ruledownload/{machine_id}", h.ruleDownload)
	r.Post("/postflight/{machine_id}", h.postflight)
}

func (h *Handler) preflight(w http.ResponseWriter, r *http.Request) {
	msg := &syncv1.PreflightRequest{}
	if err := h.decodeRequest(r, msg); err != nil {
		h.writeError(w, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(r, "machine_id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	resp, err := h.service.HandlePreflight(r.Context(), machineID, msg)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeProtoResponse(w, resp)
}

func (h *Handler) eventUpload(w http.ResponseWriter, r *http.Request) {
	msg := &syncv1.EventUploadRequest{}
	if err := h.decodeRequest(r, msg); err != nil {
		h.writeError(w, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(r, "machine_id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	resp, err := h.service.HandleEventUpload(r.Context(), machineID, msg)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeProtoResponse(w, resp)
}

func (h *Handler) ruleDownload(w http.ResponseWriter, r *http.Request) {
	msg := &syncv1.RuleDownloadRequest{}
	if err := h.decodeRequest(r, msg); err != nil {
		h.writeError(w, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(r, "machine_id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	resp, err := h.service.HandleRuleDownload(r.Context(), machineID, msg)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeProtoResponse(w, resp)
}

func (h *Handler) postflight(w http.ResponseWriter, r *http.Request) {
	msg := &syncv1.PostflightRequest{}
	if err := h.decodeRequest(r, msg); err != nil {
		h.writeError(w, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(r, "machine_id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	resp, err := h.service.HandlePostflight(r.Context(), machineID, msg)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeProtoResponse(w, resp)
}

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

func (h *Handler) writeProtoResponse(w http.ResponseWriter, msg proto.Message) {
	payload, err := marshalCompressedProto(msg)
	if err != nil {
		writeStatusOnly(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", protobufContentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(http.StatusOK)

	//nolint:gosec // Buffered protobuf response body is the intended sync success path.
	_, _ = w.Write(payload)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	writeStatusOnly(w, statusCodeForError(err))
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

func parseMachineID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: parse machine_id %q: %w", appsanta.ErrInvalidSyncRequest, raw, err)
	}

	return id, nil
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
