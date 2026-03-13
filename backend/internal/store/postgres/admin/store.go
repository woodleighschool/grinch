package admin

import (
	"github.com/woodleighschool/grinch/internal/store/db"
	"github.com/woodleighschool/grinch/internal/store/postgres"
)

type Store struct {
	store   *postgres.Store
	queries *db.Queries
}

const (
	groupMutationStatusDeleted  = "deleted"
	groupMutationStatusNotFound = "not_found"
	groupMutationStatusOK       = "ok"
	groupMutationStatusReadOnly = "read_only"
)

func New(store *postgres.Store) *Store {
	return &Store{
		store:   store,
		queries: store.Queries(),
	}
}
