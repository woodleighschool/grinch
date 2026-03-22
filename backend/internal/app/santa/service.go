// Package santa owns the sync stage flow for trusted Santa clients.
//
// It wires the stage handlers together and delegates sync
// planning and protobuf mapping to subpackages.
package santa

import (
	"errors"
	"log/slog"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
)

// ErrInvalidSyncRequest indicates invalid sync request data.
var ErrInvalidSyncRequest = errors.New("invalid sync request")

// Service is the stage-facing app service used by the /sync transport.
type Service struct {
	logger         *slog.Logger
	dataStore      model.DataStore
	eventAllowlist map[domain.EventDecision]struct{}
	ruleResolver   model.RuleResolver
}

// New creates a sync service with explicit dependencies.
func New(
	logger *slog.Logger,
	dataStore model.DataStore,
	eventAllowlist []domain.EventDecision,
	ruleResolver model.RuleResolver,
) *Service {
	allowlist := make(map[domain.EventDecision]struct{}, len(eventAllowlist))
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
