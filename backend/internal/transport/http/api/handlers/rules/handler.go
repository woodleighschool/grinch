// Package rules provides HTTP handlers for rule resources.
package rules

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/rules"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts rule handlers on the router.
func Register(r chi.Router, svc rules.Service) {
	res := &rest.Resource[rules.Rule, rules.Rule, rules.Rule]{
		Name: "rules",
		List: svc.List,
		Get:  svc.Get,
		Create: func(ctx context.Context, p rules.Rule) (rules.Rule, error) {
			p.ID = uuid.Nil
			return svc.Create(ctx, p)
		},
		Update: func(ctx context.Context, id uuid.UUID, p rules.Rule) (rules.Rule, error) {
			p.ID = id
			return svc.Update(ctx, p)
		},
		Delete: svc.Delete,
	}
	res.Register(r)
}
