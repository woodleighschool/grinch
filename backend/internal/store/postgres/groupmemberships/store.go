package groupmemberships

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
	"github.com/woodleighschool/grinch/internal/store/postgres"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

type Store struct {
	store *postgres.Store
}

func New(store *postgres.Store) *Store {
	return &Store{store: store}
}

func (store *Store) ListGroupMemberships(
	ctx context.Context,
	options domain.GroupMembershipListOptions,
) ([]domain.GroupMembershipListItem, int32, error) {
	orderBy, err := pgutil.OrderBy(
		options.Sort,
		options.Order,
		groupMembershipSortColumns(),
		[]string{"group_name ASC", "member_name ASC", "membership_sort_id ASC"},
	)
	if err != nil {
		return nil, 0, err
	}

	query := groupMembershipListQuery(orderBy)

	rows, err := store.store.Pool().Query(ctx, query, groupMembershipListArguments(options)...)
	if err != nil {
		return nil, 0, err
	}

	return pgutil.CollectRows(rows, scanGroupMembership)
}

func groupMembershipListQuery(orderBy string) string {
	return fmt.Sprintf(`
WITH actual_memberships AS (
  SELECT
    gm.id::text AS membership_sort_id,
    'actual'::text AS membership_kind,
    gm.id AS actual_membership_id,
    ''::text AS effective_membership_id,
    g.id AS group_id,
    g.name AS group_name,
    g.source AS group_source,
    gm.member_kind,
    gm.member_id,
    CASE
      WHEN gm.member_kind = 'user' THEN NULLIF(u.display_name, '')
      ELSE NULLIF(m.hostname, '')
    END AS member_name,
    gm.created_at,
    gm.updated_at
  FROM group_memberships AS gm
  JOIN groups AS g ON g.id = gm.group_id
  LEFT JOIN users AS u
    ON gm.member_kind = 'user'
    AND u.id = gm.member_id
  LEFT JOIN machines AS m
    ON gm.member_kind = 'machine'
    AND m.machine_id = gm.member_id
  WHERE ($1::uuid IS NULL OR gm.group_id = $1::uuid)
    AND ($2::uuid IS NULL OR (gm.member_kind = 'user' AND gm.member_id = $2::uuid))
    AND ($3::uuid IS NULL OR (gm.member_kind = 'machine' AND gm.member_id = $3::uuid))
),
effective_machine_memberships AS (
  SELECT
    CONCAT('effective-machine:', $3::text, ':', g.id::text) AS membership_sort_id,
    'effective'::text AS membership_kind,
    NULL::uuid AS actual_membership_id,
    CONCAT('effective-machine:', $3::text, ':', g.id::text) AS effective_membership_id,
    g.id AS group_id,
    g.name AS group_name,
    g.source AS group_source,
    'user'::text AS member_kind,
    u.id AS member_id,
    NULLIF(u.display_name, '') AS member_name,
    gm.created_at,
    gm.updated_at
  FROM machines AS machine
  JOIN users AS u
    ON u.upn = machine.primary_user
  JOIN group_memberships AS gm
    ON gm.member_kind = 'user'
    AND gm.member_id = u.id
  JOIN groups AS g
    ON g.id = gm.group_id
  WHERE $3::uuid IS NOT NULL
    AND machine.machine_id = $3::uuid
    AND machine.primary_user <> ''
    AND ($1::uuid IS NULL OR gm.group_id = $1::uuid)
),
membership_rows AS (
  SELECT * FROM actual_memberships
  UNION ALL
  SELECT * FROM effective_machine_memberships
)
SELECT
  actual_membership_id,
  effective_membership_id,
  membership_kind,
  group_id,
  group_name,
  group_source,
  member_kind,
  member_id,
  member_name,
  created_at,
  updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM membership_rows
WHERE ($4 = '' OR
  group_name ILIKE $4 OR
  COALESCE(member_name, '') ILIKE $4)
ORDER BY %s
LIMIT NULLIF($5::INT, 0)
OFFSET $6
`, orderBy)
}

func groupMembershipSortColumns() map[string]string {
	return map[string]string{
		"id":          "membership_sort_id",
		"group_name":  "group_name",
		"kind":        "membership_kind",
		"member_name": "member_name",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	}
}

func groupMembershipListArguments(options domain.GroupMembershipListOptions) []any {
	return []any{
		pgutil.NullableUUID(options.GroupID),
		pgutil.NullableUUID(options.UserID),
		pgutil.NullableUUID(options.MachineID),
		pgutil.SearchPattern(options.Search),
		options.Limit,
		options.Offset,
	}
}

func scanGroupMembership(rows pgx.Rows) (domain.GroupMembershipListItem, int32, error) {
	var (
		item               domain.GroupMembershipListItem
		membershipKindText string
		groupSourceText    string
		memberKindText     string
		total              int32
	)

	if scanErr := rows.Scan(
		&item.ActualMembershipID,
		&item.EffectiveMembershipID,
		&membershipKindText,
		&item.Group.ID,
		&item.Group.Name,
		&groupSourceText,
		&memberKindText,
		&item.Member.ID,
		&item.Member.Name,
		&item.CreatedAt,
		&item.UpdatedAt,
		&total,
	); scanErr != nil {
		return domain.GroupMembershipListItem{}, 0, scanErr
	}

	groupSourceValue, sourceErr := domain.ParsePrincipalSource(groupSourceText)
	if sourceErr != nil {
		return domain.GroupMembershipListItem{}, 0, sourceErr
	}
	memberKindValue, kindErr := domain.ParseMemberKind(memberKindText)
	if kindErr != nil {
		return domain.GroupMembershipListItem{}, 0, kindErr
	}

	item.Group.Source = groupSourceValue
	item.Member.Kind = memberKindValue
	item.Kind = domain.GroupMembershipKind(membershipKindText)
	return item, total, nil
}

func (store *Store) GetGroupMembership(ctx context.Context, id uuid.UUID) (domain.GroupMembership, error) {
	row, err := store.store.Queries().GetPersistedGroupMembershipView(ctx, id)
	if err != nil {
		return domain.GroupMembership{}, err
	}
	return mapPersistedGroupMembership(row)
}

func (store *Store) CreateGroupMembership(
	ctx context.Context,
	groupID uuid.UUID,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
	origin domain.GroupMembershipOrigin,
) (domain.GroupMembership, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.GroupMembership{}, err
	}

	row, err := store.store.Queries().CreateGroupMembership(ctx, db.CreateGroupMembershipParams{
		ID:         id,
		GroupID:    groupID,
		MemberKind: string(memberKind),
		MemberID:   memberID,
		Origin:     string(origin),
	})
	if err != nil {
		return domain.GroupMembership{}, err
	}
	return store.GetGroupMembership(ctx, row.ID)
}

func (store *Store) DeleteGroupMembership(ctx context.Context, id uuid.UUID) error {
	_, err := store.store.Queries().DeleteGroupMembership(ctx, id)
	return err
}

func (store *Store) GetGroup(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	return pgutil.GetGroup(ctx, store.store.Queries(), id)
}

func mapPersistedGroupMembership(row db.GetPersistedGroupMembershipViewRow) (domain.GroupMembership, error) {
	groupSource, err := domain.ParsePrincipalSource(row.GroupSource)
	if err != nil {
		return domain.GroupMembership{}, err
	}
	memberKind, err := domain.ParseMemberKind(row.MemberKind)
	if err != nil {
		return domain.GroupMembership{}, err
	}
	memberName, err := membershipMemberName(row.MemberName)
	if err != nil {
		return domain.GroupMembership{}, err
	}

	return domain.GroupMembership{
		ID:   row.ID,
		Kind: domain.GroupMembershipKindActual,
		Group: domain.GroupMembershipGroup{
			ID:     row.GroupID,
			Name:   row.GroupName,
			Source: groupSource,
		},
		Member: domain.GroupMembershipMember{
			Kind: memberKind,
			ID:   row.MemberID,
			Name: memberName,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func membershipMemberName(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	default:
		return "", fmt.Errorf("unsupported membership member_name type %T", value)
	}
}
