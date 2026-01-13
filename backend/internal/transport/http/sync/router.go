package sync

import (
	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/domain/santa"
)

// Router returns the HTTP router implementing the Santa sync protocol.
func Router(svc santa.SyncService) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(svc)

	// Santa sync protocol endpoints.
	// TODO: add some sort of authentication middleware that works through Cloudflare.
	r.Post("/preflight/{machine_id}", h.Preflight)
	r.Post("/eventupload/{machine_id}", h.EventUpload)
	r.Post("/ruledownload/{machine_id}", h.RuleDownload)
	r.Post("/postflight/{machine_id}", h.Postflight)

	return r
}
