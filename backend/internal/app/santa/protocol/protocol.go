// Package protocol translates Grinch's frozen sync snapshot into sync protobuf
// messages without carrying any persistence concerns itself.
package protocol

import (
	"fmt"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	"github.com/woodleighschool/grinch/internal/app/santa/model"
	"github.com/woodleighschool/grinch/internal/domain"
)

// BuildRuleDownloadResponse maps the already-prepared payload into the Santa
// wire shape. Selection and diffing happen earlier in the snapshot package.
func BuildRuleDownloadResponse(rules []model.SyncRule) (*syncv1.RuleDownloadResponse, error) {
	protoRules := make([]*syncv1.Rule, 0, len(rules))
	for _, rule := range rules {
		protoRule, err := buildRule(rule)
		if err != nil {
			return nil, err
		}
		protoRules = append(protoRules, protoRule)
	}

	return syncv1.RuleDownloadResponse_builder{Rules: protoRules}.Build(), nil
}

func MapStoredRuleSyncType(value model.RuleSyncType) (syncv1.SyncType, error) {
	switch value {
	case model.RuleSyncTypeNone:
		return syncv1.SyncType_SYNC_TYPE_UNSPECIFIED, nil
	case model.RuleSyncTypeNormal:
		return syncv1.SyncType_NORMAL, nil
	case model.RuleSyncTypeClean:
		return syncv1.SyncType_CLEAN, nil
	default:
		return syncv1.SyncType_SYNC_TYPE_UNSPECIFIED, fmt.Errorf("unsupported rule sync type %q", value)
	}
}

func buildRule(rule model.SyncRule) (*syncv1.Rule, error) {
	ruleType, err := mapRuleType(rule.RuleType)
	if err != nil {
		return nil, err
	}

	policy, err := mapPolicy(rule)
	if err != nil {
		return nil, err
	}

	ruleBuilder := syncv1.Rule_builder{
		Identifier: rule.Identifier,
		Policy:     policy,
		RuleType:   ruleType,
		CustomMsg:  rule.CustomMessage,
		CustomUrl:  rule.CustomURL,
	}

	if policy == syncv1.Policy_CEL {
		ruleBuilder.CelExpr = rule.CELExpression
	}

	return ruleBuilder.Build(), nil
}

func mapRuleType(value domain.RuleType) (syncv1.RuleType, error) {
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

func mapPolicy(rule model.SyncRule) (syncv1.Policy, error) {
	if rule.Removed {
		return syncv1.Policy_REMOVE, nil
	}

	value := rule.Policy
	switch value {
	case domain.RulePolicyAllowlist:
		return syncv1.Policy_ALLOWLIST, nil
	case domain.RulePolicyBlocklist:
		return syncv1.Policy_BLOCKLIST, nil
	case domain.RulePolicySilentBlocklist:
		return syncv1.Policy_SILENT_BLOCKLIST, nil
	case domain.RulePolicyCEL:
		return syncv1.Policy_CEL, nil
	default:
		return syncv1.Policy_POLICY_UNKNOWN, fmt.Errorf("unsupported policy %q", value)
	}
}
