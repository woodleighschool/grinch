// Package rules provides HTTP handlers for rule resources.
package rules

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	corerules "github.com/woodleighschool/grinch/internal/core/rules"
	"github.com/woodleighschool/grinch/internal/service/rules"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts rule handlers on the router.
func Register(r chi.Router, svc *rules.RuleService) {
	res := &helpers.Resource[corerules.Rule, corerules.Rule, corerules.Rule]{
		Name: "rules",
		List: svc.List,
		Get:  svc.Get,
		Create: func(ctx context.Context, p corerules.Rule) (corerules.Rule, error) {
			if err := validateRulePayload(p); err != nil {
				return corerules.Rule{}, err
			}
			p.ID = uuid.Nil
			return svc.Create(ctx, p)
		},
		Update: func(ctx context.Context, id uuid.UUID, p corerules.Rule) (corerules.Rule, error) {
			if err := validateRulePayload(p); err != nil {
				return corerules.Rule{}, err
			}
			p.ID = id
			return svc.Update(ctx, p)
		},
		Delete: svc.Delete,
	}
	res.Register(r)
}
