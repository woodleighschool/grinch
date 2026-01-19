// Package machines exposes HTTP handlers for machine resources.
package machines

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service/machines"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts machine routes on the router.
func Register(r chi.Router, svc *machines.MachineService) {
	r.Get("/", handleList(svc))
	r.Get("/{id}", handleGet(svc))
	r.Delete("/{id}", handleDelete(svc))
}

func handleList(svc *machines.MachineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "machines", "op", "list")

		items, page, err := svc.List(r.Context(), helpers.ParseListQuery(r))
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "machines", items, page.Total)
	}
}

func handleGet(svc *machines.MachineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "machines", "op", "get")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		machine, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, machine)
	}
}

func handleDelete(svc *machines.MachineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "machines", "op", "delete")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		machine, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		err = svc.Delete(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "delete failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, machine)
	}
}
