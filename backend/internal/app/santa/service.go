// Package santa owns the sync stage flow for trusted Santa clients.
package santa

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
)

// ErrInvalidSyncRequest indicates the request contains data the sync protocol cannot accept.
var ErrInvalidSyncRequest = errors.New("invalid sync request")

type Service struct {
	logger         *slog.Logger
	dataStore      model.DataStore
	eventAllowlist map[domain.ExecutionDecision]struct{}
	ruleResolver   model.RuleResolver
}

func New(
	logger *slog.Logger,
	dataStore model.DataStore,
	eventAllowlist []domain.ExecutionDecision,
	ruleResolver model.RuleResolver,
) *Service {
	allowlist := make(map[domain.ExecutionDecision]struct{}, len(eventAllowlist))
	for _, decision := range eventAllowlist {
		allowlist[decision] = struct{}{}
	}

	return &Service{
		logger:         logger,
		dataStore:      dataStore,
		eventAllowlist: allowlist,
		ruleResolver:   ruleResolver,
	}
}

func syncLogAttrs(ctx context.Context, machineID uuid.UUID, extra ...any) []any {
	attrs := []any{
		"request_id", middleware.GetReqID(ctx),
		"machine_id", machineID,
	}

	return append(attrs, extra...)
}
