package santa

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/domain/rules"
)

// noopIdentifier is a SHA256 value that does not match any real binary.
const noopIdentifier = "0000000000000000000000000000000000000000000000000000000000000000"

const syncPageSize = 200

// RuleDownload handles the Santa rule download stage.
func (s SyncService) RuleDownload(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.RuleDownloadRequest,
) (*syncv1.RuleDownloadResponse, error) {
	machine, err := s.machines.Get(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: get machine: %w", err)
	}

	machine.LastSeen = time.Now().UTC()
	if _, err = s.machines.Upsert(ctx, machine); err != nil {
		return nil, fmt.Errorf("ruledownload: upsert machine: %w", err)
	}

	if machine.PolicyID == nil {
		return noopResponse(), nil
	}

	offset, err := parseCursor(req.GetCursor())
	if err != nil {
		return nil, fmt.Errorf("ruledownload: invalid cursor: %w", err)
	}

	policy, err := s.policies.Get(ctx, *machine.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: get policy: %w", err)
	}

	if isUpToDate(machine, policy.RulesVersion) {
		return s.upToDateResponse(ctx, *machine.PolicyID)
	}

	atts, err := s.policies.ListPolicyRuleAttachmentsForSyncByPolicyID(ctx, *machine.PolicyID, syncPageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: list attachments: %w", err)
	}

	if offset == 0 && len(atts) == 0 {
		return noopResponse(), nil
	}

	rulesByID, err := s.loadRules(ctx, atts)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: load rules: %w", err)
	}

	return &syncv1.RuleDownloadResponse{
		Rules:  buildRules(atts, rulesByID),
		Cursor: nextCursor(offset, len(atts)),
	}, nil
}

func parseCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}

	n, err := strconv.Atoi(cursor)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, errors.New("negative offset")
	}

	return n, nil
}

func isUpToDate(machine machines.Machine, policyRulesVersion int32) bool {
	if policyRulesVersion == 0 || machine.AppliedRulesVersion == nil {
		return false
	}
	return policyRulesVersion == *machine.AppliedRulesVersion
}

func (s SyncService) upToDateResponse(ctx context.Context, policyID uuid.UUID) (*syncv1.RuleDownloadResponse, error) {
	atts, err := s.policies.ListPolicyRuleAttachmentsForSyncByPolicyID(ctx, policyID, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("verify attachments: %w", err)
	}
	if len(atts) == 0 {
		return noopResponse(), nil
	}

	return &syncv1.RuleDownloadResponse{}, nil
}

func nextCursor(offset, got int) string {
	if got == syncPageSize {
		return strconv.Itoa(offset + got)
	}
	return ""
}

// noopResponse returns a rule set that clears state on clients that ignore empty rule lists.
func noopResponse() *syncv1.RuleDownloadResponse {
	return &syncv1.RuleDownloadResponse{
		Rules: []*syncv1.Rule{{
			Identifier: noopIdentifier,
			Policy:     syncv1.Policy_ALLOWLIST,
			RuleType:   syncv1.RuleType_BINARY,
		}},
	}
}

func (s SyncService) loadRules(ctx context.Context, atts []policies.Attachment) (map[uuid.UUID]rules.Rule, error) {
	ids := uniqueRuleIDs(atts)
	if len(ids) == 0 {
		return make(map[uuid.UUID]rules.Rule), nil
	}

	list, err := s.rules.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]rules.Rule, len(list))
	for _, r := range list {
		out[r.ID] = r
	}
	return out, nil
}

func uniqueRuleIDs(atts []policies.Attachment) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(atts))
	var out []uuid.UUID

	for _, a := range atts {
		if a.RuleID == uuid.Nil || seen[a.RuleID] {
			continue
		}
		seen[a.RuleID] = true
		out = append(out, a.RuleID)
	}

	return out
}

func buildRules(atts []policies.Attachment, rulesByID map[uuid.UUID]rules.Rule) []*syncv1.Rule {
	if len(atts) == 0 {
		return nil
	}

	out := make([]*syncv1.Rule, 0, len(atts))
	for _, att := range atts {
		r, ok := rulesByID[att.RuleID]
		if !ok {
			continue
		}

		sr := &syncv1.Rule{
			Identifier:          r.Identifier,
			Policy:              att.Action,
			RuleType:            r.RuleType,
			CustomMsg:           r.CustomMsg,
			CustomUrl:           r.CustomURL,
			NotificationAppName: r.NotificationAppName,
		}

		if att.Action == syncv1.Policy_CEL && att.CELExpr != nil {
			sr.CelExpr = *att.CELExpr
		}

		out = append(out, sr)
	}

	return out
}
