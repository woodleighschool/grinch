// Package policies provides HTTP handlers for policy resources.
package policies

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	corepolicies "github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/platform/logging"
	"github.com/woodleighschool/grinch/internal/service/policies"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts the policy resource handlers on the router.
func Register(r chi.Router, svc *policies.PolicyService) {
	r.Get("/", handleList(svc))
	r.Get("/{id}", handleGet(svc))
	r.Post("/", handleCreate(svc))
	r.Put("/{id}", handleUpdate(svc))
	r.Delete("/{id}", handleDelete(svc))
}

func handleList(svc *policies.PolicyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "policies", "op", "list")

		items, page, err := svc.List(r.Context(), helpers.ParseListQuery(r))
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "policies", items, page.Total)
	}
}

func handleGet(svc *policies.PolicyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "policies", "op", "get")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		policy, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, policy)
	}
}

func handleCreate(svc *policies.PolicyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "policies", "op", "create")

		payload, err := helpers.DecodeJSON[corepolicies.Policy](r)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}
		err = validatePolicyPayload(payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}

		payload.ID = uuid.Nil

		policy, err := svc.Create(r.Context(), payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "create failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusCreated, policy)
	}
}

func handleUpdate(svc *policies.PolicyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "policies", "op", "update")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		payload, err := helpers.DecodeJSON[corepolicies.Policy](r)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}
		err = validatePolicyPayload(payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}

		payload.ID = id

		policy, err := svc.Update(r.Context(), payload)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "update failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, policy)
	}
}

func handleDelete(svc *policies.PolicyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "policies", "op", "delete")

		id, err := helpers.ParseUUID(r, "id")
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		policy, err := svc.Get(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		err = svc.Delete(r.Context(), id)
		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "delete failed")
			return
		}

		helpers.WriteJSON(r.Context(), w, http.StatusOK, policy)
	}
}
