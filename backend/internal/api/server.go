package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/backend/internal/auth"
	"github.com/woodleighschool/grinch/backend/internal/config"
	"github.com/woodleighschool/grinch/backend/internal/events"
	"github.com/woodleighschool/grinch/backend/internal/models"
	"github.com/woodleighschool/grinch/backend/internal/santa"
	"github.com/woodleighschool/grinch/backend/internal/store"
)

type Server struct {
	cfg         *config.Config
	store       *store.Store
	santa       *santa.Service
	sessions    *auth.SessionManager
	broadcaster *events.Broadcaster
	saml        *auth.SAMLService
	samlMu      sync.RWMutex
	logger      *slog.Logger
}

func (s *Server) samlService() *auth.SAMLService {
	s.samlMu.RLock()
	defer s.samlMu.RUnlock()
	return s.saml
}

func (s *Server) setSAMLService(service *auth.SAMLService) {
	s.samlMu.Lock()
	s.saml = service
	s.samlMu.Unlock()
}

func (s *Server) samlEnabled() bool {
	return s.samlService() != nil
}

func NewServer(ctx context.Context, cfg *config.Config, store *store.Store, santa *santa.Service, sessions *auth.SessionManager, broadcaster *events.Broadcaster, logger *slog.Logger) (*Server, error) {
	// Check if SAML is enabled before trying to initialise SAML service
	samlSettings, err := store.GetSAMLSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get SAML settings: %w", err)
	}

	var samlService *auth.SAMLService
	if samlSettings.Enabled {
		samlService, err = auth.NewSAMLServiceFromSettings(ctx, cfg, samlSettings)
		if err != nil {
			return nil, fmt.Errorf("init saml: %w", err)
		}
	}

	if logger == nil {
		logger = slog.Default()
	}

	srv := &Server{
		cfg:         cfg,
		store:       store,
		santa:       santa,
		sessions:    sessions,
		broadcaster: broadcaster,
		logger:      logger.With("component", "api"),
	}
	srv.setSAMLService(samlService)
	return srv, nil
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	if len(s.cfg.AllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   s.cfg.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/login", s.handleLogin)
		r.Get("/providers", s.handleAuthProviders)
		r.Post("/login/local", s.handleLocalLogin)
		r.Get("/callback", s.handleCallback)
		r.Post("/callback", s.handleCallback)
		r.Post("/logout", s.handleLogout)
		r.Get("/metadata", s.handleMetadata)
		r.With(s.sessions.RequireAuth).Get("/me", s.handleMe)
	})

	r.Route("/api/santa/v1", func(r chi.Router) {
		// Santa sync protocol endpoints - no authentication required as per protocol
		r.Post("/preflight/{machine_id}", s.handleSantaPreflight)
		r.Post("/eventupload/{machine_id}", s.handleSantaEventUpload)
		r.Post("/ruledownload/{machine_id}", s.handleSantaRuleDownload)
		r.Post("/postflight/{machine_id}", s.handleSantaPostflight)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(s.sessions.RequireAuth)

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(s.sessions.RequireAdmin)
			r.Get("/apps", s.handleListApplications)
			r.Get("/apps/check", s.handleCheckApplicationExists)
			r.Post("/apps", s.handleCreateApplication)
			r.Delete("/apps/{id}", s.handleDeleteApplication)
			r.Get("/apps/{id}/scopes", s.handleListApplicationScopes)
			r.Post("/apps/{id}/scopes", s.handleCreateScope)
			r.Delete("/apps/{id}/scopes/{scopeID}", s.handleDeleteScope)
			r.Get("/groups", s.handleListGroups)
			r.Get("/groups/memberships", s.handleListGroupMemberships)
			r.Get("/users", s.handleListUsers)
			r.Get("/users/{id}", s.handleGetUserDetails)
			r.Get("/devices", s.handleListDevices)
			r.Get("/events/blocked", s.handleListBlocked)
			r.Get("/events/blocked/stream", s.handleEventStream)
			// Settings endpoints
			r.Get("/settings/saml", s.handleGetSAMLSettings)
			r.Put("/settings/saml", s.handleUpdateSAMLSettings)
			r.Get("/settings/santa-config", s.handleGetSantaConfig)
		})
	})

	return r
}

func (s *Server) handleAuthProviders(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		SAML  bool `json:"saml"`
		Local bool `json:"local"`
	}{
		SAML:  s.samlEnabled(),
		Local: true,
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	samlSvc := s.samlService()
	if samlSvc == nil {
		http.Error(w, "saml not configured", http.StatusServiceUnavailable)
		return
	}
	sess, err := s.sessions.Session(r)
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	state := randomString()
	sess.Values["saml_relay"] = state
	if err := sess.Save(r, w); err != nil {
		http.Error(w, "session save", http.StatusInternalServerError)
		return
	}
	authURL, err := samlSvc.BuildAuthURL(state)
	if err != nil {
		http.Error(w, "saml redirect", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	samlSvc := s.samlService()
	if samlSvc == nil {
		http.Error(w, "saml not configured", http.StatusServiceUnavailable)
		return
	}
	ctx := r.Context()
	sess, err := s.sessions.Session(r)
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid response", http.StatusBadRequest)
		return
	}

	expectedRelay, _ := sess.Values["saml_relay"].(string)
	relayState := r.FormValue("RelayState")
	if expectedRelay != "" && relayState != expectedRelay {
		http.Error(w, "invalid relay state", http.StatusBadRequest)
		return
	}
	samlResponse := r.FormValue("SAMLResponse")
	if samlResponse == "" {
		http.Error(w, "missing saml response", http.StatusBadRequest)
		return
	}

	assertion, err := samlSvc.ParseAssertion(samlResponse)
	if err != nil {
		http.Error(w, "invalid saml assertion", http.StatusUnauthorized)
		return
	}

	identity := samlSvc.ExtractIdentity(assertion)
	if identity.ExternalID == "" {
		http.Error(w, "missing object identifier", http.StatusForbidden)
		return
	}
	principal := firstNonEmpty(identity.Principal, identity.Email, identity.RawNameID)
	if principal == "" {
		http.Error(w, "missing principal", http.StatusForbidden)
		return
	}
	displayName := firstNonEmpty(identity.DisplayName, identity.RawNameID, principal)
	email := identity.Email
	if email == "" {
		email = principal
	}

	user, err := s.store.UpsertCloudUserByExternal(ctx, identity.ExternalID, principal, displayName, email)
	if err != nil {
		http.Error(w, "user upsert failed", http.StatusInternalServerError)
		return
	}

	if err := s.sessions.SetUser(w, r, user.ID); err != nil {
		http.Error(w, "set session", http.StatusInternalServerError)
		return
	}
	delete(sess.Values, "saml_relay")
	_ = sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Clear(w, r); err != nil {
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := s.store.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "load user", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleMetadata(w http.ResponseWriter, r *http.Request) {
	samlSvc := s.samlService()
	if samlSvc == nil {
		http.Error(w, "saml not configured", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(samlSvc.Metadata())
}

func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := s.store.ListApplications(r.Context())
	if err != nil {
		http.Error(w, "list apps", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, apps)
}

func (s *Server) handleCheckApplicationExists(w http.ResponseWriter, r *http.Request) {
	identifier := r.URL.Query().Get("identifier")
	if identifier == "" {
		http.Error(w, "identifier parameter required", http.StatusBadRequest)
		return
	}

	app, err := s.store.GetApplicationByIdentifier(r.Context(), identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		s.logger.Error("Failed to check application existence", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string `json:"name"`
		RuleType    string `json:"rule_type"`
		Identifier  string `json:"identifier"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if payload.Name == "" || payload.Identifier == "" {
		http.Error(w, "name and identifier required", http.StatusBadRequest)
		return
	}
	if payload.RuleType == "" {
		payload.RuleType = "BINARY"
	}

	// Check if application with this identifier already exists
	existing, err := s.store.GetApplicationByIdentifier(r.Context(), payload.Identifier)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("Failed to check for existing application", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		errorResponse := map[string]interface{}{
			"error":   "DUPLICATE_IDENTIFIER",
			"message": fmt.Sprintf("An application rule with identifier '%s' already exists", payload.Identifier),
			"existing_application": map[string]interface{}{
				"id":   existing.ID,
				"name": existing.Name,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	app := models.Application{
		Name:        payload.Name,
		RuleType:    payload.RuleType,
		Identifier:  payload.Identifier,
		Description: payload.Description,
	}
	created, err := s.store.CreateApplication(r.Context(), app)
	if err != nil {
		s.logger.Error("Failed to create application", "error", err)
		http.Error(w, "create app", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleDeleteApplication(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteApplication(r.Context(), appID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete app", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListApplicationScopes(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	scopes, err := s.store.ListApplicationScopes(r.Context(), appID)
	if err != nil {
		http.Error(w, "list scopes", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, scopes)
}

func (s *Server) handleCreateScope(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var payload struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Action     string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if payload.TargetType == "" || payload.TargetID == "" {
		http.Error(w, "target required", http.StatusBadRequest)
		return
	}
	if payload.Action == "" {
		payload.Action = "allow"
	}
	targetID, err := uuid.Parse(payload.TargetID)
	if err != nil {
		http.Error(w, "invalid target", http.StatusBadRequest)
		return
	}
	scope := models.ApplicationScope{
		ApplicationID: appID,
		TargetType:    payload.TargetType,
		TargetID:      targetID,
		Action:        payload.Action,
	}
	created, err := s.store.AddApplicationScope(r.Context(), scope)
	if err != nil {
		http.Error(w, "create scope", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleDeleteScope(w http.ResponseWriter, r *http.Request) {
	scopeID, err := uuid.Parse(chi.URLParam(r, "scopeID"))
	if err != nil {
		http.Error(w, "invalid scope id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteApplicationScope(r.Context(), scopeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete scope", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.store.ListGroups(r.Context())
	if err != nil {
		http.Error(w, "list groups", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, groups)
}

func (s *Server) handleListGroupMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := s.store.ListGroupMemberships(r.Context())
	if err != nil {
		http.Error(w, "list group memberships", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, memberships)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context(), 0)
	if err != nil {
		http.Error(w, "list users", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, users)
}

type userDetailResponse struct {
	User         *models.User        `json:"user"`
	Groups       []models.Group      `json:"groups"`
	Devices      []models.Host       `json:"devices"`
	RecentEvents []models.UserEvent  `json:"recent_events"`
	Policies     []models.UserPolicy `json:"policies"`
}

func (s *Server) handleGetUserDetails(w http.ResponseWriter, r *http.Request) {
	userIDParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		http.Error(w, "get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	roles, err := s.store.GetUserRoles(ctx, userID)
	if err != nil {
		http.Error(w, "get user roles", http.StatusInternalServerError)
		return
	}
	user.RoleGroups = roles

	groups, err := s.store.GroupsForUser(ctx, userID)
	if err != nil {
		http.Error(w, "get user groups", http.StatusInternalServerError)
		return
	}

	groupIDs := make([]uuid.UUID, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}

	devices, err := s.store.HostsForUser(ctx, userID)
	if err != nil {
		http.Error(w, "get user devices", http.StatusInternalServerError)
		return
	}

	events, err := s.store.RecentUserEvents(ctx, userID, 10)
	if err != nil {
		http.Error(w, "get user events", http.StatusInternalServerError)
		return
	}

	policies, err := s.store.UserPoliciesForUser(ctx, userID, groupIDs)
	if err != nil {
		http.Error(w, "get user policies", http.StatusInternalServerError)
		return
	}

	resp := userDetailResponse{
		User:         user,
		Groups:       groups,
		Devices:      devices,
		RecentEvents: events,
		Policies:     policies,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	hosts, err := s.store.ListHosts(r.Context())
	if err != nil {
		http.Error(w, "list devices", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, hosts)
}

func (s *Server) handleListBlocked(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.RecentBlockedEvents(r.Context(), 100)
	if err != nil {
		http.Error(w, "list events", http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleEventStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, cancel := s.broadcaster.Subscribe(s.cfg.SSEBufferSize)
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) handleSantaPreflight(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type for Santa protocol compliance
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	machineID := chi.URLParam(r, "machine_id")
	if machineID == "" {
		http.Error(w, "machine_id required", http.StatusBadRequest)
		return
	}

	var req models.PreflightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	resp, err := s.santa.Preflight(r.Context(), machineID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSantaEventUpload(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type for Santa protocol compliance
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	machineID := chi.URLParam(r, "machine_id")
	if machineID == "" {
		http.Error(w, "machine_id required", http.StatusBadRequest)
		return
	}

	var req models.EventUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	resp, err := s.santa.EventUpload(r.Context(), machineID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Broadcast events to UI
	for _, event := range req.Events {
		if strings.Contains(string(event.Decision), "BLOCK") {
			payload, _ := json.Marshal(event)
			s.broadcaster.Publish(payload)
		}
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSantaRuleDownload(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type for Santa protocol compliance
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	machineID := chi.URLParam(r, "machine_id")
	if machineID == "" {
		http.Error(w, "machine_id required", http.StatusBadRequest)
		return
	}

	var req models.RuleDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	resp, err := s.santa.RuleDownload(r.Context(), machineID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSantaPostflight(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type for Santa protocol compliance
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	machineID := chi.URLParam(r, "machine_id")
	if machineID == "" {
		http.Error(w, "machine_id required", http.StatusBadRequest)
		return
	}

	var req models.PostflightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	resp, err := s.santa.Postflight(r.Context(), machineID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("encode json", "error", err)
	}
}

func randomString() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Settings handlers

func (s *Server) handleGetSAMLSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.GetSAMLSettings(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get SAML settings: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleUpdateSAMLSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SAMLSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	if settings.Enabled {
		newService, err := auth.NewSAMLServiceFromSettings(ctx, s.cfg, &settings)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to validate SAML settings: %v", err), http.StatusBadRequest)
			return
		}

		if err := s.store.UpdateSAMLSettings(ctx, &settings); err != nil {
			http.Error(w, fmt.Sprintf("failed to update SAML settings: %v", err), http.StatusInternalServerError)
			return
		}

		s.setSAMLService(newService)
	} else {
		if err := s.store.UpdateSAMLSettings(ctx, &settings); err != nil {
			http.Error(w, fmt.Sprintf("failed to update SAML settings: %v", err), http.StatusInternalServerError)
			return
		}
		s.setSAMLService(nil)
	}

	s.writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleGetSantaConfig(w http.ResponseWriter, r *http.Request) {
	// Generate Santa client configuration XML
	config := santa.GenerateConfigXML(r)

	resp := struct {
		XML string `json:"xml"`
	}{
		XML: config,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// Local authentication handler

func (s *Server) handleLocalLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := s.store.AuthenticateLocalUser(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session
	if err := s.sessions.SetUser(w, r, user.ID); err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
