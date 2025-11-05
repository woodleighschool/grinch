package entra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/backend/internal/config"
	"github.com/woodleighschool/grinch/backend/internal/models"
	"github.com/woodleighschool/grinch/backend/internal/store"
)

const graphBase = "https://graph.microsoft.com/v1.0"

type Service struct {
	cfg        *config.Config
	store      *store.Store
	credential *azidentity.ClientSecretCredential
	httpClient *http.Client
	logger     *slog.Logger
}

func NewService(cfg *config.Config, st *store.Store, logger *slog.Logger) (*Service, error) {
	cred, err := azidentity.NewClientSecretCredential(cfg.AzureTenantID, cfg.AzureClientID, cfg.AzureClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("create azure credential: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		cfg:        cfg,
		store:      st,
		credential: cred,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger.With("component", "entra"),
	}, nil
}

func (s *Service) Start(ctx context.Context, interval time.Duration) {
	go func() {
		s.logger.Info("starting Entra sync scheduler", "interval", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			s.logger.Debug("initiating scheduled Entra sync")
			if err := s.Sync(ctx); err != nil {
				s.logger.Error("periodic sync failed", "error", err)
			} else {
				s.logger.Debug("scheduled Entra sync completed")
			}
			select {
			case <-ctx.Done():
				s.logger.Debug("stopping Entra sync scheduler")
				return
			case <-ticker.C:
				s.logger.Debug("Entra sync ticker fired")
			}
		}
	}()
}

func (s *Service) Sync(ctx context.Context) error {
	start := time.Now()
	s.logger.Info("starting Entra sync cycle")
	token, err := s.credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://graph.microsoft.com/.default"}})
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	// First, get all current Entra users to track deletions
	entraSyncStart := time.Now()
	stats := struct {
		usersSynced                 int
		usersSkipped                int
		usersConverted              int
		usersEvaluatedForConversion int
		usersRemoved                int
		usersEvaluatedForRemoval    int
		groupsProcessed             int
		membershipsUpdated          int
	}{}

	// Sync all users from the tenant
	usersSynced, usersSkipped, err := s.syncAllUsers(ctx, token.Token)
	if err != nil {
		return fmt.Errorf("sync all users: %w", err)
	}
	stats.usersSynced = usersSynced
	stats.usersSkipped = usersSkipped
	s.logger.Debug("synchronised Entra users into store", "processed", usersSynced, "skipped", usersSkipped)

	// Mark this as a full Entra sync to identify cloud users
	usersConverted, usersEvaluated, err := s.convertLocalToCloudUsers(ctx, token.Token)
	if err != nil {
		return fmt.Errorf("convert local to cloud users: %w", err)
	}
	stats.usersConverted = usersConverted
	stats.usersEvaluatedForConversion = usersEvaluated
	s.logger.Debug("evaluated local users for conversion",
		"evaluated", usersEvaluated,
		"converted", usersConverted,
	)

	// Remove users that no longer exist in Entra
	usersRemoved, usersEvaluatedForRemoval, err := s.removeDeletedEntraUsers(ctx, token.Token, entraSyncStart)
	if err != nil {
		return fmt.Errorf("remove deleted entra users: %w", err)
	}
	stats.usersRemoved = usersRemoved
	stats.usersEvaluatedForRemoval = usersEvaluatedForRemoval
	s.logger.Debug("evaluated cloud users for removal",
		"evaluated", usersEvaluatedForRemoval,
		"removed", usersRemoved,
	)

	// Then sync groups and their memberships
	groups, err := s.fetchGroups(ctx, token.Token)
	if err != nil {
		return err
	}
	stats.groupsProcessed = len(groups)
	s.logger.Debug("fetched Entra groups", "count", len(groups))

	for _, g := range groups {
		s.logger.Debug("synchronising group", "group_id", g.ID, "display_name", g.DisplayName)
		groupModel, err := s.store.UpsertGroup(ctx, g.ID, g.DisplayName, g.Description)
		if err != nil {
			return fmt.Errorf("upsert group %s: %w", g.ID, err)
		}

		members, err := s.fetchGroupMembers(ctx, token.Token, g.ID)
		if err != nil {
			return fmt.Errorf("fetch members for %s: %w", g.ID, err)
		}
		s.logger.Debug("fetched group members", "group_id", g.ID, "member_count", len(members))
		stats.membershipsUpdated += len(members)
		memberIDs := make([]uuid.UUID, 0, len(members))
		for _, m := range members {
			user, err := s.store.UpsertCloudUserByExternal(ctx, m.ID, m.PrincipalName(), m.DisplayName, m.Mail)
			if err != nil {
				return fmt.Errorf("upsert user %s: %w", m.ID, err)
			}
			memberIDs = append(memberIDs, user.ID)
		}
		if err := s.store.ReplaceGroupMemberships(ctx, groupModel.ID, memberIDs); err != nil {
			return fmt.Errorf("replace memberships for %s: %w", g.ID, err)
		}
		s.logger.Debug("replaced group memberships", "group_id", g.ID, "member_count", len(memberIDs))
	}

	s.logger.Info("Entra sync cycle completed",
		"duration", time.Since(start),
		"users_synced", stats.usersSynced,
		"users_skipped", stats.usersSkipped,
		"users_converted", stats.usersConverted,
		"users_evaluated_for_conversion", stats.usersEvaluatedForConversion,
		"users_removed", stats.usersRemoved,
		"users_evaluated_for_removal", stats.usersEvaluatedForRemoval,
		"groups_processed", stats.groupsProcessed,
		"memberships_updated", stats.membershipsUpdated,
	)

	return nil
}

type graphGroup struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type graphMember struct {
	ID                string         `json:"id"`
	ODataType         string         `json:"@odata.type"`
	DisplayName       string         `json:"displayName"`
	UserPrincipalName string         `json:"userPrincipalName"`
	Mail              string         `json:"mail"`
	AccountEnabled    bool           `json:"accountEnabled"`
	AdditionalData    map[string]any `json:"-"`
}

func (m graphMember) PrincipalName() string {
	if m.UserPrincipalName != "" {
		return m.UserPrincipalName
	}
	return m.Mail
}

// shouldIncludeUser returns true if the user should be included (enabled and not external)
func (m graphMember) shouldIncludeUser() bool {
	if !m.AccountEnabled {
		return false
	}
	if strings.Contains(m.UserPrincipalName, "#EXT#") {
		return false
	}
	return true
}

func (s *Service) fetchGroups(ctx context.Context, token string) ([]graphGroup, error) {
	var groups []graphGroup
	url := graphBase + "/groups?$select=id,displayName,description"
	for url != "" {
		body, next, err := s.doGraphRequest(ctx, token, url)
		if err != nil {
			return nil, fmt.Errorf("request groups: %w", err)
		}
		var payload struct {
			Value []graphGroup `json:"value"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode groups: %w", err)
		}
		groups = append(groups, payload.Value...)
		url = next
	}
	return groups, nil
}

func (s *Service) fetchGroupMembers(ctx context.Context, token, groupID string) ([]graphMember, error) {
	// Query only user-type members and filter them
	var members []graphMember
	nextURL := fmt.Sprintf("%s/groups/%s/members/microsoft.graph.user?$select=id,displayName,mail,userPrincipalName,accountEnabled", graphBase, groupID)
	for nextURL != "" {
		body, next, err := s.doGraphRequest(ctx, token, nextURL)
		if err != nil {
			return nil, err
		}
		var payload struct {
			Value []graphMember `json:"value"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode members: %w", err)
		}
		for _, member := range payload.Value {
			if member.shouldIncludeUser() {
				members = append(members, member)
			}
		}
		nextURL = next
	}
	return members, nil
}

func (s *Service) doGraphRequest(ctx context.Context, token, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	start := time.Now()
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("graph request failed", "url", url, "error", err)
		return nil, "", err
	}
	defer resp.Body.Close()
	duration := time.Since(start)
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		s.logger.Warn("graph request returned error",
			"url", url,
			"status", resp.StatusCode,
			"duration", duration,
			"body", string(b),
		)
		return nil, "", fmt.Errorf("graph %s returned %d: %s", url, resp.StatusCode, string(b))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	var aux struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if err := json.Unmarshal(body, &aux); err != nil {
		return nil, "", err
	}
	s.logger.Debug("graph request succeeded", "url", url, "status", resp.StatusCode, "duration", duration, "next_link", aux.NextLink)
	return body, aux.NextLink, nil
}

// syncAllUsers fetches all users from the tenant and syncs them to the database
func (s *Service) syncAllUsers(ctx context.Context, token string) (int, int, error) {
	url := graphBase + "/users?$select=id,displayName,mail,userPrincipalName,accountEnabled"
	var processed, skipped int
	for url != "" {
		body, next, err := s.doGraphRequest(ctx, token, url)
		if err != nil {
			return processed, skipped, fmt.Errorf("request all users: %w", err)
		}
		var payload struct {
			Value []graphMember `json:"value"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return processed, skipped, fmt.Errorf("decode users: %w", err)
		}
		for _, user := range payload.Value {
			// Skip disabled users and external users
			if !user.shouldIncludeUser() {
				skipped++
				continue
			}
			if _, err := s.store.UpsertCloudUserByExternal(ctx, user.ID, user.PrincipalName(), user.DisplayName, user.Mail); err != nil {
				return processed, skipped, fmt.Errorf("upsert user %s: %w", user.ID, err)
			}
			processed++
		}
		url = next
	}
	return processed, skipped, nil
}

// convertLocalToCloudUsers checks if any local users now exist in Entra and converts them
func (s *Service) convertLocalToCloudUsers(ctx context.Context, token string) (int, int, error) {
	localUsers, err := s.store.ListLocalUsers(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("list local users: %w", err)
	}

	converted := 0
	for _, localUser := range localUsers {
		// Check if this user now exists in Entra by searching for their principal name
		convertedUser, err := s.checkAndConvertLocalUser(ctx, token, localUser)
		if err != nil {
			// Log but don't fail the entire sync
			s.logger.Warn("failed to check local user", "principal", localUser.PrincipalName, "error", err)
			continue
		}
		if convertedUser {
			converted++
		}
	}

	return converted, len(localUsers), nil
}

// checkAndConvertLocalUser checks if a local user exists in Entra and converts them if found
func (s *Service) checkAndConvertLocalUser(ctx context.Context, token string, localUser *models.User) (bool, error) {
	// Search for user by userPrincipalName and filter for enabled, non-external users
	// TO:DO - Fix potential OData injection vulnerability
	upn := strings.ReplaceAll(localUser.PrincipalName, "'", "''")
	u, _ := url.Parse(graphBase + "/users")
	q := url.Values{}
	q.Set("$filter", fmt.Sprintf("userPrincipalName eq '%s'", upn))
	q.Set("$select", "id,displayName,mail,userPrincipalName,accountEnabled")
	u.RawQuery = q.Encode()
	searchURL := u.String()

	body, _, err := s.doGraphRequest(ctx, token, searchURL)
	if err != nil {
		return false, fmt.Errorf("search for user %s: %w", localUser.PrincipalName, err)
	}

	var payload struct {
		Value []graphMember `json:"value"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Errorf("decode search results: %w", err)
	}

	if len(payload.Value) > 0 {
		entraMember := payload.Value[0]
		// Only convert if the user is enabled and not external
		if entraMember.shouldIncludeUser() {
			if err := s.store.ConvertLocalToCloudUser(ctx, localUser.ID, entraMember.ID, entraMember.DisplayName, entraMember.Mail); err != nil {
				return false, fmt.Errorf("convert local user to cloud: %w", err)
			}
			s.logger.Info("converted local user to cloud user", "principal", localUser.PrincipalName)
			return true, nil
		}
		s.logger.Debug("local user present in Entra but not eligible for conversion",
			"principal", localUser.PrincipalName,
			"account_enabled", entraMember.AccountEnabled,
			"external", strings.Contains(entraMember.UserPrincipalName, "#EXT#"),
		)
	}

	return false, nil
}

// removeDeletedEntraUsers removes users that no longer exist in Entra, or have become disabled/external
func (s *Service) removeDeletedEntraUsers(ctx context.Context, token string, syncStart time.Time) (int, int, error) {
	// Get all cloud users that haven't been synced in this sync cycle
	deletedUsers, err := s.store.GetCloudUsersNotSyncedSince(ctx, syncStart)
	if err != nil {
		return 0, 0, fmt.Errorf("get users not synced: %w", err)
	}

	removed := 0
	for _, user := range deletedUsers {
		// Double-check by looking up the user in Entra
		if user.ExternalID != nil {
			exists, err := s.verifyUserExistsInEntra(ctx, token, *user.ExternalID)
			if err != nil {
				s.logger.Warn("failed to verify user existence", "principal", user.PrincipalName, "error", err)
				continue
			}

			if !exists {
				if err := s.store.DeleteUser(ctx, user.ID); err != nil {
					s.logger.Error("failed to delete user", "principal", user.PrincipalName, "error", err)
				} else {
					s.logger.Info("deleted user not found in Entra or became disabled/external", "principal", user.PrincipalName)
					removed++
				}
			}
		} else {
			s.logger.Debug("skipping user removal check without external ID", "principal", user.PrincipalName)
		}
	}

	return removed, len(deletedUsers), nil
}

// verifyUserExistsInEntra checks if a user still exists in Entra and is enabled/non-external
func (s *Service) verifyUserExistsInEntra(ctx context.Context, token, externalID string) (bool, error) {
	url := fmt.Sprintf("%s/users/%s?$select=id,userPrincipalName,accountEnabled", graphBase, externalID)
	body, _, err := s.doGraphRequest(ctx, token, url)
	if err != nil {
		// If we get a 404, user doesn't exist
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	var user graphMember
	if err := json.Unmarshal(body, &user); err != nil {
		return false, fmt.Errorf("decode user: %w", err)
	}

	// User exists, but check if they should be included
	return user.shouldIncludeUser(), nil
}
