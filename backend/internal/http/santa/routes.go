package santa

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
)

// Dependencies bundles the shared services Santa handlers require.
type Dependencies struct {
	Store    *store.Store
	Logger   *slog.Logger
	Compiler *rules.Compiler
}

// RegisterRoutes attaches the Santa sync endpoints for agents.
func RegisterRoutes(r chi.Router, deps Dependencies) {
	preflight := &preflightHandler{store: deps.Store, logger: deps.Logger}
	ruleDownload := &ruleDownloadHandler{store: deps.Store, logger: deps.Logger, compiler: deps.Compiler}
	events := &eventUploadHandler{store: deps.Store, logger: deps.Logger}
	postflight := &postflightHandler{store: deps.Store, logger: deps.Logger}

	r.Group(func(r chi.Router) {
		r.Post("/preflight/{machineID}", preflight.Handle)
		r.Post("/eventupload/{machineID}", events.Handle)
		r.Post("/ruledownload/{machineID}", ruleDownload.Handle)
		r.Post("/postflight/{machineID}", postflight.Handle)
	})
}
