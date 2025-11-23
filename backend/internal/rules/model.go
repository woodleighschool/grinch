package rules

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RuleType string

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

type RuleAction string

const (
	RuleActionAllow RuleAction = "allow"
	RuleActionBlock RuleAction = "block"
)

type RuleMetadata struct {
	Description  string      `json:"description"`
	BlockMessage string      `json:"block_message"`
	Users        []uuid.UUID `json:"users"`
	Groups       []uuid.UUID `json:"groups"`
}

type SyncRule struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Type      RuleType   `json:"type"`
	Target    string     `json:"target"`
	Scope     RuleScope  `json:"scope"`
	Action    RuleAction `json:"action"`
	CustomMsg string     `json:"custom_msg"`
	CreatedAt time.Time  `json:"created_at"`
}

type SyncPayload struct {
	Cursor string     `json:"cursor"`
	Rules  []SyncRule `json:"rules"`
}

func ParseMetadata(raw []byte) (RuleMetadata, error) {
	if len(raw) == 0 {
		return RuleMetadata{}, nil
	}
	var wire struct {
		Description  string   `json:"description"`
		BlockMessage string   `json:"block_message"`
		Users        []string `json:"users"`
		Groups       []string `json:"groups"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return RuleMetadata{}, err
	}
	return RuleMetadata{
		Description:  strings.TrimSpace(wire.Description),
		BlockMessage: strings.TrimSpace(wire.BlockMessage),
		Users:        parseUUIDs(wire.Users),
		Groups:       parseUUIDs(wire.Groups),
	}, nil
}

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
