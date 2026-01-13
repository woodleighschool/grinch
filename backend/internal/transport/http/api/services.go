package apihttp

import (
	"github.com/go-pkgz/auth/v2"

	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/domain/groups"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/domain/memberships"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/domain/rules"
	"github.com/woodleighschool/grinch/internal/domain/users"
)

// Services aggregates the domain services exposed by the HTTP API.
type Services struct {
	Auth        *auth.Service
	Users       users.Service
	Groups      groups.Service
	Memberships memberships.Service
	Machines    machines.Service
	Events      events.Service
	Rules       rules.Service
	Policies    policies.Service
}
