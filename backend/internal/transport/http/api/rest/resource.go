package rest

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/logging"
)

// Resource registers REST endpoints for a single resource.
//
// Handlers are mounted only for non nil operations.
type Resource[T, L, P any] struct {
	Name   string
	List   func(ctx context.Context, query listing.Query) ([]L, listing.Page, error)
	Get    func(ctx context.Context, id uuid.UUID) (T, error)
	Create func(ctx context.Context, payload P) (T, error)
	Update func(ctx context.Context, id uuid.UUID, payload P) (T, error)
	Delete func(ctx context.Context, id uuid.UUID) error
}

// Register mounts the resource handlers on the router.
func (res *Resource[T, L, P]) Register(r chi.Router) {
	if res.List != nil {
		r.Get("/", res.handleList())
	}
	if res.Get != nil {
		r.Get("/{id}", res.handleGet())
	}
	if res.Create != nil {
		r.Post("/", res.handleCreate())
	}
	if res.Update != nil {
		r.Put("/{id}", res.handleUpdate())
	}
	if res.Delete != nil && res.Get != nil {
		r.Delete("/{id}", res.handleDelete())
	}
}

func (res *Resource[T, L, P]) handleList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", res.Name, "op", "list")

		items, page, err := res.List(r.Context(), ParseListQuery(r))
		if err != nil {
			WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		WriteList(w, r, res.Name, items, page.Total)
	}
}

func (res *Resource[T, L, P]) handleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", res.Name, "op", "get")

		id, err := ParseUUID(r, "id")
		if err != nil {
			WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		item, err := res.Get(r.Context(), id)
		if err != nil {
			WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		WriteJSON(r.Context(), w, http.StatusOK, item)
	}
}

func (res *Resource[T, L, P]) handleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", res.Name, "op", "create")

		payload, err := DecodeJSON[P](r)
		if err != nil {
			WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}

		item, err := res.Create(r.Context(), payload)
		if err != nil {
			WriteError(r.Context(), w, log, err, "create failed")
			return
		}

		WriteJSON(r.Context(), w, http.StatusCreated, item)
	}
}

func (res *Resource[T, L, P]) handleUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", res.Name, "op", "update")

		id, err := ParseUUID(r, "id")
		if err != nil {
			WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		payload, err := DecodeJSON[P](r)
		if err != nil {
			WriteError(r.Context(), w, log, err, "invalid payload")
			return
		}

		item, err := res.Update(r.Context(), id, payload)
		if err != nil {
			WriteError(r.Context(), w, log, err, "update failed")
			return
		}

		WriteJSON(r.Context(), w, http.StatusOK, item)
	}
}

func (res *Resource[T, L, P]) handleDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", res.Name, "op", "delete")

		id, err := ParseUUID(r, "id")
		if err != nil {
			WriteError(r.Context(), w, log, err, "invalid id")
			return
		}

		item, err := res.Get(r.Context(), id)
		if err != nil {
			WriteError(r.Context(), w, log, err, "get failed")
			return
		}

		// The response includes the deleted record for simple rest compatibility.
		if err = res.Delete(r.Context(), id); err != nil {
			WriteError(r.Context(), w, log, err, "delete failed")
			return
		}

		WriteJSON(r.Context(), w, http.StatusOK, item)
	}
}
