package synchttp

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/subtle"
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
	protobufContentType    = "application/x-protobuf"
	maxRequestBodyBytes    = 16 << 20
	sharedSecretHeaderName = "X-Grinch-Shared-Secret" //nolint:gosec // fixed internal protocol header name, not a credential value.
)

var errSyncUnauthorized = errors.New("sync authentication failed")

func (handler *Handler) preflight(writer http.ResponseWriter, request *http.Request) {
	message := &syncv1.PreflightRequest{}
	if err := handler.decodeRequest(request, message); err != nil {
		handler.fail(writer, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(request, "machine_id"))
	if err != nil {
		handler.fail(writer, err)
		return
	}
	response, err := handler.service.HandlePreflight(request.Context(), machineID, message)
	if err != nil {
		handler.fail(writer, err)
		return
	}
	handler.writeResponse(writer, response)
}

func (handler *Handler) eventUpload(writer http.ResponseWriter, request *http.Request) {
	message := &syncv1.EventUploadRequest{}
	if err := handler.decodeRequest(request, message); err != nil {
		handler.fail(writer, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(request, "machine_id"))
	if err != nil {
		handler.fail(writer, err)
		return
	}
	response, err := handler.service.HandleEventUpload(request.Context(), machineID, message)
	if err != nil {
		handler.fail(writer, err)
		return
	}
	handler.writeResponse(writer, response)
}

func (handler *Handler) ruleDownload(writer http.ResponseWriter, request *http.Request) {
	message := &syncv1.RuleDownloadRequest{}
	if err := handler.decodeRequest(request, message); err != nil {
		handler.fail(writer, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(request, "machine_id"))
	if err != nil {
		handler.fail(writer, err)
		return
	}
	response, err := handler.service.HandleRuleDownload(request.Context(), machineID, message)
	if err != nil {
		handler.fail(writer, err)
		return
	}
	handler.writeResponse(writer, response)
}

func (handler *Handler) postflight(writer http.ResponseWriter, request *http.Request) {
	message := &syncv1.PostflightRequest{}
	if err := handler.decodeRequest(request, message); err != nil {
		handler.fail(writer, err)
		return
	}
	machineID, err := parseMachineID(chi.URLParam(request, "machine_id"))
	if err != nil {
		handler.fail(writer, err)
		return
	}
	response, err := handler.service.HandlePostflight(request.Context(), machineID, message)
	if err != nil {
		handler.fail(writer, err)
		return
	}
	handler.writeResponse(writer, response)
}

// Service captures the sync stage behavior used by HTTP handlers.
type Service interface {
	HandlePreflight(context.Context, uuid.UUID, *syncv1.PreflightRequest) (*syncv1.PreflightResponse, error)
	HandleEventUpload(context.Context, uuid.UUID, *syncv1.EventUploadRequest) (*syncv1.EventUploadResponse, error)
	HandleRuleDownload(context.Context, uuid.UUID, *syncv1.RuleDownloadRequest) (*syncv1.RuleDownloadResponse, error)
	HandlePostflight(context.Context, uuid.UUID, *syncv1.PostflightRequest) (*syncv1.PostflightResponse, error)
}

// Handler serves syncv1 stage endpoints with proto+gzip transport.
type Handler struct {
	service      Service
	sharedSecret string
}

// New returns a sync handler that can register stage routes on an existing chi router.
func New(service Service, sharedSecret string) *Handler {
	return &Handler{
		service:      service,
		sharedSecret: sharedSecret,
	}
}

// RegisterRoutes registers /sync stage endpoints onto the provided router.
func (handler *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/preflight/{machine_id}", handler.preflight)
	router.Post("/eventupload/{machine_id}", handler.eventUpload)
	router.Post("/ruledownload/{machine_id}", handler.ruleDownload)
	router.Post("/postflight/{machine_id}", handler.postflight)
}

func (handler *Handler) decodeRequest(request *http.Request, message proto.Message) error {
	if !handler.authenticate(request) {
		return errSyncUnauthorized
	}

	reader, err := gzip.NewReader(request.Body)
	if err != nil {
		return fmt.Errorf("%w: new gzip reader: %w", appsanta.ErrInvalidSyncRequest, err)
	}
	defer reader.Close()

	payload, err := io.ReadAll(io.LimitReader(reader, maxRequestBodyBytes))
	if err != nil {
		return fmt.Errorf("%w: read request body: %w", appsanta.ErrInvalidSyncRequest, err)
	}
	if unmarshalErr := proto.Unmarshal(payload, message); unmarshalErr != nil {
		return fmt.Errorf("%w: unmarshal proto: %w", appsanta.ErrInvalidSyncRequest, unmarshalErr)
	}

	return nil
}

func (handler *Handler) writeResponse(writer http.ResponseWriter, message proto.Message) {
	payload, err := marshalGzippedProto(message)
	if err != nil {
		writeSyncError(writer, http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", protobufContentType)
	writer.Header().Set("Content-Encoding", "gzip")
	writer.WriteHeader(http.StatusOK)
	//nolint:gosec // Writing the buffered protobuf response is the intended /sync success path.
	_, _ = writer.Write(payload)
}

func (handler *Handler) authenticate(request *http.Request) bool {
	if handler.sharedSecret == "" {
		return true
	}

	headerValue := request.Header.Get(sharedSecretHeaderName)
	return subtle.ConstantTimeCompare([]byte(headerValue), []byte(handler.sharedSecret)) == 1
}

func (handler *Handler) fail(writer http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	switch {
	case errors.Is(err, errSyncUnauthorized):
		statusCode = http.StatusUnauthorized
	case errors.Is(err, appsanta.ErrInvalidSyncRequest):
		statusCode = http.StatusBadRequest
	}
	writeSyncError(writer, statusCode)
}

func writeSyncError(writer http.ResponseWriter, statusCode int) {
	writer.Header().Del("Content-Type")
	writer.Header().Del("Content-Encoding")
	writer.WriteHeader(statusCode)
}

func parseMachineID(raw string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: parse machine_id %q: %w", appsanta.ErrInvalidSyncRequest, raw, err)
	}

	return parsed, nil
}

func marshalGzippedProto(message proto.Message) ([]byte, error) {
	payload, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, writeErr := writer.Write(payload); writeErr != nil {
		return nil, writeErr
	}
	if closeErr := writer.Close(); closeErr != nil {
		return nil, closeErr
	}

	return buffer.Bytes(), nil
}
