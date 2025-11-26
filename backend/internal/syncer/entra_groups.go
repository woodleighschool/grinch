package syncer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/graph"
	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// NewGroupJob synchronises Entra ID groups and their memberships.
func NewGroupJob(store *store.Store, graphClient *graph.Client, logger *slog.Logger) Job {
	return func(ctx context.Context) error {
		if graphClient == nil || !graphClient.Enabled() {
			return graph.ErrNotConfigured
		}
		groups, err := graphClient.FetchGroups(ctx)
		if err != nil {
			return fmt.Errorf("fetch groups: %w", err)
		}
		for _, g := range groups {
			id := uuid.New()
			if g.ObjectID != "" {
				if parsed, err := uuid.Parse(g.ObjectID); err == nil {
					id = parsed
				}
			}
			row, err := store.UpsertGroup(ctx, sqlc.UpsertGroupParams{
				ID:          id,
				DisplayName: g.DisplayName,
				Description: pgtype.Text{String: g.Description, Valid: g.Description != ""},
			})
			if err != nil {
				logger.Error("upsert group", "group", g.DisplayName, "err", err)
				continue
			}
			var members []uuid.UUID
			for _, memberID := range g.Members {
				if parsed, err := uuid.Parse(memberID); err == nil {
					members = append(members, parsed)
				}
			}
			if len(members) > 0 {
				if err := store.ReplaceGroupMembers(ctx, row.ID, members); err != nil {
					logger.Error("replace group members", "group", row.DisplayName, "err", err)
				}
			}
		}
		return nil
	}
}
