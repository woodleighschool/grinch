package authhttp

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-pkgz/auth/v2"
	"github.com/go-pkgz/auth/v2/avatar"
	"github.com/go-pkgz/auth/v2/provider"
	"github.com/go-pkgz/auth/v2/token"
	"golang.org/x/oauth2/microsoft"
)

const (
	sessionCookieName = "grinch_session"
	xsrfCookieName    = "grinch_xsrf"

	localProviderName     = "local"
	microsoftProviderName = "microsoft"

	defaultTokenDuration  = 12 * time.Hour
	defaultCookieDuration = 7 * 24 * time.Hour
	localAdminUsername    = "admin"
)

var errNoAuthProviders = errors.New("no auth providers registered")

type Config struct {
	RootURL string

	EntraTenantID     string
	EntraClientID     string
	EntraClientSecret string

	JWTSecret          string
	LocalAdminPassword string
}

type Service struct {
	auth    *auth.Service
	session func(http.Handler) http.Handler
}

func New(cfg Config) (*Service, error) {
	svc := auth.NewService(authOptions(cfg))

	registerMicrosoftProvider(svc, cfg)
	registerLocalProvider(svc, cfg)

	if len(svc.Providers()) == 0 {
		return nil, errNoAuthProviders
	}

	mw := svc.Middleware()
	return &Service{auth: svc, session: mw.Auth}, nil
}

func (s *Service) RegisterRoutes(r chi.Router) {
	authHandler, avatarHandler := s.auth.Handlers()

	r.Handle("/avatar/*", avatarHandler)
	r.Handle("/*", authHandler)
}

func (s *Service) SessionAuthMiddleware() func(http.Handler) http.Handler {
	return s.session
}

func authOptions(cfg Config) auth.Opts {
	return auth.Opts{
		SecretReader: token.SecretFunc(func(_ string) (string, error) {
			return cfg.JWTSecret, nil
		}),
		TokenDuration:  defaultTokenDuration,
		CookieDuration: defaultCookieDuration,
		Issuer:         "grinch",
		URL:            cfg.RootURL,
		SecureCookies:  hasHTTPSRootURL(cfg.RootURL),
		SameSiteCookie: http.SameSiteLaxMode,
		JWTCookieName:  sessionCookieName,
		XSRFCookieName: xsrfCookieName,
		Validator: token.ValidatorFunc(func(_ string, claims token.Claims) bool {
			return claims.User != nil
		}),
		AvatarStore:     avatar.NewNoOp(),
		AvatarRoutePath: "/auth/avatar",
	}
}

func registerMicrosoftProvider(svc *auth.Service, cfg Config) {
	if cfg.EntraTenantID == "" || cfg.EntraClientID == "" || cfg.EntraClientSecret == "" {
		return
	}

	svc.AddCustomProvider(
		microsoftProviderName,
		auth.Client{
			Cid:     cfg.EntraClientID,
			Csecret: cfg.EntraClientSecret,
		},
		provider.CustomHandlerOpt{
			Endpoint: microsoft.AzureADEndpoint(cfg.EntraTenantID),
			Scopes:   []string{"User.Read"},
			InfoURL:  "https://graph.microsoft.com/v1.0/me",
			MapUserFn: func(data provider.UserData, _ []byte) token.User {
				return token.User{
					ID:      microsoftProviderName + "_" + token.HashID(sha256.New(), data.Value("id")),
					Name:    data.Value("displayName"),
					Email:   data.Value("userPrincipalName"),
					Picture: "https://graph.microsoft.com/beta/me/photo/$value",
				}
			},
		},
	)
}

func registerLocalProvider(svc *auth.Service, cfg Config) {
	if cfg.LocalAdminPassword == "" {
		return
	}

	svc.AddDirectProvider(
		localProviderName,
		provider.CredCheckerFunc(func(user, password string) (bool, error) {
			return user == localAdminUsername && password == cfg.LocalAdminPassword, nil
		}),
	)
}

func hasHTTPSRootURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), "https://")
}
