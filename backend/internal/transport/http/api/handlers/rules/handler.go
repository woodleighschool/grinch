// Package rules provides HTTP handlers for rule resources.
package rules

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	corerules "github.com/woodleighschool/grinch/internal/core/rules"
	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service/rules"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts rule handlers on the router.
func Register(r chi.Router, svc *rules.RuleService) {
	r.Get("/", handleList(svc))
	r.Get("/{id}", handleGet(svc))
	r.Post("/", handleCreate(svc))
	r.Put("/{id}", handleUpdate(svc))
	r.Delete("/{id}", handleDelete(svc))
}

func handleList(svc *rules.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "rules", "op", "list")

		items, page, err := svc.List(r.Context(), helpers.ParseListQuery(r))
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "rules", items, page.Total)
	}
}

func handleGet(svc *rules.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "rules", "op", "get")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		rule, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, rule)
	}
}

func handleCreate(svc *rules.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "rules", "op", "create")

		payload, err := helpers.DecodeJSON[corerules.Rule](r)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}
		err = validateRulePayload(payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}

		payload.ID = uuid.Nil

		rule, err := svc.Create(r.Context(), payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "create failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusCreated, rule)
	}
}

func handleUpdate(svc *rules.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "rules", "op", "update")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		payload, err := helpers.DecodeJSON[corerules.Rule](r)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}
		err = validateRulePayload(payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}

		payload.ID = id

		rule, err := svc.Update(r.Context(), payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "update failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, rule)
	}
}

func handleDelete(svc *rules.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "rules", "op", "delete")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		rule, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		err = svc.Delete(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "delete failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, rule)
	}
}
