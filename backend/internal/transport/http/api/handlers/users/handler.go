// Package users provides HTTP handlers for user resources.
package users

import (
	"github.com/go-chi/chi/v5"

	coreusers "github.com/woodleighschool/grinch/internal/core/users"
	"github.com/woodleighschool/grinch/internal/service/users"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts user routes on the router.
func Register(r chi.Router, svc *users.UserService) {
	res := &helpers.Resource[coreusers.User, coreusers.User, coreusers.User]{
		Name: "users",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
