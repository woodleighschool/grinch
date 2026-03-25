package apihttp

import (
	appgroups "github.com/woodleighschool/grinch/internal/app/groups"
	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/store/postgres"
)

type Server struct {
	store  *postgres.Store
	groups *appgroups.Service
	rules  *apprules.Service
}

func New(
	store *postgres.Store,
	groups *appgroups.Service,
	rules *apprules.Service,
) *Server {
	return &Server{
		store:  store,
		groups: groups,
		rules:  rules,
	}
}
