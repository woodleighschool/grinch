// Package events provides HTTP handlers for event resources.
package events

import (
	"github.com/go-chi/chi/v5"

	coreevents "github.com/woodleighschool/grinch/internal/core/events"
	"github.com/woodleighschool/grinch/internal/service/events"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts event resource handlers on the router.
func Register(r chi.Router, svc *events.EventService) {
	res := &helpers.Resource[coreevents.Event, coreevents.EventListItem, coreevents.Event]{
		Name: "events",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
