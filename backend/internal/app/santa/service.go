// Package santa owns the sync stage flow for trusted Santa clients.
package santa

import (
	"errors"
	"log/slog"

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
