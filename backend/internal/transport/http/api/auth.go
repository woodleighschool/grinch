package apihttp

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-pkgz/auth/v2"
	"github.com/go-pkgz/auth/v2/avatar"
	"github.com/go-pkgz/auth/v2/provider"
	"github.com/go-pkgz/auth/v2/token"
)

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	AppName   string
	SecretKey string
	BaseURL   string

	MicrosoftClientID     string
	MicrosoftClientSecret string

	AdminPassword string

	TokenDuration  time.Duration
	CookieDuration time.Duration
}

// NewAuthService configures and returns a go-pkgz/auth service.
func NewAuthService(cfg AuthConfig) (*auth.Service, error) {
	// Just in case check, nothing like an empty secret key.
	if cfg.SecretKey == "" {
		return nil, errors.New("auth secret key is required")
	}

	opts := auth.Opts{
		SecretReader: token.SecretFunc(func(_ string) (string, error) {
			return cfg.SecretKey, nil
		}),
		TokenDuration:  cfg.TokenDuration,
		CookieDuration: cfg.CookieDuration,
		Issuer:         cfg.AppName,
		URL:            cfg.BaseURL,
		SameSiteCookie: http.SameSiteStrictMode,
		//DisableXSRF:    true,
		AvatarStore: avatar.NewLocalFS("/tmp"),
		Validator: token.ValidatorFunc(func(_ string, claims token.Claims) bool {
			return claims.User != nil
		}),
	}

	service := auth.NewService(opts)

	// Microsoft OAuth
	// Could add dynamic OAuth maybe?
	if cfg.MicrosoftClientID != "" && cfg.MicrosoftClientSecret != "" {
		service.AddProvider("microsoft", cfg.MicrosoftClientID, cfg.MicrosoftClientSecret)
	}

	// Local auth is only enabled when an admin password is provided.
	if cfg.AdminPassword != "" {
		service.AddDirectProvider("local", provider.CredCheckerFunc(func(user, password string) (bool, error) {
			return user == "admin" && password == cfg.AdminPassword, nil
		}))
	}

	return service, nil
}

// AuthMiddleware returns the authentication middleware.
func AuthMiddleware(service *auth.Service) func(http.Handler) http.Handler {
	m := service.Middleware()
	return m.Auth
}

// AuthHandlers returns the HTTP handlers for login, logout, and callbacks.
func AuthHandlers(service *auth.Service) (http.Handler, http.Handler) {
	return service.Handlers()
}
