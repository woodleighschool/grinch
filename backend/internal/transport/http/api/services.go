package apihttp

import (
	"github.com/go-pkgz/auth/v2"

	"github.com/woodleighschool/grinch/internal/service/events"
	"github.com/woodleighschool/grinch/internal/service/groups"
	"github.com/woodleighschool/grinch/internal/service/machines"
	"github.com/woodleighschool/grinch/internal/service/memberships"
	"github.com/woodleighschool/grinch/internal/service/policies"
	"github.com/woodleighschool/grinch/internal/service/rules"
	"github.com/woodleighschool/grinch/internal/service/users"
)

// Services aggregates the domain services exposed by the HTTP API.
type Services struct {
	Auth        *auth.Service
	Users       *users.UserService
	Groups      *groups.GroupService
	Memberships *memberships.MembershipService
	Machines    *machines.MachineService
	Events      *events.EventService
	Rules       *rules.RuleService
	Policies    *policies.PolicyService
}
