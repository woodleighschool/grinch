package admin

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type machineDTO struct {
	ID                 string    `json:"id"`
	MachineIdentifier  string    `json:"machineIdentifier"`
	Serial             string    `json:"serial"`
	Hostname           string    `json:"hostname"`
	PrimaryUser        string    `json:"primaryUser,omitempty"`
	ClientMode         string    `json:"clientMode"`
	CleanSyncRequested bool      `json:"cleanSyncRequested"`
	LastSeen           time.Time `json:"lastSeen"`
	LastPreflightAt    time.Time `json:"lastPreflightAt,omitempty"`
	LastPostflightAt   time.Time `json:"lastPostflightAt,omitempty"`
	LastRulesReceived  *int      `json:"lastRulesReceived,omitempty"`
	LastRulesProcessed *int      `json:"lastRulesProcessed,omitempty"`
	RuleCursor         string    `json:"ruleCursor"`
	SyncCursor         string    `json:"syncCursor"`
}

func (h Handler) machinesRoutes(r chi.Router) {
	r.Get("/", h.listMachines)
}

func (h Handler) listMachines(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	machines, err := h.Store.ListMachines(r.Context(), int32(limit), int32(offset), search)
	if err != nil {
		h.Logger.Error("list machines", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list machines")
		return
	}
	resp := make([]machineDTO, 0, len(machines))
	for _, m := range machines {
		resp = append(resp, mapMachine(m))
	}
	respondJSON(w, http.StatusOK, resp)
}

func mapMachine(m sqlc.Machine) machineDTO {
	var lastSeen time.Time
	if m.LastSeen.Valid {
		lastSeen = m.LastSeen.Time
	}
	var lastPreflight time.Time
	if m.LastPreflightAt.Valid {
		lastPreflight = m.LastPreflightAt.Time
	}
	var lastPostflight time.Time
	if m.LastPostflightAt.Valid {
		lastPostflight = m.LastPostflightAt.Time
	}
	return machineDTO{
		ID:                 m.ID.String(),
		MachineIdentifier:  m.MachineIdentifier,
		Serial:             m.Serial,
		Hostname:           m.Hostname,
		PrimaryUser:        m.PrimaryUser.String,
		ClientMode:         strings.ToUpper(m.ClientMode),
		CleanSyncRequested: m.CleanSyncRequested,
		LastSeen:           lastSeen,
		LastPreflightAt:    lastPreflight,
		LastPostflightAt:   lastPostflight,
		LastRulesReceived:  intPtrFromInt4(m.LastRulesReceived),
		LastRulesProcessed: intPtrFromInt4(m.LastRulesProcessed),
		RuleCursor:         m.RuleCursor.String,
		SyncCursor:         m.SyncCursor.String,
	}
}

func intPtrFromInt4(val pgtype.Int4) *int {
	if !val.Valid {
		return nil
	}
	v := int(val.Int32)
	return &v
}
