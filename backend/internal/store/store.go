package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// RuleFilter narrows the rule listing in the admin API.
type RuleFilter struct {
	Search     string
	RuleType   string
	Identifier string
	Enabled    *bool
}

func (s *Store) ListUsers(ctx context.Context, search string) ([]sqlc.User, error) {
	return s.queries.ListUsers(ctx, strings.TrimSpace(search))
}

func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return s.queries.GetUser(ctx, id)
}

func (s *Store) GetUserByUPN(ctx context.Context, upn string) (sqlc.User, error) {
	return s.queries.GetUserByUPN(ctx, upn)
}

func (s *Store) GetUserByLogin(ctx context.Context, login string) (sqlc.User, error) {
	return s.queries.GetUserByLogin(ctx, login)
}

func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteUser(ctx, id)
}

func (s *Store) DeleteUserByUPN(ctx context.Context, upn string) error {
	return s.queries.DeleteUserByUPN(ctx, upn)
}

func (s *Store) UpsertUser(ctx context.Context, user sqlc.UpsertUserParams) (sqlc.User, error) {
	return s.queries.UpsertUser(ctx, user)
}

func (s *Store) ListGroups(ctx context.Context, search string) ([]sqlc.Group, error) {
	return s.queries.ListGroups(ctx, strings.TrimSpace(search))
}

func (s *Store) GetGroup(ctx context.Context, id uuid.UUID) (sqlc.Group, error) {
	return s.queries.GetGroup(ctx, id)
}

func (s *Store) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]sqlc.Group, error) {
	return s.queries.GetUserGroups(ctx, userID)
}

func (s *Store) UpsertGroup(ctx context.Context, group sqlc.UpsertGroupParams) (sqlc.Group, error) {
	return s.queries.UpsertGroup(ctx, group)
}

func (s *Store) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteGroup(ctx, id)
}

// ReplaceGroupMembers swaps the entire membership set inside a transaction.
func (s *Store) ReplaceGroupMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	return s.WithTx(ctx, func(tx pgx.Tx) error {
		queries := sqlc.New(tx)
		if err := queries.DeleteGroupMembers(ctx, groupID); err != nil {
			return err
		}
		for _, userID := range userIDs {
			if err := queries.AddGroupMember(ctx, sqlc.AddGroupMemberParams{
				GroupID: groupID,
				UserID:  userID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) ListGroupMemberIDs(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	return s.queries.ListGroupMemberIDs(ctx, groupID)
}

func (s *Store) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]sqlc.User, error) {
	return s.queries.ListGroupMembers(ctx, groupID)
}

func (s *Store) ListMachines(ctx context.Context, limit, offset int32, search string) ([]sqlc.Machine, error) {
	return s.queries.ListMachines(ctx, sqlc.ListMachinesParams{
		Limit:  limit,
		Offset: offset,
		Search: strings.TrimSpace(search),
	})
}

func (s *Store) GetMachine(ctx context.Context, id uuid.UUID) (sqlc.Machine, error) {
	return s.queries.GetMachine(ctx, id)
}

func (s *Store) GetUserMachines(ctx context.Context, userID pgtype.UUID) ([]sqlc.Machine, error) {
	return s.queries.GetUserMachines(ctx, userID)
}

func (s *Store) UpsertMachine(ctx context.Context, params sqlc.UpsertMachineParams) (sqlc.Machine, error) {
	return s.queries.UpsertMachine(ctx, params)
}

func (s *Store) GetMachineByIdentifier(ctx context.Context, machineIdentifier string) (sqlc.Machine, error) {
	return s.queries.GetMachineByIdentifier(ctx, machineIdentifier)
}

func (s *Store) UpdateMachinePreflightState(ctx context.Context, params sqlc.UpdateMachinePreflightStateParams) (sqlc.Machine, error) {
	return s.queries.UpdateMachinePreflightState(ctx, params)
}

func (s *Store) UpdateMachinePostflightState(ctx context.Context, params sqlc.UpdateMachinePostflightStateParams) (sqlc.Machine, error) {
	return s.queries.UpdateMachinePostflightState(ctx, params)
}

// TouchMachine updates the machine heartbeat + cursors whenever the agent calls home.
func (s *Store) TouchMachine(ctx context.Context, machineID uuid.UUID, clientVersion string, syncCursor, ruleCursor string) (sqlc.Machine, error) {
	return s.queries.UpdateMachineSyncState(ctx, sqlc.UpdateMachineSyncStateParams{
		ID:         machineID,
		LastSeen:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		SyncCursor: pgtype.Text{String: syncCursor, Valid: syncCursor != ""},
		RuleCursor: pgtype.Text{String: ruleCursor, Valid: ruleCursor != ""},
	})
}

func (s *Store) CreateRule(ctx context.Context, rule sqlc.CreateRuleParams) (sqlc.Rule, error) {
	return s.queries.CreateRule(ctx, rule)
}

func (s *Store) GetRule(ctx context.Context, id uuid.UUID) (sqlc.Rule, error) {
	return s.queries.GetRule(ctx, id)
}

func (s *Store) GetRuleByTarget(ctx context.Context, target string) (sqlc.Rule, error) {
	return s.queries.GetRuleByTarget(ctx, target)
}

func (s *Store) UpdateRule(ctx context.Context, rule sqlc.UpdateRuleParams) (sqlc.Rule, error) {
	return s.queries.UpdateRule(ctx, rule)
}

func (s *Store) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteRule(ctx, id)
}

func (s *Store) ListRules(ctx context.Context) ([]sqlc.Rule, error) {
	return s.queries.ListRules(ctx)
}

// FilterRules applies search + status filters for the applications page.
func (s *Store) FilterRules(ctx context.Context, filter RuleFilter) ([]sqlc.Rule, error) {
	var enabled pgtype.Bool
	if filter.Enabled != nil {
		enabled = pgtype.Bool{Bool: *filter.Enabled, Valid: true}
	}
	return s.queries.FilterRules(ctx, sqlc.FilterRulesParams{
		Search:     strings.TrimSpace(filter.Search),
		RuleType:   strings.TrimSpace(filter.RuleType),
		Identifier: strings.TrimSpace(filter.Identifier),
		Enabled:    enabled,
	})
}

func (s *Store) ListRuleScopes(ctx context.Context, ruleID uuid.UUID) ([]sqlc.RuleScope, error) {
	return s.queries.ListRuleScopes(ctx, ruleID)
}

func (s *Store) ListRulesByGroupTarget(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	return s.queries.ListRulesByGroupTarget(ctx, groupID)
}

func (s *Store) ListAllRuleScopes(ctx context.Context) ([]sqlc.RuleScope, error) {
	return s.queries.ListAllRuleScopes(ctx)
}

func (s *Store) GetRuleScope(ctx context.Context, scopeID uuid.UUID) (sqlc.RuleScope, error) {
	return s.queries.GetRuleScope(ctx, scopeID)
}

func (s *Store) GetRuleScopeByTarget(ctx context.Context, ruleID uuid.UUID, targetType string, targetID uuid.UUID) (sqlc.RuleScope, error) {
	return s.queries.GetRuleScopeByTarget(ctx, sqlc.GetRuleScopeByTargetParams{
		RuleID:     ruleID,
		TargetType: targetType,
		TargetID:   targetID,
	})
}

func (s *Store) CreateRuleScope(ctx context.Context, params sqlc.InsertRuleScopeParams) (sqlc.RuleScope, error) {
	return s.queries.InsertRuleScope(ctx, params)
}

func (s *Store) DeleteRuleScope(ctx context.Context, scopeID uuid.UUID) error {
	return s.queries.DeleteRuleScope(ctx, scopeID)
}

// ReplaceRuleAssignments atomically refreshes all compiled assignments for a rule.
func (s *Store) ReplaceRuleAssignments(ctx context.Context, ruleID uuid.UUID, assignments []sqlc.InsertRuleAssignmentParams) error {
	return s.WithTx(ctx, func(tx pgx.Tx) error {
		queries := sqlc.New(tx)
		if err := queries.DeleteAssignmentsByRule(ctx, ruleID); err != nil {
			return err
		}
		for _, assignment := range assignments {
			if err := queries.InsertRuleAssignment(ctx, assignment); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) ListRuleAssignments(ctx context.Context, ruleID uuid.UUID) ([]sqlc.RuleAssignment, error) {
	return s.queries.ListRuleAssignments(ctx, ruleID)
}

func (s *Store) ListAllAssignments(ctx context.Context) ([]sqlc.RuleAssignment, error) {
	return s.queries.ListAllAssignments(ctx)
}

func (s *Store) ListUserAssignments(ctx context.Context, userID uuid.UUID) ([]sqlc.ListUserAssignmentsRow, error) {
	return s.queries.ListUserAssignments(ctx, userID)
}

func (s *Store) ListApplicationAssignmentStats(ctx context.Context) ([]sqlc.ListApplicationAssignmentStatsRow, error) {
	return s.queries.ListApplicationAssignmentStats(ctx)
}

func (s *Store) RequestCleanSyncAllMachines(ctx context.Context) error {
	return s.queries.RequestCleanSyncAllMachines(ctx)
}

func (s *Store) RequestCleanSyncForUser(ctx context.Context, userID uuid.UUID) error {
	return s.queries.RequestCleanSyncForUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
}

func (s *Store) RequestCleanSyncForGroup(ctx context.Context, groupID uuid.UUID) error {
	return s.queries.RequestCleanSyncForGroup(ctx, groupID)
}

func (s *Store) InsertEvent(ctx context.Context, params sqlc.InsertEventParams) (sqlc.Event, error) {
	return s.queries.InsertEvent(ctx, params)
}

func (s *Store) ListEvents(ctx context.Context, limit, offset int32) ([]sqlc.ListEventSummariesRow, error) {
	return s.queries.ListEventSummaries(ctx, sqlc.ListEventSummariesParams{Limit: limit, Offset: offset})
}

func (s *Store) ListBlocksByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListBlocksByUserRow, error) {
	return s.queries.ListBlocksByUser(ctx, userID)
}

// SummariseEvents returns aggregate counts used for dashboards.
func (s *Store) SummariseEvents(ctx context.Context, days int32) ([]sqlc.SummariseEventsRow, error) {
	if days <= 0 {
		days = 14
	}
	return s.queries.SummariseEvents(ctx, days)
}

// MarshalRuleMetadata encodes optional metadata while tolerating nil values.
func (s *Store) MarshalRuleMetadata(meta any) ([]byte, error) {
	if meta == nil {
		return nil, nil
	}
	return json.Marshal(meta)
}

func (s *Store) UpsertFile(ctx context.Context, params sqlc.UpsertFileParams) error {
	return s.queries.UpsertFile(ctx, params)
}

func (s *Store) GetFile(ctx context.Context, sha256 string) (sqlc.File, error) {
	return s.queries.GetFile(ctx, sha256)
}

func (s *Store) ListFiles(ctx context.Context, limit, offset int32) ([]sqlc.ListFilesRow, error) {
	return s.queries.ListFiles(ctx, sqlc.ListFilesParams{Limit: limit, Offset: offset})
}
