package santa

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/store"
)

// resolveUserID attempts to match Santa's reported username to a directory user.
func resolveUserID(ctx context.Context, store *store.Store, logger *slog.Logger, reported string) *uuid.UUID {
	reported = strings.TrimSpace(reported)
	if reported == "" {
		return nil
	}
	lower := strings.ToLower(reported)

	// TODO: This requires machines are configured in such the local user's username is the prefix of the UPN.
	// This may not always be the case. Maybe find a better way to do this?
	login := lower
	if idx := strings.IndexRune(login, '@'); idx > 0 {
		login = login[:idx]
	}
	if login == "" {
		return nil
	}
	user, err := store.GetUserByLogin(ctx, login)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logger.Warn("lookup user by login", "login", login, "err", err)
		}
		return nil
	}
	return &user.ID
}
