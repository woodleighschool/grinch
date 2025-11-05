package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"github.com/woodleighschool/grinch/backend/internal/store"
)

type contextKey string

const (
	userIDKey contextKey = "auth.user"
)

type SessionManager struct {
	store      *sessions.CookieStore
	cookieName string
	userStore  *store.Store
}

func NewSessionManager(secret []byte, name string) *SessionManager {
	store := sessions.NewCookieStore(secret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int((12 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	return &SessionManager{
		store:      store,
		cookieName: name,
		userStore:  nil, // Will be set later
	}
}

func (m *SessionManager) SetUserStore(userStore *store.Store) {
	m.userStore = userStore
}

func (m *SessionManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.GetUserID(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *SessionManager) GetUserID(r *http.Request) (uuid.UUID, error) {
	session, err := m.store.Get(r, m.cookieName)
	if err != nil {
		return uuid.Nil, err
	}
	raw, ok := session.Values["user_id"].(string)
	if !ok || raw == "" {
		return uuid.Nil, errors.New("user not in session")
	}
	return uuid.Parse(raw)
}

func (m *SessionManager) SetUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) error {
	session, err := m.store.Get(r, m.cookieName)
	if err != nil {
		return err
	}
	session.Values["user_id"] = id.String()
	return sessions.Save(r, w)
}

func (m *SessionManager) Clear(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, m.cookieName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return sessions.Save(r, w)
}

func (m *SessionManager) Session(r *http.Request) (*sessions.Session, error) {
	return m.store.Get(r, m.cookieName)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(userIDKey).(uuid.UUID)
	return raw, ok
}

func (m *SessionManager) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.GetUserID(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if m.userStore == nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		user, err := m.userStore.GetUser(r.Context(), userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		hasRole, err := m.userStore.UserHasRole(r.Context(), userID, "admin")
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !hasRole {
			http.Error(w, "forbidden: admin access required", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
