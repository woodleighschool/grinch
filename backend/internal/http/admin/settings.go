package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (h Handler) settingsRoutes(r chi.Router) {
	r.Get("/santa-config", h.getSantaConfig)
}

type santaConfigResponse struct {
	XML string `json:"xml"`
}

func (h Handler) getSantaConfig(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(h.Config.SiteBaseURL, "/")
	if base == "" {
		respondError(w, http.StatusInternalServerError, "site URL not configured")
		return
	}

	syncURL := base + "/santa"

	var config strings.Builder
	config.WriteString(fmt.Sprintf("<key>SyncBaseURL</key>\n<string>%s</string>\n", syncURL))
	config.WriteString("<key>MachineOwner</key>\n<string>{{username}}</string>\n")
	config.WriteString("<key>SyncClientContentEncoding</key>\n<string>gzip</string>\n")

	respondJSON(w, http.StatusOK, santaConfigResponse{XML: config.String()})
}
