package entra

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/groups"
	"github.com/woodleighschool/grinch/internal/domain/users"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
)

// Syncer syncs users, groups, and group memberships from Entra into the local store.
type Syncer struct {
	client *Client
	users  users.Service
	groups groups.Service
	log    *slog.Logger
}

// NewSyncer constructs a Syncer.
func NewSyncer(client *Client, users users.Service, groups groups.Service, log *slog.Logger) *Syncer {
	return &Syncer{
		client: client,
		users:  users,
		groups: groups,
		log:    log.With("component", "entra_sync"),
	}
}

// Sync fetches users and groups from Entra and upserts them into the local store.
func (s *Syncer) Sync(ctx context.Context) error {
	s.log.InfoContext(ctx, "entra sync started")

	fetchedUsers, err := s.client.FetchUsers(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "fetch users", "error", err)
		return err
	}

	fetchedGroups, err := s.client.FetchGroups(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "fetch groups", "error", err)
		return err
	}

	for _, u := range fetchedUsers {
		if err = s.users.Upsert(ctx, users.User{
			ID:          u.ID,
			UPN:         u.UPN,
			DisplayName: u.DisplayName,
		}); err != nil {
			s.log.WarnContext(ctx, "upsert user", "user_id", u.ID, "upn", u.UPN, "error", err)
			return err
		}
	}

	var memberIDs []uuid.UUID
	for _, g := range fetchedGroups {
		memberIDs, err = s.client.FetchGroupMembers(ctx, g.ID)
		if err != nil {
			s.log.WarnContext(ctx, "fetch group members", "group_id", g.ID, "error", err)
			return err
		}

		group := groups.Group{
			ID:          g.ID,
			DisplayName: g.DisplayName,
			Description: g.Description,
			MemberCount: pgconv.IntToInt32(len(memberIDs)),
		}

		if err = s.groups.Upsert(ctx, group); err != nil {
			s.log.WarnContext(ctx, "upsert group", "group_id", g.ID, "error", err)
			return err
		}

		if err = s.groups.ReplaceMemberships(ctx, g.ID, memberIDs); err != nil {
			s.log.WarnContext(ctx, "replace memberships", "group_id", g.ID, "error", err)
			return err
		}

		s.log.DebugContext(ctx, "group synced", "group_id", g.ID, "members", len(memberIDs))
	}

	s.log.InfoContext(ctx, "entra sync completed", "users", len(fetchedUsers), "groups", len(fetchedGroups))
	return nil
}
