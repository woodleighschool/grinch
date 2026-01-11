// Package machines contains domain models and logic for Santa client machines.
package machines

import (
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
)

// PolicyStatus represents the state of a machine's policy assignment.
type PolicyStatus int16

const (
	PolicyStatusUnassigned PolicyStatus = 0
	PolicyStatusPending    PolicyStatus = 1
	PolicyStatusUpToDate   PolicyStatus = 2
)

// Machine represents a Santa client and its current policy and rule state.
type Machine struct {
	ID                   uuid.UUID         `json:"id"`
	SerialNumber         string            `json:"serial_number"`
	Hostname             string            `json:"hostname"`
	Model                string            `json:"model"`
	OSVersion            string            `json:"os_version"`
	OSBuild              string            `json:"os_build"`
	SantaVersion         string            `json:"santa_version"`
	PrimaryUser          *string           `json:"primary_user"`
	PrimaryUserGroups    []string          `json:"primary_user_groups"`
	PushToken            *string           `json:"push_token"`
	SIPStatus            int32             `json:"sip_status"`
	ClientMode           syncv1.ClientMode `json:"client_mode"`
	RequestCleanSync     bool              `json:"request_clean_sync"`
	PushNotificationSync bool              `json:"push_notification_sync"`
	BinaryRuleCount      int32             `json:"binary_rule_count"`
	CertificateRuleCount int32             `json:"certificate_rule_count"`
	CompilerRuleCount    int32             `json:"compiler_rule_count"`
	TransitiveRuleCount  int32             `json:"transitive_rule_count"`
	TeamIDRuleCount      int32             `json:"team_id_rule_count"`
	SigningIDRuleCount   int32             `json:"signing_id_rule_count"`
	CDHashRuleCount      int32             `json:"cdhash_rule_count"`
	RulesHash            *string           `json:"rules_hash"`

	// Server side tracking.
	UserID       *uuid.UUID   `json:"user_id"`
	LastSeen     time.Time    `json:"last_seen"`
	PolicyID     *uuid.UUID   `json:"policy_id"`
	PolicyStatus PolicyStatus `json:"policy_status"`

	// Applied policy versions acknowledged by the client.
	AppliedPolicyID        *uuid.UUID `json:"applied_policy_id"`
	AppliedSettingsVersion *int32     `json:"applied_settings_version"`
	AppliedRulesVersion    *int32     `json:"applied_rules_version"`
}

// ListItem represents a machine projection used by list endpoints.
type ListItem struct {
	ID                     uuid.UUID    `json:"id"`
	SerialNumber           string       `json:"serial_number"`
	Hostname               string       `json:"hostname"`
	Model                  string       `json:"model"`
	OSVersion              string       `json:"os_version"`
	PrimaryUser            *string      `json:"primary_user"`
	UserID                 *uuid.UUID   `json:"user_id"`
	LastSeen               time.Time    `json:"last_seen"`
	PolicyID               *uuid.UUID   `json:"policy_id"`
	PolicyStatus           PolicyStatus `json:"policy_status"`
	AppliedPolicyID        *uuid.UUID   `json:"applied_policy_id"`
	AppliedSettingsVersion *int32       `json:"applied_settings_version"`
	AppliedRulesVersion    *int32       `json:"applied_rules_version"`
}
