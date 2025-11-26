package admin

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

// settingsRoutes exposes helper endpoints for config files.
func (h Handler) settingsRoutes(r chi.Router) {
	r.Get("/santa-config", h.getSantaConfig)
}

// santaConfigResponse wraps the generated XML snippet.
type santaConfigResponse struct {
	XML string `json:"xml"`
}

// getSantaConfig renders the XML plist snippet the MDM profile expects.
func (h Handler) getSantaConfig(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(h.Config.SiteBaseURL, "/")
	if base == "" {
		respondError(w, http.StatusInternalServerError, "site URL not configured")
		return
	}

	syncURL := base + "/santa"
	if u, err := url.Parse(base); err == nil {
		if _, port, err := net.SplitHostPort(h.Config.SantaListenAddr); err == nil {
			u.Host = net.JoinHostPort(u.Hostname(), port)
			syncURL = u.String() + "/santa"
		}
	}

	var config strings.Builder
	config.WriteString(fmt.Sprintf("<key>SyncBaseURL</key>\n<string>%s</string>\n", syncURL))
	config.WriteString("<key>MachineOwner</key>\n<string>{{username}}</string>\n")
	config.WriteString("<key>SyncClientContentEncoding</key>\n<string>gzip</string>\n")

	respondJSON(w, http.StatusOK, santaConfigResponse{XML: config.String()})
}
