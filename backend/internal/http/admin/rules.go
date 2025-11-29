package admin

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// recompileRuleAssignments rebuilds compiled assignments for a single rule.
func (h Handler) recompileRuleAssignments(ctx context.Context, rule sqlc.Rule) {
	meta, err := rules.ParseMetadata(rule.Metadata)
	if err != nil {
		h.Logger.Warn("skip rule metadata", "rule", rule.ID, "err", err)
		return
	}
	ruleScopes, err := h.Store.ListRuleScopes(ctx, rule.ID)
	if err != nil {
		h.Logger.Error("list rule scopes", "rule", rule.ID, "err", err)
		return
	}
	memberMap := map[uuid.UUID][]uuid.UUID{}
	for _, groupID := range rules.CollectGroupIDs(meta, ruleScopes) {
		members, err := h.Store.ListGroupMemberIDs(ctx, groupID)
		if err != nil {
			h.Logger.Warn("list group members", "group", groupID, "err", err)
			continue
		}
		memberMap[groupID] = members
	}
	assignments := h.Compiler.CompileAssignments(rule, ruleScopes, meta, memberMap)
	if err := h.Store.ReplaceRuleAssignments(ctx, rule.ID, assignments); err != nil {
		h.Logger.Error("replace rule assignments", "rule", rule.ID, "err", err)
	}
}
