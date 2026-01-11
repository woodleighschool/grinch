// Package policies provides HTTP handlers for policy resources.
package policies

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/transport/http/api/rest"
)

// Register mounts the policy resource handlers on the router.
func Register(r chi.Router, svc policies.Service) {
	res := &rest.Resource[policies.Policy, policies.ListItem, policies.Policy]{
		Name: "policies",
		List: svc.List,
		Get:  svc.Get,
		Create: func(ctx context.Context, p policies.Policy) (policies.Policy, error) {
			p.ID = uuid.Nil
			return svc.Create(ctx, p)
		},
		Update: func(ctx context.Context, id uuid.UUID, p policies.Policy) (policies.Policy, error) {
			p.ID = id
			return svc.Update(ctx, p)
		},
		Delete: svc.Delete,
	}
	res.Register(r)
}
