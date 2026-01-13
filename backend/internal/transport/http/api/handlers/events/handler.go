// Package events provides HTTP handlers for event resources.
package events

import (
	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts event resource handlers on the router.
func Register(r chi.Router, svc events.Service) {
	res := &rest.Resource[events.Event, events.ListItem, any]{
		Name: "events",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
