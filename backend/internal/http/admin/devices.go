package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type deviceDTO struct {
	ID                   string          `json:"id"`
	MachineIdentifier    string          `json:"machineIdentifier"`
	Serial               string          `json:"serial"`
	Hostname             string          `json:"hostname"`
	PrimaryUser          string          `json:"primaryUser,omitempty"`
	UserID               uuid.UUID       `json:"userId,omitempty"`
	ClientMode           string          `json:"clientMode"`
	CleanSyncRequested   bool            `json:"cleanSyncRequested"`
	LastSeen             time.Time       `json:"lastSeen"`
	LastPreflightAt      time.Time       `json:"lastPreflightAt,omitempty"`
	LastPostflightAt     time.Time       `json:"lastPostflightAt,omitempty"`
	LastRulesReceived    *int            `json:"lastRulesReceived,omitempty"`
	LastRulesProcessed   *int            `json:"lastRulesProcessed,omitempty"`
	LastPreflightPayload json.RawMessage `json:"lastPreflightPayload,omitempty"`
	RuleCursor           string          `json:"ruleCursor"`
	SyncCursor           string          `json:"syncCursor"`
}

type deviceDetailsResponse struct {
	Device       deviceDTO    `json:"device"`
	PrimaryUser  userDTO      `json:"primary_user"`
	RecentBlocks []eventDTO   `json:"recent_blocks"`
	Policies     []userPolicy `json:"policies"`
}

func (h Handler) devicesRoutes(r chi.Router) {
	r.Get("/", h.listDevices)
	r.Get("/{id}", h.deviceDetails)
}

func (h Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	machines, err := h.Store.ListMachines(r.Context(), int32(limit), int32(offset), search)
	if err != nil {
		h.Logger.Error("list devices", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}
	resp := make([]deviceDTO, 0, len(machines))
	for _, m := range machines {
		resp = append(resp, mapDevice(m))
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h Handler) deviceDetails(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	machineID, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx := r.Context()
	machine, err := h.Store.GetMachine(ctx, machineID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "device not found")
			return
		}
		h.Logger.Error("get device", "err", err, "device", machineID)
		respondError(w, http.StatusInternalServerError, "failed to load device")
		return
	}

	var (
		primaryUser sqlc.User
		events      []sqlc.ListBlocksByUserRow
		assignments []sqlc.ListUserAssignmentsRow
	)
	if machine.UserID.Valid {
		userID, err := uuid.FromBytes(machine.UserID.Bytes[:])
		if err != nil {
			h.Logger.Error("parse device primary user id", "err", err, "device", machineID)
			respondError(w, http.StatusInternalServerError, "failed to parse user id")
			return
		}
		primaryUser, err = h.Store.GetUser(ctx, userID)
		if err != nil {
			h.Logger.Error("get device primary user", "err", err, "device", machine, "user", machine.PrimaryUser)
			respondError(w, http.StatusInternalServerError, "failed to load device primary user")
			return
		}
		events, err = h.Store.ListBlocksByUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			h.Logger.Error("get device primary user block events", "err", err, "device", machineID, "user", machine.PrimaryUser)
			respondError(w, http.StatusInternalServerError, "failed to load device primary user block events")
			return
		}
		assignments, err = h.Store.ListUserAssignments(ctx, userID)
		if err != nil {
			h.Logger.Error("list device primary user policies", "err", err, "device", machineID, "user", machine.UserID)
			respondError(w, http.StatusInternalServerError, "failed to load device primary user policies")
			return
		}
	} else {
		h.Logger.Warn("device missing primary user", "device", machineID)
	}

	resp := deviceDetailsResponse{
		Device:       mapDevice(machine),
		PrimaryUser:  mapUserDTO(primaryUser),
		RecentBlocks: mapUserBlocks(events),
		Policies:     mapUserPolicies(assignments, primaryUser),
	}
	respondJSON(w, http.StatusOK, resp)
}

func mapDevice(m sqlc.Machine) deviceDTO {
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
	return deviceDTO{
		ID:                   m.ID.String(),
		MachineIdentifier:    m.MachineIdentifier,
		Serial:               m.Serial,
		Hostname:             m.Hostname,
		PrimaryUser:          m.PrimaryUser.String,
		ClientMode:           strings.ToUpper(m.ClientMode),
		CleanSyncRequested:   m.CleanSyncRequested,
		LastSeen:             lastSeen,
		LastPreflightAt:      lastPreflight,
		LastPostflightAt:     lastPostflight,
		LastRulesReceived:    intPtrFromInt4(m.LastRulesReceived),
		LastRulesProcessed:   intPtrFromInt4(m.LastRulesProcessed),
		LastPreflightPayload: m.LastPreflightPayload,
		RuleCursor:           m.RuleCursor.String,
		SyncCursor:           m.SyncCursor.String,
	}
}

func intPtrFromInt4(val pgtype.Int4) *int {
	if !val.Valid {
		return nil
	}
	v := int(val.Int32)
	return &v
}
