// Package memberships provides HTTP handlers for membership resources.
package memberships

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	corememberships "github.com/woodleighschool/grinch/internal/core/memberships"
	"github.com/woodleighschool/grinch/internal/logging"
	"github.com/woodleighschool/grinch/internal/service/memberships"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts membership handlers on the router.
func Register(r chi.Router, svc *memberships.MembershipService) {
	r.Get("/", handleList(svc))
}

func handleList(svc *memberships.MembershipService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context()).With("resource", "memberships", "op", "list")

		q := r.URL.Query()
		start, limit := helpers.ParseRange(q.Get("range"))
		filter := helpers.ParseFilterMap(q.Get("filter"))

		userID := helpers.ParseUUIDValue(filter["user_id"])
		groupID := helpers.ParseUUIDValue(filter["group_id"])

		var (
			items []corememberships.Membership
			total int64
			err   error
		)

		switch {
		case userID != uuid.Nil:
			items, total, err = svc.ListByUser(r.Context(), userID, limit, start)
		case groupID != uuid.Nil:
			items, total, err = svc.ListByGroup(r.Context(), groupID, limit, start)
		default:
			helpers.WriteError(
				r.Context(),
				w,
				log,
				coreerrors.NotFound("user_id or group_id required in filter"),
				"invalid filter",
			)
			return
		}

		if err != nil {
			helpers.WriteError(r.Context(), w, log, err, "list failed")
			return
		}

		helpers.WriteList(w, r, "memberships", items, total)
	}
}
