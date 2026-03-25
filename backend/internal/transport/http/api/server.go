package apihttp

import (
	appfileaccessevents "github.com/woodleighschool/grinch/internal/app/fileaccessevents"
	appgroups "github.com/woodleighschool/grinch/internal/app/groups"
	appmemberships "github.com/woodleighschool/grinch/internal/app/memberships"
	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/store/postgres"
)

type Server struct {
	store            *postgres.Store
	groups           *appgroups.Service
	fileAccessEvents *appfileaccessevents.Service
	memberships      *appmemberships.Service
	rules            *apprules.Service
}

func New(
	store *postgres.Store,
	groups *appgroups.Service,
	fileAccessEvents *appfileaccessevents.Service,
	rules *apprules.Service,
	memberships *appmemberships.Service,
) *Server {
	return &Server{
		store:            store,
		groups:           groups,
		fileAccessEvents: fileAccessEvents,
		memberships:      memberships,
		rules:            rules,
	}
}
