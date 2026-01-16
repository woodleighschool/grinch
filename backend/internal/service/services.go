package service

import (
	"github.com/woodleighschool/grinch/internal/service/events"
	"github.com/woodleighschool/grinch/internal/service/groups"
	"github.com/woodleighschool/grinch/internal/service/machines"
	"github.com/woodleighschool/grinch/internal/service/memberships"
	"github.com/woodleighschool/grinch/internal/service/policies"
	"github.com/woodleighschool/grinch/internal/service/rules"
	syncex "github.com/woodleighschool/grinch/internal/service/sync"
	"github.com/woodleighschool/grinch/internal/service/users"
)

// Services bundles the use-case layer for consumption by transports.
type Services struct {
	Users       *users.UserService
	Groups      *groups.GroupService
	Memberships *memberships.MembershipService
	Machines    *machines.MachineService
	Events      *events.EventService
	Rules       *rules.RuleService
	Policies    *policies.PolicyService
	Sync        *syncex.Service
}

// Stores provides the persistence interfaces required to build Services.
type Stores struct {
	Users       users.UserStore
	Groups      groups.GroupStore
	Memberships memberships.MembershipStore
	Machines    machines.MachineStore
	Events      events.EventStore
	Rules       rules.RuleStore
	Policies    policies.PolicyStore
}

// NewServices wires repositories into use-case services and connects cross-cutting dependencies.
func NewServices(stores Stores) Services {
	userSvc := users.NewUserService(stores.Users)
	membershipSvc := memberships.NewMembershipService(stores.Memberships)
	machineSvc := machines.NewMachineService(stores.Machines, userSvc)
	policySvc := policies.NewPolicyService(stores.Policies, membershipSvc, machineSvc)
	groupSvc := groups.NewGroupService(stores.Groups, policySvc)
	ruleSvc := rules.NewRuleService(stores.Rules, policySvc)
	eventSvc := events.NewEventService(stores.Events)
	syncSvc := syncex.NewService(machineSvc, policySvc, ruleSvc, eventSvc)

	return Services{
		Users:       userSvc,
		Groups:      groupSvc,
		Memberships: membershipSvc,
		Machines:    machineSvc,
		Events:      eventSvc,
		Rules:       ruleSvc,
		Policies:    policySvc,
		Sync:        syncSvc,
	}
}
