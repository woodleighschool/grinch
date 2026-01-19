// Package users provides HTTP handlers for user resources.
package users

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service/users"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts user routes on the router.
func Register(r chi.Router, svc *users.UserService) {
	r.Get("/", handleList(svc))
	r.Get("/{id}", handleGet(svc))
}

func handleList(svc *users.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "users", "op", "list")

		items, page, err := svc.List(r.Context(), helpers.ParseListQuery(r))
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "users", items, page.Total)
	}
}

func handleGet(svc *users.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "users", "op", "get")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		user, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, user)
	}
}
