package rules

import (
	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
)

// Rule describes a reusable rule template that can be applied to policies.
type Rule struct {
	ID                  uuid.UUID       `json:"id"`
	Name                string          `json:"name"                  validate:"required"`
	Description         string          `json:"description"`
	Identifier          string          `json:"identifier"            validate:"required"`
	RuleType            syncv1.RuleType `json:"rule_type"`
	CustomMsg           string          `json:"custom_msg"`
	CustomURL           string          `json:"custom_url"`
	NotificationAppName string          `json:"notification_app_name"`
}
