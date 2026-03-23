package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	machineRuleTargetKeySeparator  = "|"
	machineRuleTargetHashSeparator = "\x1f"
)

// ExecutionRuleCounts contains counts of execution rules by rule type.
type ExecutionRuleCounts struct {
	Binary      int32
	Certificate int32
	TeamID      int32
	SigningID   int32
	CDHash      int32
}

// MachineRuleTargetKey returns the stable key for a machine rule target.
func MachineRuleTargetKey(target MachineRuleTarget) string {
	return string(target.RuleType) + machineRuleTargetKeySeparator + target.Identifier
}

// CountExecutionRules counts machine rule targets by rule type.
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

// MachineRuleTargetPayloadHash returns a stable hash of the target payload fields.
func MachineRuleTargetPayloadHash(target MachineRuleTarget) string {
	sum := sha256.Sum256([]byte(machineRuleTargetHashInput(target)))
	return hex.EncodeToString(sum[:])
}

func machineRuleTargetHashInput(target MachineRuleTarget) string {
	return strings.Join([]string{
		string(target.RuleType),
		target.Identifier,
		string(target.Policy),
		target.CustomMessage,
		target.CustomURL,
		target.CELExpression,
	}, machineRuleTargetHashSeparator)
}
