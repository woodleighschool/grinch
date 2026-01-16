// Package machines exposes HTTP handlers for machine resources.
package machines

import (
	"github.com/go-chi/chi/v5"

	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/service/machines"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

// Register mounts machine routes on the router.
func Register(r chi.Router, svc *machines.MachineService) {
	res := &helpers.Resource[coremachines.Machine, coremachines.MachineListItem, coremachines.Machine]{
		Name: "machines",
		List: svc.List,
		Get:  svc.Get,
	}
	res.Register(r)
}
