// Package groups exposes HTTP handlers for group resources.
package groups

import (
	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/domain/groups"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts the group resource handlers on the router.
func Register(r chi.Router, svc groups.Service) {
	res := &rest.Resource[groups.Group, groups.Group, any]{
		Name: "groups",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
