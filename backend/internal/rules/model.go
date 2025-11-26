package rules

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RuleType describes the Santa rule primitive (binary, signing ID, etc).
type RuleType string

// RuleScope indicates which target the rule applies to.
type RuleScope string

const (
	RuleTypeAllow       RuleType = "allow"
	RuleTypeBlock       RuleType = "block"
	RuleTypeBinary      RuleType = "binary"
	RuleTypeCertificate RuleType = "certificate"
	RuleTypeSigningID   RuleType = "signing_id"

	RuleScopeGlobal RuleScope = "global"
	RuleScopeGroup  RuleScope = "group"
	RuleScopeUser   RuleScope = "user"
)

// RuleAction instructs Santa whether to allow, block, or evaluate CEL.
type RuleAction string

const (
	RuleActionAllow RuleAction = "allow"
	RuleActionBlock RuleAction = "block"
	RuleActionCel   RuleAction = "cel"
)

// RuleMetadata encodes the optional, user-facing properties on a rule.
type RuleMetadata struct {
	Description   string      `json:"description"`
	BlockMessage  string      `json:"block_message"`
	CelEnabled    bool        `json:"cel_enabled"`
	CelExpression string      `json:"cel_expression"`
	Users         []uuid.UUID `json:"users"`
	Groups        []uuid.UUID `json:"groups"`
}

// SyncRule is the canonical representation sent to Santa agents.
type SyncRule struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Type          RuleType   `json:"type"`
	Target        string     `json:"target"`
	Scope         RuleScope  `json:"scope"`
	Action        RuleAction `json:"action"`
	CustomMsg     string     `json:"custom_msg"`
	CelExpression string     `json:"cel_expression"`
	CreatedAt     time.Time  `json:"created_at"`
}

// SyncPayload contains the rules and cursor Santa expects.
type SyncPayload struct {
	Cursor string     `json:"cursor"`
	Rules  []SyncRule `json:"rules"`
}

// ParseMetadata unpacks optional JSON metadata stored for a rule.
func ParseMetadata(raw []byte) (RuleMetadata, error) {
	if len(raw) == 0 {
		return RuleMetadata{}, nil
	}
	var wire struct {
		Description   string   `json:"description"`
		BlockMessage  string   `json:"block_message"`
		CelEnabled    bool     `json:"cel_enabled"`
		CelExpression string   `json:"cel_expression"`
		Users         []string `json:"users"`
		Groups        []string `json:"groups"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return RuleMetadata{}, err
	}
	return RuleMetadata{
		Description:   strings.TrimSpace(wire.Description),
		BlockMessage:  strings.TrimSpace(wire.BlockMessage),
		CelEnabled:    wire.CelEnabled,
		CelExpression: strings.TrimSpace(wire.CelExpression),
		Users:         parseUUIDs(wire.Users),
		Groups:        parseUUIDs(wire.Groups),
	}, nil
}

// parseUUIDs filters out invalid identifiers without failing parsing.
func parseUUIDs(values []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(values))
	for _, val := range values {
		if val == "" {
			continue
		}
		if id, err := uuid.Parse(val); err == nil {
			out = append(out, id)
		}
	}
	return out
}
