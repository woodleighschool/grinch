package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/woodleighschool/grinch/backend/internal/config"
)

type OAuthUserInfo struct {
	ID                string `json:"id"`
	Email             string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
}

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OAuthService provides Microsoft OAuth authentication
type OAuthService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(cfg *config.Config) (*OAuthService, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	if cfg.AzureTenantID == "" {
		return nil, errors.New("azure tenant ID is required")
	}
	if cfg.AzureClientID == "" {
		return nil, errors.New("azure client ID is required")
	}
	if cfg.AzureClientSecret == "" {
		return nil, errors.New("azure client secret is required")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &OAuthService{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

// BuildAuthURL creates the authorization URL for Microsoft OAuth flow
func (s *OAuthService) BuildAuthURL(req *http.Request, state string) (string, error) {
	authorizeURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", s.cfg.AzureTenantID)

	params := url.Values{}
	params.Set("client_id", s.cfg.AzureClientID)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email User.Read")
	params.Set("redirect_uri", s.getRedirectURL(req))
	params.Set("state", state)
	params.Set("response_mode", "query")

	authURL := authorizeURL + "?" + params.Encode()
	return authURL, nil
}

// ExchangeCodeForToken exchanges authorization code for access token
func (s *OAuthService) ExchangeCodeForToken(ctx context.Context, req *http.Request, code string) (*OAuthTokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", s.cfg.AzureTenantID)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.cfg.AzureClientID)
	data.Set("client_secret", s.cfg.AzureClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.getRedirectURL(req))

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves user information using the access token
func (s *OAuthService) GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	userInfoURL := "https://graph.microsoft.com/v1.0/me"

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status %d", resp.StatusCode)
	}

	var userInfo OAuthUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decode userinfo response: %w", err)
	}

	return &userInfo, nil
}

// getRedirectURL returns the configured OAuth redirect URL
func (s *OAuthService) getRedirectURL(req *http.Request) string {
	forwardedHost := req.Header.Get("X-Forwarded-Host")
	host := req.Host
	if forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("https://%s/api/auth/oauth/callback", strings.TrimSpace(host))
}

// GenerateOAuthState generates a random state parameter for OAuth flow
func GenerateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ExtractIdentity extracts user identity information from Microsoft Graph user info
func (s *OAuthService) ExtractIdentity(userInfo *OAuthUserInfo) OAuthIdentity {
	email := userInfo.Email
	if email == "" {
		email = userInfo.UserPrincipalName
	}

	displayName := userInfo.DisplayName
	if displayName == "" && userInfo.GivenName != "" && userInfo.Surname != "" {
		displayName = userInfo.GivenName + " " + userInfo.Surname
	}

	return OAuthIdentity{
		ExternalID:  userInfo.ID,
		Principal:   userInfo.UserPrincipalName,
		DisplayName: displayName,
		Email:       email,
		Sub:         userInfo.ID,
	}
}

type OAuthIdentity struct {
	ExternalID  string
	Principal   string
	DisplayName string
	Email       string
	Sub         string
}
