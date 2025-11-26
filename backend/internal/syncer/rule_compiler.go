package syncer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// NewRuleCompilerJob periodically recomputes rule assignments for every rule.
func NewRuleCompilerJob(store *store.Store, compiler *rules.Compiler, logger *slog.Logger) Job {
	return func(ctx context.Context) error {
		if store == nil || compiler == nil {
			return fmt.Errorf("rule compiler prerequisites missing")
		}
		allRules, err := store.ListRules(ctx)
		if err != nil {
			return fmt.Errorf("list rules: %w", err)
		}
		allScopes, err := store.ListAllRuleScopes(ctx)
		if err != nil {
			return fmt.Errorf("list rule scopes: %w", err)
		}
		scopeIndex := map[uuid.UUID][]sqlc.RuleScope{}
		for _, scope := range allScopes {
			scopeIndex[scope.RuleID] = append(scopeIndex[scope.RuleID], scope)
		}
		groupCache := map[uuid.UUID][]uuid.UUID{}
		for _, rule := range allRules {
			meta, err := rules.ParseMetadata(rule.Metadata)
			if err != nil {
				logger.Warn("skip rule metadata", "rule", rule.ID, "err", err)
				continue
			}
			ruleScopes := scopeIndex[rule.ID]
			memberMap := map[uuid.UUID][]uuid.UUID{}
			for _, groupID := range rules.CollectGroupIDs(meta, ruleScopes) {
				if _, ok := groupCache[groupID]; !ok {
					members, err := store.ListGroupMemberIDs(ctx, groupID)
					if err != nil {
						logger.Warn("list group members", "group", groupID, "err", err)
						groupCache[groupID] = nil
					} else {
						groupCache[groupID] = members
					}
				}
				memberMap[groupID] = groupCache[groupID]
				if len(groupCache[groupID]) == 0 {
					logger.Debug("group has no members during rule compile", "rule", rule.ID, "group", groupID)
				}
			}
			assignments := compiler.CompileAssignments(rule, ruleScopes, meta, memberMap)
			if err := store.ReplaceRuleAssignments(ctx, rule.ID, assignments); err != nil {
				logger.Error("replace rule assignments", "rule", rule.ID, "err", err)
			}
		}
		logger.Debug("rule compiler job complete", "rules", len(allRules))
		return nil
	}
}
