package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

var (
	membershipListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":          "id",
		"group_name":  "group_name",
		"member_name": "member_name",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	}

	membershipListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"group_name ASC",
		"member_name ASC",
		"id ASC",
	}
)

func (s *Store) ListMemberships(
	ctx context.Context,
	opts domain.MembershipListOptions,
) ([]domain.Membership, int32, error) {
	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		membershipListSortColumns,
		membershipListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(membershipListQuery, orderBy)

	rows, err := s.Pool().Query(ctx, query, membershipListArgs(opts)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list memberships: %w", err)
	}

	return collectRows(rows, scanMembershipRow)
}

func (s *Store) GetMembership(ctx context.Context, id uuid.UUID) (domain.Membership, error) {
	row, err := s.Queries().GetPersistedMembershipView(ctx, id)
	if err != nil {
		return domain.Membership{}, err
	}

	return mapMembership(row)
}

func (s *Store) CreateMembership(
	ctx context.Context,
	groupID uuid.UUID,
	memberKind domain.MemberKind,
	memberID uuid.UUID,
	origin domain.MembershipOrigin,
) (domain.Membership, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Membership{}, fmt.Errorf("create membership id: %w", err)
	}

	switch memberKind {
	case domain.MemberKindUser:
		_, err = s.Queries().CreateUserMembership(ctx, db.CreateUserMembershipParams{
			ID:      id,
			GroupID: groupID,
			UserID:  memberID,
			Origin:  db.MembershipOrigin(origin),
		})
	case domain.MemberKindMachine:
		_, err = s.Queries().CreateMachineMembership(ctx, db.CreateMachineMembershipParams{
			ID:        id,
			GroupID:   groupID,
			MachineID: memberID,
			Origin:    db.MembershipOrigin(origin),
		})
	default:
		return domain.Membership{}, fmt.Errorf("unsupported member kind %q", memberKind)
	}
	if err != nil {
		return domain.Membership{}, err
	}

	return s.GetMembership(ctx, id)
}

func (s *Store) DeleteMembership(ctx context.Context, id uuid.UUID, kind domain.MemberKind) error {
	switch kind {
	case domain.MemberKindUser:
		n, err := s.Queries().DeleteUserMembership(ctx, id)
		if err != nil {
			return err
		}
		if n == 0 {
			return pgx.ErrNoRows
		}
	case domain.MemberKindMachine:
		n, err := s.Queries().DeleteMachineMembership(ctx, id)
		if err != nil {
			return err
		}
		if n == 0 {
			return pgx.ErrNoRows
		}
	default:
		return fmt.Errorf("unsupported member kind %q", kind)
	}

	return nil
}

func membershipListArgs(opts domain.MembershipListOptions) []any {
	return []any{
		nullableUUID(opts.GroupID),
		nullableUUID(opts.UserID),
		nullableUUID(opts.MachineID),
		searchPattern(opts.Search),
		opts.Limit,
		opts.Offset,
	}
}

func scanMembershipRow(rows pgx.Rows) (domain.Membership, int32, error) {
	var (
		item            domain.Membership
		groupSourceText string
		memberKindText  string
		total           int32
	)

	if err := rows.Scan(
		&item.ID,
		&item.Group.ID,
		&item.Group.Name,
		&groupSourceText,
		&memberKindText,
		&item.Member.ID,
		&item.Member.Name,
		&item.CreatedAt,
		&item.UpdatedAt,
		&total,
	); err != nil {
		return domain.Membership{}, 0, err
	}

	groupSource, err := domain.ParsePrincipalSource(groupSourceText)
	if err != nil {
		return domain.Membership{}, 0, fmt.Errorf("parse group source: %w", err)
	}

	memberKind, err := domain.ParseMemberKind(memberKindText)
	if err != nil {
		return domain.Membership{}, 0, fmt.Errorf("parse member kind: %w", err)
	}

	item.Group.Source = groupSource
	item.Member.Kind = memberKind

	return item, total, nil
}

func mapMembership(row db.GetPersistedMembershipViewRow) (domain.Membership, error) {
	groupSource, err := domain.ParsePrincipalSource(string(row.GroupSource))
	if err != nil {
		return domain.Membership{}, fmt.Errorf("parse group source: %w", err)
	}

	memberKind, err := domain.ParseMemberKind(string(row.MemberKind))
	if err != nil {
		return domain.Membership{}, fmt.Errorf("parse member kind: %w", err)
	}

	return domain.Membership{
		ID: row.ID,
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

const membershipListQuery = `
WITH memberships AS (
  SELECT
    gum.id,
    g.id AS group_id,
    g.name AS group_name,
    g.source::text AS group_source,
    'user'::text AS member_kind,
    gum.user_id AS member_id,
    NULLIF(u.display_name, '') AS member_name,
    gum.created_at,
    gum.updated_at
  FROM group_user_memberships AS gum
  JOIN groups AS g
    ON g.id = gum.group_id
  LEFT JOIN users AS u
    ON u.id = gum.user_id
  WHERE ($1::uuid IS NULL OR gum.group_id = $1::uuid)
    AND ($2::uuid IS NULL OR gum.user_id = $2::uuid)
    AND $3::uuid IS NULL

  UNION ALL

  SELECT
    gmm.id,
    g.id AS group_id,
    g.name AS group_name,
    g.source::text AS group_source,
    'machine'::text AS member_kind,
    gmm.machine_id AS member_id,
    NULLIF(m.hostname, '') AS member_name,
    gmm.created_at,
    gmm.updated_at
  FROM group_machine_memberships AS gmm
  JOIN groups AS g
    ON g.id = gmm.group_id
  LEFT JOIN machines AS m
    ON m.id = gmm.machine_id
  WHERE ($1::uuid IS NULL OR gmm.group_id = $1::uuid)
    AND $2::uuid IS NULL
    AND ($3::uuid IS NULL OR gmm.machine_id = $3::uuid)
)
SELECT
  id,
  group_id,
  group_name,
  group_source,
  member_kind,
  member_id,
  member_name,
  created_at,
  updated_at,
  COUNT(*) OVER()::INT4 AS total
FROM memberships
WHERE (
  $4 = ''
  OR group_name ILIKE $4
  OR COALESCE(member_name, '') ILIKE $4
)
ORDER BY %s
LIMIT NULLIF($5::INT, 0)
OFFSET $6
`

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}

	return *id
}
