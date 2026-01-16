// Package groups exposes HTTP handlers for group resources.
package groups

import (
	"github.com/go-chi/chi/v5"

	coregroups "github.com/woodleighschool/grinch/internal/core/groups"
	"github.com/woodleighschool/grinch/internal/service/groups"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts the group resource handlers on the router.
func Register(r chi.Router, svc *groups.GroupService) {
	res := &helpers.Resource[coregroups.Group, coregroups.Group, coregroups.Group]{
		Name: "groups",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
