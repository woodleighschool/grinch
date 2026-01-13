// Package machines exposes HTTP handlers for machine resources.
package machines

import (
	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts machine routes on the router.
func Register(r chi.Router, svc machines.Service) {
	res := &rest.Resource[machines.Machine, machines.ListItem, any]{
		Name: "machines",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
