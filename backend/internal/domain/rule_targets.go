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

type AppliedRuleTarget struct {
	RuleType    RuleType `json:"rule_type"`
	Identifier  string   `json:"identifier"`
	PayloadHash string   `json:"payload_hash"`
}

type PendingRuleTarget struct {
	MachineRuleTarget

	PayloadHash string `json:"payload_hash"`
}

type ExecutionRuleCounts struct {
	Binary      int32
	Certificate int32
	TeamID      int32
	SigningID   int32
	CDHash      int32
}

func MachineRuleTargetKey(target MachineRuleTarget) string {
	return string(target.RuleType) + "|" + target.Identifier
}

func CountExecutionRules(targets []MachineRuleTarget) ExecutionRuleCounts {
	var counts ExecutionRuleCounts
	for _, target := range targets {
		switch target.RuleType {
		case RuleTypeBinary:
			counts.Binary++
		case RuleTypeCertificate:
			counts.Certificate++
		case RuleTypeTeamID:
			counts.TeamID++
		case RuleTypeSigningID:
			counts.SigningID++
		case RuleTypeCDHash:
			counts.CDHash++
		}
	}

	return counts
}

func MachineRuleTargetPayloadHash(target MachineRuleTarget) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		string(target.RuleType),
		target.Identifier,
		string(target.Policy),
		target.CustomMessage,
		target.CustomURL,
		target.CELExpression,
	}, "\x1f")))
	return hex.EncodeToString(sum[:])
}
