// Package users provides HTTP handlers for user resources.
package users

import (
	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/domain/users"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts user routes on the router.
func Register(r chi.Router, svc users.Service) {
	res := &rest.Resource[users.User, users.User, any]{
		Name: "users",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
