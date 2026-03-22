package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	graphsync "github.com/woodleighschool/go-entrasync"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

// ReconcileSnapshot upserts current Entra objects and converts missing Entra objects to local.
func (store *Store) ReconcileSnapshot(
	ctx context.Context,
	snapshot *graphsync.Snapshot,
) (domain.EntrasyncResult, error) {
	if snapshot == nil {
		snapshot = &graphsync.Snapshot{
			Users:   []graphsync.User{},
			Groups:  []graphsync.Group{},
			Members: map[uuid.UUID][]uuid.UUID{},
		}
	}

	userIDs := collectUserIDs(snapshot.Users)
	groupIDs := collectGroupIDs(snapshot.Groups)

	err := store.RunInTx(ctx, func(queries *db.Queries) error {
		if upsertErr := upsertEntraUsers(ctx, queries, snapshot.Users); upsertErr != nil {
			return upsertErr
		}

		if upsertErr := upsertEntraGroups(ctx, queries, snapshot.Groups); upsertErr != nil {
			return upsertErr
		}

		if convertErr := queries.ConvertMissingEntraUsersToLocal(ctx, userIDs); convertErr != nil {
			return fmt.Errorf("convert missing users to local: %w", convertErr)
		}

		if convertErr := queries.ConvertMissingEntraGroupsToLocal(ctx, groupIDs); convertErr != nil {
			return fmt.Errorf("convert missing groups to local: %w", convertErr)
		}

		if deleteErr := queries.DeleteUserMembersForEntraGroups(ctx); deleteErr != nil {
			return fmt.Errorf("delete entra user memberships: %w", deleteErr)
		}

		return upsertMemberships(ctx, queries, snapshot.Members)
	})
	if err != nil {
		return domain.EntrasyncResult{}, err
	}

	membershipCount := 0
	for _, memberIDs := range snapshot.Members {
		membershipCount += len(memberIDs)
	}

	return domain.EntrasyncResult{
		Users:       len(snapshot.Users),
		Groups:      len(snapshot.Groups),
		Memberships: membershipCount,
	}, nil
}

func upsertEntraUsers(ctx context.Context, queries *db.Queries, users []graphsync.User) error {
	for _, user := range users {
		_, err := queries.UpsertUser(ctx, db.UpsertUserParams{
			ID:          user.ID,
			Upn:         user.UPN,
			DisplayName: user.DisplayName,
			Source:      string(domain.PrincipalSourceEntra),
		})
		if err != nil {
			return fmt.Errorf("upsert user %q: %w", user.ID, err)
		}
	}

	return nil
}

func upsertEntraGroups(ctx context.Context, queries *db.Queries, groups []graphsync.Group) error {
	for _, group := range groups {
		_, err := queries.UpsertGroup(ctx, db.UpsertGroupParams{
			ID:          group.ID,
			Name:        group.DisplayName,
			Description: group.Description,
			Source:      string(domain.PrincipalSourceEntra),
		})
		if err != nil {
			return fmt.Errorf("upsert group %q: %w", group.ID, err)
		}
	}

	return nil
}

func upsertMemberships(ctx context.Context, queries *db.Queries, membersByGroup map[uuid.UUID][]uuid.UUID) error {
	for groupID, memberIDs := range membersByGroup {
		for _, memberID := range memberIDs {
			membershipID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("create synced membership id: %w", err)
			}

			err = queries.AddSyncedMembership(ctx, db.AddSyncedMembershipParams{
				ID:         membershipID,
				GroupID:    groupID,
				MemberKind: string(domain.MemberKindUser),
				MemberID:   memberID,
			})
			if err != nil {
				return fmt.Errorf("add member %q to group %q: %w", memberID, groupID, err)
			}
		}
	}

	return nil
}

func collectUserIDs(users []graphsync.User) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func collectGroupIDs(groups []graphsync.Group) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.ID)
	}
	return ids
}
