// Package memberships provides HTTP handlers for membership resources.
package memberships

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/memberships"
	"github.com/woodleighschool/grinch/internal/logging"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts membership handlers on the router.
func Register(r chi.Router, svc memberships.Service) {
	r.Get("/", handleList(svc))
}

func handleList(svc memberships.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "memberships", "op", "list")

		q := r.URL.Query()
		start, limit := rest.ParseRange(q.Get("range"))
		filter := parseFilter(q.Get("filter"))

		userID := rest.ParseUUIDValue(filter["user_id"])
		groupID := rest.ParseUUIDValue(filter["group_id"])

		var (
			items []memberships.Membership
			total int64
			err   error
		)

		switch {
		case userID != uuid.Nil:
			items, total, err = svc.ListByUser(r.Context(), userID, limit, start)
		case groupID != uuid.Nil:
			items, total, err = svc.ListByGroup(r.Context(), groupID, limit, start)
		default:
			rest.WriteError(
				r.Context(),
				w,
				log,
				errx.NotFound("user_id or group_id required in filter"),
				"invalid filter",
			)
			return
		}

		if err != nil {
			rest.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		rest.WriteList(w, r, "memberships", items, total)
	}
}

func parseFilter(raw string) map[string]any {
	if raw == "" {
		return nil
	}

	var m map[string]any
	_ = json.Unmarshal([]byte(raw), &m)
	return m
}
