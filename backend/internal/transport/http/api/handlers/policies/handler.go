// Package policies provides HTTP handlers for policy resources.
package policies

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	corepolicies "github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/service/policies"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts the policy resource handlers on the router.
func Register(r chi.Router, svc *policies.PolicyService) {
	res := &helpers.Resource[corepolicies.Policy, corepolicies.PolicyListItem, corepolicies.Policy]{
		Name: "policies",
		List: svc.List,
		Get:  svc.Get,
		Create: func(ctx context.Context, p corepolicies.Policy) (corepolicies.Policy, error) {
			if err := validatePolicyPayload(p); err != nil {
				return corepolicies.Policy{}, err
			}
			p.ID = uuid.Nil
			return svc.Create(ctx, p)
		},
		Update: func(ctx context.Context, id uuid.UUID, p corepolicies.Policy) (corepolicies.Policy, error) {
			if err := validatePolicyPayload(p); err != nil {
				return corepolicies.Policy{}, err
			}
			p.ID = id
			return svc.Update(ctx, p)
		},
		Delete: svc.Delete,
	}
	res.Register(r)
}
