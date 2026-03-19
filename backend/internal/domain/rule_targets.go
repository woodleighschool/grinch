package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

type StoredRuleTarget struct {
	MachineRuleTarget

	RuleID      *uuid.UUID `json:"rule_id,omitempty"`
	RuleName    string     `json:"rule_name"`
	PayloadHash string     `json:"payload_hash"`
}

func MachineRuleTargetKey(target MachineRuleTarget) string {
	return string(target.RuleType) + "|" + target.IdentifierKey
}

func MachineRuleTargetPayloadHash(target MachineRuleTarget) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		string(target.RuleType),
		target.IdentifierKey,
		target.Identifier,
		string(target.Policy),
		target.CustomMessage,
		target.CustomURL,
		target.CELExpression,
	}, "\x1f")))
	return hex.EncodeToString(sum[:])
}
