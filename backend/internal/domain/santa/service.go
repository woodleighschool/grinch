// Package santa implements the Santa client sync protocol.
package santa

import (
	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/domain/rules"
)

// SyncService coordinates Santa sync operations across domain services.
type SyncService struct {
	machines machines.Service
	policies policies.Service
	rules    rules.Service
	events   events.Service
}

// NewSyncService constructs a SyncService with its required dependencies.
func NewSyncService(
	machinesSvc machines.Service,
	policiesSvc policies.Service,
	rulesSvc rules.Service,
	eventsSvc events.Service,
) SyncService {
	return SyncService{
		machines: machinesSvc,
		policies: policiesSvc,
		rules:    rulesSvc,
		events:   eventsSvc,
	}
}
