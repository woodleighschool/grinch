package snapshot

import (
	"fmt"
	"math"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
)

// BuildRuleDownloadResponse converts the prepared sync payload into the Santa
// wire format.
func BuildRuleDownloadResponse(rules []model.SyncRule) (*syncv1.RuleDownloadResponse, error) {
	protoRules := make([]*syncv1.Rule, 0, len(rules))
	for _, rule := range rules {
		protoRule, err := protoRuleFromSyncRule(rule)
		if err != nil {
			return nil, fmt.Errorf("build proto rule %s/%s: %w", rule.RuleType, rule.Identifier, err)
		}
		protoRules = append(protoRules, protoRule)
	}

	return syncv1.RuleDownloadResponse_builder{
		Rules: protoRules,
	}.Build(), nil
}

// SyncTypeFromPendingFullSync maps the stored pending sync mode to the Santa
// wire enum.
func SyncTypeFromPendingFullSync(fullSync bool) syncv1.SyncType {
	if fullSync {
		return syncv1.SyncType_CLEAN
	}
	return syncv1.SyncType_NORMAL
}

// MachineClientModeFromProto maps the Santa client mode enum to the internal
// machine client mode.
func MachineClientModeFromProto(value syncv1.ClientMode) domain.MachineClientMode {
	switch value {
	case syncv1.ClientMode_MONITOR:
		return domain.MachineClientModeMonitor
	case syncv1.ClientMode_LOCKDOWN:
		return domain.MachineClientModeLockdown
	case syncv1.ClientMode_STANDALONE:
		return domain.MachineClientModeStandalone
	case syncv1.ClientMode_UNKNOWN_CLIENT_MODE:
		fallthrough
	default:
		return domain.MachineClientModeUnknown
	}
}

// ClampRuleCount converts a uint32 rule count to int32 without overflow.
func ClampRuleCount(value uint32) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func protoRuleFromSyncRule(rule model.SyncRule) (*syncv1.Rule, error) {
	ruleType, err := protoRuleType(rule.RuleType)
	if err != nil {
		return nil, err
	}

	policy, err := protoPolicy(rule)
	if err != nil {
		return nil, err
	}

	builder := syncv1.Rule_builder{
		Identifier: rule.Identifier,
		Policy:     policy,
		RuleType:   ruleType,
		CustomMsg:  rule.CustomMessage,
		CustomUrl:  rule.CustomURL,
	}

	if policy == syncv1.Policy_CEL {
		builder.CelExpr = rule.CELExpression
	}

	return builder.Build(), nil
}

func protoRuleType(value domain.RuleType) (syncv1.RuleType, error) {
	switch value {
	case domain.RuleTypeBinary:
		return syncv1.RuleType_BINARY, nil
	case domain.RuleTypeCertificate:
		return syncv1.RuleType_CERTIFICATE, nil
	case domain.RuleTypeTeamID:
		return syncv1.RuleType_TEAMID, nil
	case domain.RuleTypeSigningID:
		return syncv1.RuleType_SIGNINGID, nil
	case domain.RuleTypeCDHash:
		return syncv1.RuleType_CDHASH, nil
	default:
		return syncv1.RuleType_RULETYPE_UNKNOWN, fmt.Errorf("unsupported rule type %q", value)
	}
}

func protoPolicy(rule model.SyncRule) (syncv1.Policy, error) {
	if rule.Removed {
		return syncv1.Policy_REMOVE, nil
	}

	switch rule.Policy {
	case domain.RulePolicyAllowlist:
		return syncv1.Policy_ALLOWLIST, nil
	case domain.RulePolicyBlocklist:
		return syncv1.Policy_BLOCKLIST, nil
	case domain.RulePolicySilentBlocklist:
		return syncv1.Policy_SILENT_BLOCKLIST, nil
	case domain.RulePolicyCEL:
		return syncv1.Policy_CEL, nil
	default:
		return syncv1.Policy_POLICY_UNKNOWN, fmt.Errorf("unsupported policy %q", rule.Policy)
	}
}
