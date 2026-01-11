// Package apihttp defines the React Admin HTTP API router.
package apihttp

import (
	"github.com/go-chi/chi/v5"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/events"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/groups"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/machines"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/memberships"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/policies"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/rules"
	"github.com/woodleighschool/grinch/internal/transport/http/api/handlers/users"
)

// Router builds the authenticated API routes exposed to React Admin.
func Router(svc Services) chi.Router {
	r := chi.NewRouter()

	// All resource routes require an authenticated session.
	r.Use(AuthMiddleware(svc.Auth))

	r.Route("/users", func(r chi.Router) {
		users.Register(r, svc.Users)
	})

	r.Route("/groups", func(r chi.Router) {
		groups.Register(r, svc.Groups)
	})

	r.Route("/memberships", func(r chi.Router) {
		memberships.Register(r, svc.Memberships)
	})

	r.Route("/machines", func(r chi.Router) {
		machines.Register(r, svc.Machines)
	})

	r.Route("/events", func(r chi.Router) {
		events.Register(r, svc.Events)
	})

	r.Route("/rules", func(r chi.Router) {
		rules.Register(r, svc.Rules)
	})

	r.Route("/policies", func(r chi.Router) {
		policies.Register(r, svc.Policies)
	})

	return r
}
