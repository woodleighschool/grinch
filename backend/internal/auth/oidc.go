package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider wraps the oauth2 + ID token verifier used for admin authentication.
type OIDCProvider struct {
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
}

// NewOIDCProvider discovers the remote provider and prepares OAuth config.
func NewOIDCProvider(ctx context.Context, issuer, clientID, clientSecret, siteBaseURL string) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover oidc provider: %w", err)
	}
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  siteBaseURL + "/api/auth/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return &OIDCProvider{verifier: verifier, oauth: config}, nil
}

// AuthCodeURL builds an authorization URL with optional request parameters.
func (p *OIDCProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth.AuthCodeURL(state, opts...)
}

// Exchange redeems the OAuth authorization code for an access token.
func (p *OIDCProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauth.Exchange(ctx, code)
}

// VerifyIDToken validates the ID token signature and claims.
func (p *OIDCProvider) VerifyIDToken(ctx context.Context, raw string) (*oidc.IDToken, error) {
	return p.verifier.Verify(ctx, raw)
}

// OAuth2Config exposes the underlying OAuth configuration for advanced flows.
func (p *OIDCProvider) OAuth2Config() *oauth2.Config {
	return p.oauth
}
