package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type fileDTO struct {
	SHA256       string          `json:"sha256"`
	Name         string          `json:"name"`
	SigningID    string          `json:"signingId,omitempty"`
	CDHash       string          `json:"cdhash"`
	SigningChain json.RawMessage `json:"signingChain"`
	Entitlements json.RawMessage `json:"entitlements"`
	EventCount   int64           `json:"eventCount"`
	FirstSeen    time.Time       `json:"firstSeen"`
	LastSeen     time.Time       `json:"lastSeen"`
}

func (h Handler) filesRoutes(r chi.Router) {
	r.Get("/", h.listFiles)
}

func (h Handler) listFiles(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 100)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	files, err := h.Store.ListFiles(r.Context(), int32(limit), int32(offset))
	if err != nil {
		h.Logger.Error("list files", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list files")
		return
	}
	respondJSON(w, http.StatusOK, mapFiles(files))
}

func mapFiles(files []sqlc.ListFilesRow) []fileDTO {
	resp := make([]fileDTO, 0, len(files))
	for _, f := range files {
		resp = append(resp, mapFile(f))
	}
	return resp
}

func mapFile(f sqlc.ListFilesRow) fileDTO {
	var signingChain json.RawMessage
	if len(f.SigningChain) > 0 {
		signingChain = cloneJSONMessage(f.SigningChain)
	}
	var entitlements json.RawMessage
	if len(f.Entitlements) > 0 {
		entitlements = cloneJSONMessage(f.Entitlements)
	}
	return fileDTO{
		SHA256:       f.Sha256,
		Name:         f.Name,
		SigningID:    strings.TrimSpace(f.SigningID.String),
		CDHash:       f.Cdhash.String,
		SigningChain: signingChain,
		Entitlements: entitlements,
		EventCount:   f.EventCount,
		FirstSeen:    f.CreatedAt.Time,
		LastSeen:     f.UpdatedAt.Time,
	}
}

func cloneJSONMessage(data []byte) json.RawMessage {
	if len(data) == 0 {
		return nil
	}
	dup := make([]byte, len(data))
	copy(dup, data)
	return json.RawMessage(dup)
}
