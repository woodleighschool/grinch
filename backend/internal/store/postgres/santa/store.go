package santa

import "github.com/woodleighschool/grinch/internal/store/postgres"

type Store struct {
	store *postgres.Store
}

func New(store *postgres.Store) *Store {
	return &Store{store: store}
}
