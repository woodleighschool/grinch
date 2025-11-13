package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) BuildPayload(machine sqlc.Machine, ruleRows []sqlc.Rule, assignmentRows []sqlc.RuleAssignment) SyncPayload {
	assignments := map[uuid.UUID][]sqlc.RuleAssignment{}
	for _, a := range assignmentRows {
		assignments[a.RuleID] = append(assignments[a.RuleID], a)
	}
	var syncRules []SyncRule
	for _, row := range ruleRows {
		if !row.Enabled {
			continue
		}
		meta, _ := ParseMetadata(row.Metadata)
		var createdAt time.Time
		if row.CreatedAt.Valid {
			createdAt = row.CreatedAt.Time
		}
		scope := RuleScope(row.Scope)
		if scope == RuleScopeGlobal {
			syncRules = append(syncRules, SyncRule{
				ID:        row.ID,
				Name:      row.Name,
				Type:      RuleType(row.Type),
				Target:    row.Target,
				Scope:     RuleScopeGlobal,
				Action:    RuleActionAllow,
				CustomMsg: meta.Description,
				CreatedAt: createdAt,
			})
			continue
		}
		matched := filterAssignments(machine, assignments[row.ID])
		for _, assignment := range matched {
			syncRules = append(syncRules, SyncRule{
				ID:        assignment.ScopeID,
				Name:      row.Name,
				Type:      RuleType(row.Type),
				Target:    row.Target,
				Scope:     scopeFromTargetType(assignment.TargetType),
				Action:    RuleAction(assignment.Action),
				CustomMsg: meta.Description,
				CreatedAt: createdAt,
			})
		}
	}
	return SyncPayload{Cursor: ComputeCursor(ruleRows), Rules: syncRules}
}

func (c *Compiler) CompileAssignments(rule sqlc.Rule, scopes []sqlc.RuleScope, meta RuleMetadata, groupMembers map[uuid.UUID][]uuid.UUID) []sqlc.InsertRuleAssignmentParams {
	normalised := append([]sqlc.RuleScope(nil), scopes...)
	if len(meta.Users) > 0 || len(meta.Groups) > 0 {
		normalised = append(normalised, legacyScopes(rule.ID, meta)...)
	}

	type key string
	seen := map[key]struct{}{}
	var assignments []sqlc.InsertRuleAssignmentParams

	add := func(k key, param sqlc.InsertRuleAssignmentParams) {
		if _, exists := seen[k]; exists {
			return
		}
		seen[k] = struct{}{}
		assignments = append(assignments, param)
	}

	for _, scope := range normalised {
		targetType := scope.TargetType
		action := normaliseAction(scope.Action)
		switch targetType {
		case "user":
			add(key("user:"+scope.TargetID.String()+":"+string(action)), sqlc.InsertRuleAssignmentParams{
				RuleID:     rule.ID,
				ScopeID:    scope.ID,
				TargetType: targetType,
				Action:     string(action),
				UserID:     uuidToPgtype(scope.TargetID),
			})
		case "group":
			members := groupMembers[scope.TargetID]
			for _, userID := range members {
				add(key("group:"+scope.TargetID.String()+":user:"+userID.String()+":"+string(action)), sqlc.InsertRuleAssignmentParams{
					RuleID:     rule.ID,
					ScopeID:    scope.ID,
					TargetType: targetType,
					Action:     string(action),
					GroupID:    uuidToPgtype(scope.TargetID),
					UserID:     uuidToPgtype(userID),
				})
			}
		}
	}

	return assignments
}

func filterAssignments(machine sqlc.Machine, assignments []sqlc.RuleAssignment) []sqlc.RuleAssignment {
	if !machine.UserID.Valid {
		return nil
	}
	userID := uuid.UUID(machine.UserID.Bytes)
	var matched []sqlc.RuleAssignment
	for _, assignment := range assignments {
		if matchesUUID(userID, assignment.UserID) {
			matched = append(matched, assignment)
		}
	}
	return matched
}

func scopeFromTargetType(targetType string) RuleScope {
	switch targetType {
	case "group":
		return RuleScopeGroup
	case "user":
		return RuleScopeUser
	default:
		return RuleScopeUser
	}
}

func legacyScopes(ruleID uuid.UUID, meta RuleMetadata) []sqlc.RuleScope {
	var scopes []sqlc.RuleScope
	for _, userID := range meta.Users {
		scopes = append(scopes, sqlc.RuleScope{
			ID:         uuid.New(),
			RuleID:     ruleID,
			TargetType: "user",
			TargetID:   userID,
			Action:     string(RuleActionAllow),
		})
	}
	for _, groupID := range meta.Groups {
		scopes = append(scopes, sqlc.RuleScope{
			ID:         uuid.New(),
			RuleID:     ruleID,
			TargetType: "group",
			TargetID:   groupID,
			Action:     string(RuleActionAllow),
		})
	}
	return scopes
}

func normaliseAction(value string) RuleAction {
	switch RuleAction(value) {
	case RuleActionBlock:
		return RuleActionBlock
	default:
		return RuleActionAllow
	}
}

func matchesUUID(id uuid.UUID, candidate pgtype.UUID) bool {
	return candidate.Valid && id == candidate.Bytes
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func ComputeCursor(rules []sqlc.Rule) string {
	type tuple struct {
		ID    string
		Stamp string
	}
	parts := make([]tuple, 0, len(rules))
	for _, r := range rules {
		var stamp string
		if r.UpdatedAt.Valid {
			stamp = r.UpdatedAt.Time.UTC().Format(time.RFC3339Nano)
		}
		parts = append(parts, tuple{ID: r.ID.String(), Stamp: stamp})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].ID < parts[j].ID })
	sum := sha256.New()
	for _, p := range parts {
		sum.Write([]byte(p.ID))
		sum.Write([]byte(p.Stamp))
	}
	return hex.EncodeToString(sum.Sum(nil))
}

func SerialiseMetadata(meta []byte) map[string]any {
	if len(meta) == 0 {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(meta, &out)
	return out
}
