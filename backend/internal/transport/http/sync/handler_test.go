package synchttp_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/woodleighschool/grinch/internal/app/santa"
	synchttp "github.com/woodleighschool/grinch/internal/transport/http/sync"
)

type stubService struct {
	preflightErr error
}

func (stub stubService) HandlePreflight(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.PreflightRequest,
) (*syncv1.PreflightResponse, error) {
	if stub.preflightErr != nil {
		return nil, stub.preflightErr
	}
	syncType := syncv1.SyncType_NORMAL
	return syncv1.PreflightResponse_builder{
		ClientMode: syncv1.ClientMode_MONITOR,
		SyncType:   &syncType,
	}.Build(), nil
}

func (stubService) HandleEventUpload(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.EventUploadRequest,
) (*syncv1.EventUploadResponse, error) {
	return syncv1.EventUploadResponse_builder{}.Build(), nil
}

func (stubService) HandleRuleDownload(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.RuleDownloadRequest,
) (*syncv1.RuleDownloadResponse, error) {
	return syncv1.RuleDownloadResponse_builder{}.Build(), nil
}

func (stubService) HandlePostflight(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.PostflightRequest,
) (*syncv1.PostflightResponse, error) {
	return syncv1.PostflightResponse_builder{}.Build(), nil
}

func TestPreflight_ReturnsGzippedProtoResponse(t *testing.T) {
	t.Parallel()

	handler := synchttp.New(stubService{})
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	requestMessage := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
		Hostname:  "host1",
	}.Build()
	body := mustGzipProto(t, requestMessage)

	request := httptest.NewRequest(
		http.MethodPost,
		"/preflight/00000000-0000-0000-0000-000000000001",
		body,
	)
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Content-Encoding", "gzip")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if got := response.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip response, got %q", got)
	}

	decoded := &syncv1.PreflightResponse{}
	mustUnzipProto(t, response.Body.Bytes(), decoded)
	if decoded.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("expected sync type NORMAL, got %v", decoded.GetSyncType())
	}
}

func TestPreflight_ReturnsBadRequestForNonGzipBody(t *testing.T) {
	t.Parallel()

	handler := synchttp.New(stubService{})
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	requestMessage := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
	}.Build()
	payload, err := proto.Marshal(requestMessage)
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/preflight/00000000-0000-0000-0000-000000000001",
		bytes.NewReader(payload),
	)
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Content-Encoding", "identity")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
	assertSyncFailureResponse(t, response)
}

func TestPreflight_ReturnsBadRequestForInvalidSyncRequest(t *testing.T) {
	t.Parallel()

	handler := synchttp.New(stubService{preflightErr: santa.ErrInvalidSyncRequest})
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	requestMessage := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
	}.Build()
	body := mustGzipProto(t, requestMessage)

	request := httptest.NewRequest(
		http.MethodPost,
		"/preflight/00000000-0000-0000-0000-000000000001",
		body,
	)
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Content-Encoding", "gzip")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
	assertSyncFailureResponse(t, response)
}

func TestPreflight_ReturnsInternalServerErrorForUnexpectedServiceError(t *testing.T) {
	t.Parallel()

	handler := synchttp.New(stubService{preflightErr: errors.New("boom")})
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	requestMessage := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
	}.Build()
	body := mustGzipProto(t, requestMessage)

	request := httptest.NewRequest(
		http.MethodPost,
		"/preflight/00000000-0000-0000-0000-000000000001",
		body,
	)
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Content-Encoding", "gzip")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}
	assertSyncFailureResponse(t, response)
}

func mustGzipProto(t *testing.T, message proto.Message) io.Reader {
	t.Helper()

	payload, err := proto.Marshal(message)
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, writeErr := writer.Write(payload); writeErr != nil {
		t.Fatalf("write gzip payload: %v", writeErr)
	}
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("close gzip writer: %v", closeErr)
	}

	return bytes.NewReader(buffer.Bytes())
}

func mustUnzipProto(t *testing.T, payload []byte, message proto.Message) {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create gzip reader: %v", err)
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip payload: %v", err)
	}

	if unmarshalErr := proto.Unmarshal(decoded, message); unmarshalErr != nil {
		t.Fatalf("unmarshal proto: %v", unmarshalErr)
	}
}

func assertSyncFailureResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "" {
		t.Fatalf("Content-Type = %q, want empty", got)
	}
	if got := response.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q, want empty", got)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("response body length = %d, want 0", response.Body.Len())
	}
}
