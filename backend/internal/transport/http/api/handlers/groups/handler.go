// Package groups exposes HTTP handlers for group resources.
package groups

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service/groups"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts the group resource handlers on the router.
func Register(r chi.Router, svc *groups.GroupService) {
	r.Get("/", handleList(svc))
	r.Get("/{id}", handleGet(svc))
}

func handleList(svc *groups.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "groups", "op", "list")

		items, page, err := svc.List(r.Context(), helpers.ParseListQuery(r))
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "groups", items, page.Total)
	}
}

func handleGet(svc *groups.GroupService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "groups", "op", "get")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		group, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, group)
	}
}
