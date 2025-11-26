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

// Compiler turns rule rows + assignments into Santa wire representations.
type Compiler struct{}

// NewCompiler creates a ready-to-use compiler.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// BuildPayload assembles the SyncPayload for a machine using precomputed assignments.
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
				ID:            row.ID,
				Name:          row.Name,
				Type:          RuleType(row.Type),
				Target:        row.Target,
				Scope:         RuleScopeGlobal,
				Action:        RuleActionAllow,
				CustomMsg:     meta.BlockMessage,
				CelExpression: meta.CelExpression,
				CreatedAt:     createdAt,
			})
			continue
		}
		matched := filterAssignments(machine, assignments[row.ID])
		for _, assignment := range matched {
			syncRules = append(syncRules, SyncRule{
				ID:            assignment.ScopeID,
				Name:          row.Name,
				Type:          RuleType(row.Type),
				Target:        row.Target,
				Scope:         scopeFromTargetType(assignment.TargetType),
				Action:        RuleAction(assignment.Action),
				CustomMsg:     meta.BlockMessage,
				CelExpression: meta.CelExpression,
				CreatedAt:     createdAt,
			})
		}
	}
	syncRules = normaliseSyncRules(syncRules)

	if len(syncRules) == 0 {
		// Santa ignores empty rule lists during sync, so we must provide at least one rule.
		// TODO: This rule is not removed when rules are re-added
		syncRules = append(syncRules, SyncRule{
			ID:            uuid.Nil,
			Name:          "NOOP",
			Type:          RuleTypeBinary,
			Target:        "0000000000000000000000000000000000000000000000000000000000000000",
			Scope:         RuleScopeGlobal,
			Action:        RuleActionAllow,
			CustomMsg:     "NOOP",
			CelExpression: "",
			CreatedAt:     time.Time{},
		})
	}

	return SyncPayload{Cursor: ComputeCursor(syncRules), Rules: syncRules}
}

// CompileAssignments expands scopes + metadata into the per-user assignments Santa expects.
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

// filterAssignments narrows assignments to those relevant to the machine's primary user.
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

// scopeFromTargetType normalises DB scope semantics to the API enum.
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

// legacyScopes keeps legacy metadata-based target lists compatible with scopes.
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

// normaliseAction ensures empty/unknown actions default to allow.
func normaliseAction(value string) RuleAction {
	switch RuleAction(value) {
	case RuleActionBlock:
		return RuleActionBlock
	case RuleActionCel:
		return RuleActionCel
	default:
		return RuleActionAllow
	}
}

// matchesUUID compares a UUID to a nullable pgtype.UUID.
func matchesUUID(id uuid.UUID, candidate pgtype.UUID) bool {
	return candidate.Valid && id == candidate.Bytes
}

// uuidToPgtype wraps a UUID in the pgtype helper struct.
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// ComputeCursor hashes the rule list to create a stable sync cursor.
func ComputeCursor(rules []SyncRule) string {
	sum := sha256.New()
	for _, r := range rules {
		sum.Write(r.ID[:])
		sum.Write([]byte(r.Target))
		sum.Write([]byte(r.Name))
		sum.Write([]byte(r.Type))
		sum.Write([]byte(r.Scope))
		sum.Write([]byte(r.Action))
		sum.Write([]byte(r.CustomMsg))
		sum.Write([]byte(r.CelExpression))
		sum.Write([]byte(r.CreatedAt.UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(sum.Sum(nil))
}

// normaliseSyncRules sorts the rule list for deterministic cursors.
func normaliseSyncRules(rules []SyncRule) []SyncRule {
	out := append([]SyncRule(nil), rules...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Target == out[j].Target {
			if out[i].Action == out[j].Action {
				return out[i].ID.String() < out[j].ID.String()
			}
			return out[i].Action < out[j].Action
		}
		return out[i].Target < out[j].Target
	})
	return out
}

// SerialiseMetadata converts raw JSON metadata into a map for API responses.
func SerialiseMetadata(meta []byte) map[string]any {
	if len(meta) == 0 {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(meta, &out)
	return out
}

// CollectGroupIDs enumerates all group IDs that appear across metadata + scopes.
func CollectGroupIDs(meta RuleMetadata, scopes []sqlc.RuleScope) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	var ids []uuid.UUID
	for _, gid := range meta.Groups {
		if _, ok := seen[gid]; ok {
			continue
		}
		seen[gid] = struct{}{}
		ids = append(ids, gid)
	}
	for _, scope := range scopes {
		if scope.TargetType != "group" {
			continue
		}
		gid := scope.TargetID
		if _, ok := seen[gid]; ok {
			continue
		}
		seen[gid] = struct{}{}
		ids = append(ids, gid)
	}
	return ids
}
