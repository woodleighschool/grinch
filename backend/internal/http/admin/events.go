package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type eventDTO struct {
	ID        int64           `json:"id"`
	Occurred  time.Time       `json:"occurredAt"`
	Kind      string          `json:"kind"`
	Payload   json.RawMessage `json:"payload"`
	Hostname  string          `json:"hostname"`
	MachineID string          `json:"machineId"`
	Email     string          `json:"email"`
	UserID    string          `json:"userId"`
}

func (h Handler) eventsRoutes(r chi.Router) {
	r.Get("/", h.listEvents)
	r.Get("/stats", h.eventStats)
}

func (h Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 100)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	events, err := h.Store.ListEvents(r.Context(), int32(limit), int32(offset))
	if err != nil {
		h.Logger.Error("list events", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	resp := make([]eventDTO, 0, len(events))
	for _, e := range events {
		resp = append(resp, mapEvent(e))
	}
	respondJSON(w, http.StatusOK, resp)
}

type eventStatDTO struct {
	Bucket time.Time `json:"bucket"`
	Kind   string    `json:"kind"`
	Total  int64     `json:"total"`
}

func (h Handler) eventStats(w http.ResponseWriter, r *http.Request) {
	days := parseInt(r.URL.Query().Get("days"), 14)
	stats, err := h.Store.SummariseEvents(r.Context(), int32(days))
	if err != nil {
		h.Logger.Error("summarise events", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to summarise events")
		return
	}
	resp := make([]eventStatDTO, 0, len(stats))
	for _, row := range stats {
		resp = append(resp, eventStatDTO{
			Bucket: row.Bucket.Time,
			Kind:   row.Kind,
			Total:  row.Total,
		})
	}
	respondJSON(w, http.StatusOK, resp)
}

func mapEvent(e sqlc.ListEventSummariesRow) eventDTO {
	var email string
	if e.Upn.Valid {
		email = e.Upn.String
	}
	var occurred time.Time
	if e.OccurredAt.Valid {
		occurred = e.OccurredAt.Time
	}
	var userId string
	if e.Userid.Valid {
		userId = e.Userid.String()
	}

	return eventDTO{
		ID:        e.ID,
		Kind:      e.Kind,
		Payload:   e.Payload,
		Occurred:  occurred,
		Hostname:  e.Hostname,
		MachineID: e.Machineid.String(),
		Email:     email,
		UserID:    userId,
	}
}

func mapUserBlock(e sqlc.ListBlocksByUserRow) eventDTO {
	var occurred time.Time
	if e.OccurredAt.Valid {
		occurred = e.OccurredAt.Time
	}

	return eventDTO{
		ID:        e.ID,
		Kind:      e.Kind,
		Payload:   e.Payload,
		Occurred:  occurred,
		Hostname:  e.Hostname,
		MachineID: e.Machineid.String(),
		Email:     e.Upn,
		UserID:    e.Userid.String(),
	}
}
