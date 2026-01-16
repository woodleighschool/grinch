package sync

import (
	"context"
	"strconv"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/core/policies"
	corerules "github.com/woodleighschool/grinch/internal/core/rules"
)

const noopIdentifier = "0000000000000000000000000000000000000000000000000000000000000000"

func upToDate(machine coremachines.Machine, policyRulesVersion int32) bool {
	if policyRulesVersion == 0 || machine.AppliedRulesVersion == nil {
		return false
	}
	return policyRulesVersion == *machine.AppliedRulesVersion
}

func (s *Service) ensureStatefulNoop(ctx context.Context, policyID uuid.UUID) (*syncv1.RuleDownloadResponse, error) {
	atts, err := s.policies.ListPolicyRuleAttachmentsForSyncByPolicyID(ctx, policyID, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(atts) == 0 {
		return noopResponse(), nil
	}
	return &syncv1.RuleDownloadResponse{}, nil
}

func (s *Service) loadRules(
	ctx context.Context,
	atts []policies.PolicyAttachment,
) (map[uuid.UUID]corerules.Rule, error) {
	ids := uniqueRuleIDs(atts)
	if len(ids) == 0 {
		return map[uuid.UUID]corerules.Rule{}, nil
	}

	list, err := s.rules.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]corerules.Rule, len(list))
	for _, r := range list {
		out[r.ID] = r
	}
	return out, nil
}

func buildRuleset(atts []policies.PolicyAttachment, rulesByID map[uuid.UUID]corerules.Rule) []*syncv1.Rule {
	if len(atts) == 0 {
		return nil
	}

	out := make([]*syncv1.Rule, 0, len(atts))
	for _, att := range atts {
		rule, ok := rulesByID[att.RuleID]
		if !ok {
			continue
		}

		result := &syncv1.Rule{
			Identifier:          rule.Identifier,
			Policy:              att.Action,
			RuleType:            rule.RuleType,
			CustomMsg:           rule.CustomMsg,
			CustomUrl:           rule.CustomURL,
			NotificationAppName: rule.NotificationAppName,
		}

		if att.Action == syncv1.Policy_CEL && att.CELExpr != nil {
			result.CelExpr = *att.CELExpr
		}

		out = append(out, result)
	}

	return out
}

func uniqueRuleIDs(atts []policies.PolicyAttachment) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(atts))
	var ids []uuid.UUID

	for _, att := range atts {
		if att.RuleID == uuid.Nil || seen[att.RuleID] {
			continue
		}
		seen[att.RuleID] = true
		ids = append(ids, att.RuleID)
	}
	return ids
}

func nextCursor(offset, count, pageSize int) string {
	if count == pageSize {
		return strconv.Itoa(offset + count)
	}
	return ""
}

func noopResponse() *syncv1.RuleDownloadResponse {
	return &syncv1.RuleDownloadResponse{
		Rules: []*syncv1.Rule{{
			Identifier: noopIdentifier,
			Policy:     syncv1.Policy_ALLOWLIST,
			RuleType:   syncv1.RuleType_BINARY,
		}},
	}
}
