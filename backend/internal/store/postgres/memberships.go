package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

func (store *Store) ListMemberships(
	ctx context.Context,
	options domain.MembershipListOptions,
) ([]domain.MembershipListItem, int32, error) {
	orderBy, err := orderBy(
		options.Sort,
		options.Order,
		membershipSortColumns(),
		[]string{"group_name ASC", "member_name ASC", "membership_sort_id ASC"},
	)
	if err != nil {
		return nil, 0, err
	}

	query := membershipListQuery(orderBy)

	rows, err := store.Pool().Query(ctx, query, membershipListArguments(options)...)
	if err != nil {
		return nil, 0, err
	}

	return collectRows(rows, scanMembership)
}

func membershipListQuery(orderBy string) string {
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

func membershipSortColumns() map[string]string {
	return map[string]string{
		"id":          "membership_sort_id",
		"group_name":  "group_name",
		"kind":        "membership_kind",
		"member_name": "member_name",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	}
}

func membershipListArguments(options domain.MembershipListOptions) []any {
	return []any{
		nullableUUID(options.GroupID),
		nullableUUID(options.UserID),
		nullableUUID(options.MachineID),
		searchPattern(options.Search),
		options.Limit,
		options.Offset,
	}
}

func scanMembership(rows pgx.Rows) (domain.MembershipListItem, int32, error) {
	var (
		item               domain.MembershipListItem
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
		return domain.MembershipListItem{}, 0, scanErr
	}

	groupSourceValue, sourceErr := domain.ParsePrincipalSource(groupSourceText)
	if sourceErr != nil {
		return domain.MembershipListItem{}, 0, sourceErr
	}
	memberKindValue, kindErr := domain.ParseMemberKind(memberKindText)
	if kindErr != nil {
		return domain.MembershipListItem{}, 0, kindErr
	}

	item.Group.Source = groupSourceValue
	item.Member.Kind = memberKindValue
	item.Kind = domain.MembershipKind(membershipKindText)
	return item, total, nil
}

func (store *Store) GetMembership(ctx context.Context, id uuid.UUID) (domain.Membership, error) {
	row, err := store.Queries().GetPersistedMembershipView(ctx, id)
	if err != nil {
		return domain.Membership{}, err
	}
	return mapPersistedMembership(row)
}

func (store *Store) CreateMembership(
	ctx context.Context,
	groupID uuid.UUID,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
	origin domain.MembershipOrigin,
) (domain.Membership, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Membership{}, err
	}

	err = store.RunInTx(ctx, func(queries *db.Queries) error {
		_, createErr := queries.CreateMembership(ctx, db.CreateMembershipParams{
			ID:         id,
			GroupID:    groupID,
			MemberKind: string(memberKind),
			MemberID:   memberID,
			Origin:     string(origin),
		})
		if createErr != nil {
			return createErr
		}
		return nil
	})
	if err != nil {
		return domain.Membership{}, err
	}

	return store.GetMembership(ctx, id)
}

func (store *Store) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	return store.RunInTx(ctx, func(queries *db.Queries) error {
		_, err := queries.DeleteMembership(ctx, id)
		return err
	})
}

func mapPersistedMembership(row db.GetPersistedMembershipViewRow) (domain.Membership, error) {
	groupSource, err := domain.ParsePrincipalSource(row.GroupSource)
	if err != nil {
		return domain.Membership{}, err
	}
	memberKind, err := domain.ParseMemberKind(row.MemberKind)
	if err != nil {
		return domain.Membership{}, err
	}

	return domain.Membership{
		ID:   row.ID,
		Kind: domain.MembershipKindActual,
		Group: domain.MembershipGroup{
			ID:     row.GroupID,
			Name:   row.GroupName,
			Source: groupSource,
		},
		Member: domain.MembershipMember{
			Kind: memberKind,
			ID:   row.MemberID,
			Name: row.MemberName,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}
