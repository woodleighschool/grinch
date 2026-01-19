package apihttp

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/go-pkgz/auth/v2"
	"github.com/go-pkgz/auth/v2/avatar"
	"github.com/go-pkgz/auth/v2/provider"
	"github.com/go-pkgz/auth/v2/token"
	"golang.org/x/oauth2/microsoft"
)

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	AppName   string
	SecretKey string
	BaseURL   string

	MicrosoftClientID     string
	MicrosoftClientSecret string
	MicrosoftTenantID     string

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
		Validator: token.ValidatorFunc(func(_ string, claims token.Claims) bool {
			return claims.User != nil
		}),
	}

	// If /tmp is writable, use it for avatar storage.
	if isWritableDir("/tmp") {
		opts.AvatarStore = avatar.NewLocalFS("/tmp")
	}

	service := auth.NewService(opts)

	// Microsoft OAuth
	// Could add dynamic OAuth maybe?
	if cfg.MicrosoftClientID != "" && cfg.MicrosoftClientSecret != "" {
		tenantID := cfg.MicrosoftTenantID
		if tenantID == "" {
			tenantID = "common"
		}

		service.AddCustomProvider("microsoft", auth.Client{
			Cid:     cfg.MicrosoftClientID,
			Csecret: cfg.MicrosoftClientSecret,
		}, provider.CustomHandlerOpt{
			Endpoint: microsoft.AzureADEndpoint(tenantID),
			Scopes:   []string{"User.Read"},
			InfoURL:  "https://graph.microsoft.com/v1.0/me",
			MapUserFn: func(data provider.UserData, _ []byte) token.User {
				return token.User{
					ID:      "microsoft_" + token.HashID(sha256.New(), data.Value("id")),
					Name:    data.Value("displayName"),
					Picture: "https://graph.microsoft.com/beta/me/photo/$value",
				}
			},
		})
	}

	// Local auth is only enabled when an admin password is provided.
	if cfg.AdminPassword != "" {
		service.AddDirectProvider("local", provider.CredCheckerFunc(func(user, password string) (bool, error) {
			return user == "admin" && password == cfg.AdminPassword, nil
		}))
	}

	return service, nil
}

func isWritableDir(path string) bool {
	f, err := os.CreateTemp(path, ".writetest-*")
	if err != nil {
		return false
	}
	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(f.Name())
		return false
	}
	_ = os.Remove(f.Name())
	return true
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
