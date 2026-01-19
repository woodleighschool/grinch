// Package events provides HTTP handlers for event resources.
package events

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service/events"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts event resource handlers on the router.
func Register(r chi.Router, svc *events.EventService) {
	r.Get("/", handleList(svc))
	r.Get("/{id}", handleGet(svc))
	r.Delete("/{id}", handleDelete(svc))
}

func handleList(svc *events.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "events", "op", "list")

		items, page, err := svc.List(r.Context(), helpers.ParseListQuery(r))
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "events", items, page.Total)
	}
}

func handleGet(svc *events.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "events", "op", "get")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		event, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, event)
	}
}

func handleDelete(svc *events.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "events", "op", "delete")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		ev, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		err = svc.Delete(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "delete failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, ev)
	}
}
