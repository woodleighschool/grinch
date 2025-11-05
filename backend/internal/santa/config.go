package santa

import (
	"fmt"
	"net/http"
	"strings"
)

// GenerateConfigXML returns Santa configuration XML for MDM deployment.
func GenerateConfigXML(req *http.Request) string {
	// Prefer proxy-provided host when available.
	forwardedHost := req.Header.Get("X-Forwarded-Host")
	host := req.Host
	if forwardedHost != "" {
		host = forwardedHost
	}

	// Santa sync requires HTTPS, always build a https:// base.
	baseURL := fmt.Sprintf("https://%s", strings.TrimSpace(host))
	syncURL := baseURL + "/api/santa/v1"

	var config strings.Builder
	config.WriteString(fmt.Sprintf("<key>SyncBaseURL</key>\n<string>%s</string>\n", syncURL))
	config.WriteString("<key>MachineOwner</key>\n<string>{{username}}</string>\n")
	config.WriteString("<key>EnableAllEventUpload</key>\n<true/>\n")
	config.WriteString("<key>DisableUnknownEventUpload</key>\n<false/>\n")

	return config.String()
}
