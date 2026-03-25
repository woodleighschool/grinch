package synchttp

import (
	"context"
	"log/slog"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
)

type Service interface {
	HandlePreflight(context.Context, uuid.UUID, *syncv1.PreflightRequest) (*syncv1.PreflightResponse, error)
	HandleEventUpload(context.Context, uuid.UUID, *syncv1.EventUploadRequest) (*syncv1.EventUploadResponse, error)
	HandleRuleDownload(context.Context, uuid.UUID, *syncv1.RuleDownloadRequest) (*syncv1.RuleDownloadResponse, error)
	HandlePostflight(context.Context, uuid.UUID, *syncv1.PostflightRequest) (*syncv1.PostflightResponse, error)
}

type Handler struct {
	logger  *slog.Logger
	service Service
}

func New(logger *slog.Logger, service Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}
