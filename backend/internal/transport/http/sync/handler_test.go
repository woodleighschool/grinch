package synchttp_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"log/slog"
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

type testService struct {
	preflightErr error
}

func (s *testService) HandlePreflight(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.PreflightRequest,
) (*syncv1.PreflightResponse, error) {
	if s.preflightErr != nil {
		return nil, s.preflightErr
	}

	syncType := syncv1.SyncType_NORMAL
	return syncv1.PreflightResponse_builder{
		ClientMode: syncv1.ClientMode_MONITOR,
		SyncType:   &syncType,
	}.Build(), nil
}

func (*testService) HandleEventUpload(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.EventUploadRequest,
) (*syncv1.EventUploadResponse, error) {
	return syncv1.EventUploadResponse_builder{}.Build(), nil
}

func (*testService) HandleRuleDownload(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.RuleDownloadRequest,
) (*syncv1.RuleDownloadResponse, error) {
	return syncv1.RuleDownloadResponse_builder{}.Build(), nil
}

func (*testService) HandlePostflight(
	_ context.Context,
	_ uuid.UUID,
	_ *syncv1.PostflightRequest,
) (*syncv1.PostflightResponse, error) {
	return syncv1.PostflightResponse_builder{}.Build(), nil
}

func newTestRouter(service *testService) http.Handler {
	handler := synchttp.New(newTestLogger(), service)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	return router
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPreflight_ReturnsGzippedProtoResponse(t *testing.T) {
	router := newTestRouter(&testService{})

	req := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
		Hostname:  "host1",
	}.Build()
	body := mustGzipProto(t, req)

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
		t.Fatalf("Code = %d, want 200", response.Code)
	}
	if contentEncoding := response.Header().Get("Content-Encoding"); contentEncoding != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", contentEncoding)
	}

	resp := &syncv1.PreflightResponse{}
	mustUnzipProto(t, response.Body.Bytes(), resp)
	if resp.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("SyncType = %v, want NORMAL", resp.GetSyncType())
	}
}

func TestPreflight_ReturnsBadRequestForNonGzipBody(t *testing.T) {
	router := newTestRouter(&testService{})

	req := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
	}.Build()
	payload, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
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
		t.Fatalf("Code = %d, want 400", response.Code)
	}
	assertSyncFailureResponse(t, response)
}

func TestPreflight_ReturnsBadRequestForInvalidSyncRequest(t *testing.T) {
	router := newTestRouter(&testService{preflightErr: santa.ErrInvalidSyncRequest})

	req := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
	}.Build()
	body := mustGzipProto(t, req)

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
		t.Fatalf("Code = %d, want 400", response.Code)
	}
	assertSyncFailureResponse(t, response)
}

func TestPreflight_ReturnsInternalServerErrorForUnexpectedServiceError(t *testing.T) {
	router := newTestRouter(&testService{preflightErr: errors.New("boom")})

	req := syncv1.PreflightRequest_builder{
		MachineId: "00000000-0000-0000-0000-000000000001",
	}.Build()
	body := mustGzipProto(t, req)

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
		t.Fatalf("Code = %d, want 500", response.Code)
	}
	assertSyncFailureResponse(t, response)
}

func mustGzipProto(t *testing.T, message proto.Message) io.Reader {
	t.Helper()

	payload, err := proto.Marshal(message)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err = writer.Write(payload); err != nil {
		t.Fatalf("writer.Write() error = %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	return bytes.NewReader(buffer.Bytes())
}

func mustUnzipProto(t *testing.T, payload []byte, message proto.Message) {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if err = proto.Unmarshal(decoded, message); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
}

func assertSyncFailureResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := response.Header().Get("Content-Type"); contentType != "" {
		t.Fatalf("Content-Type = %q, want empty", contentType)
	}
	if contentEncoding := response.Header().Get("Content-Encoding"); contentEncoding != "" {
		t.Fatalf("Content-Encoding = %q, want empty", contentEncoding)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("Body.Len() = %d, want 0", response.Body.Len())
	}
}
