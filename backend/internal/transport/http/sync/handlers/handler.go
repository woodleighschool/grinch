package handlers

import (
	"context"

	"github.com/woodleighschool/grinch/internal/platform/logging"
	syncsvc "github.com/woodleighschool/grinch/internal/service/sync"
)

// Handler owns the Santa sync HTTP endpoints.
type Handler struct {
	sync *syncsvc.Service
	log  logging.Logger
}

// NewHandler builds a Handler.
func NewHandler(svc *syncsvc.Service, log logging.Logger) *Handler {
	return &Handler{sync: svc, log: log}
}

func (h *Handler) logger(ctx context.Context) logging.Logger {
	if ctx != nil {
		if log := logging.FromContext(ctx); log != nil {
			return log
		}
	}
	if h.log != nil {
		return h.log
	}
	return logging.FromContext(context.Background())
}
