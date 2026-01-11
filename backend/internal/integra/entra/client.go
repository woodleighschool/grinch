// Package entra integrates with Microsoft Entra ID via the Microsoft Graph API.
package entra

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphgroups "github.com/microsoftgraph/msgraph-sdk-go/groups"
	msgraphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

const maxGraphAPIPageSize = int32(999)

// Config holds the configuration required to authenticate to Microsoft Graph.
type Config struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

// Client wraps a Microsoft Graph client.
type Client struct {
	graph *msgraphsdk.GraphServiceClient
}

// User is a minimal Entra user record returned by Graph.
type User struct {
	ID          uuid.UUID
	DisplayName string
	UPN         string
}

// Group is a minimal Entra group record returned by Graph.
type Group struct {
	ID          uuid.UUID
	DisplayName string
	Description string
}

// NewClient creates a Graph client using client credential authentication.
func NewClient(cfg Config) (*Client, error) {
	cred, err := azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}

	g, err := msgraphsdk.NewGraphServiceClientWithCredentials(
		cred,
		[]string{"https://graph.microsoft.com/.default"},
	)
	if err != nil {
		return nil, fmt.Errorf("create graph client: %w", err)
	}

	return &Client{graph: g}, nil
}

// FetchUsers returns enabled member users.
func (c *Client) FetchUsers(ctx context.Context) ([]User, error) {
	builder := c.graph.Users()
	adapter := c.graph.GetAdapter()

	top := maxGraphAPIPageSize
	selectFields := []string{"id", "displayName", "userPrincipalName", "accountEnabled", "userType"}
	filter := "accountEnabled eq true and userType eq 'Member'"
	count := true

	var users []User
	for {
		resp, err := builder.Get(ctx, &msgraphusers.UsersRequestBuilderGetRequestConfiguration{
			Headers: advancedQueryHeaders(),
			QueryParameters: &msgraphusers.UsersRequestBuilderGetQueryParameters{
				Top:    &top,
				Select: selectFields,
				Filter: &filter,
				Count:  &count,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}

		for _, u := range resp.GetValue() {
			user, ok := userFromGraph(u)
			if ok {
				users = append(users, user)
			}
		}

		next := resp.GetOdataNextLink()
		if next == nil || strings.TrimSpace(*next) == "" {
			break
		}
		builder = msgraphusers.NewUsersRequestBuilder(*next, adapter)
	}

	return users, nil
}

// FetchGroups returns security enabled groups.
func (c *Client) FetchGroups(ctx context.Context) ([]Group, error) {
	builder := c.graph.Groups()
	adapter := c.graph.GetAdapter()

	top := maxGraphAPIPageSize
	selectFields := []string{"id", "displayName", "description"}
	filter := "securityEnabled eq true"
	count := true

	var groups []Group
	for {
		resp, err := builder.Get(ctx, &msgraphgroups.GroupsRequestBuilderGetRequestConfiguration{
			Headers: advancedQueryHeaders(),
			QueryParameters: &msgraphgroups.GroupsRequestBuilderGetQueryParameters{
				Top:    &top,
				Select: selectFields,
				Filter: &filter,
				Count:  &count,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list groups: %w", err)
		}

		for _, g := range resp.GetValue() {
			if g == nil {
				continue
			}
			groups = append(groups, Group{
				ID:          uuidFromStringPtr(g.GetId()),
				DisplayName: stringOrEmpty(g.GetDisplayName()),
				Description: stringOrEmpty(g.GetDescription()),
			})
		}

		next := resp.GetOdataNextLink()
		if next == nil || strings.TrimSpace(*next) == "" {
			break
		}
		builder = msgraphgroups.NewGroupsRequestBuilder(*next, adapter)
	}

	return groups, nil
}

// FetchGroupMembers returns enabled member user IDs for the group's transitive membership.
func (c *Client) FetchGroupMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	builder := c.graph.Groups().ByGroupId(groupID.String()).TransitiveMembers().GraphUser()

	top := maxGraphAPIPageSize
	selectFields := []string{"id", "accountEnabled", "userType"}
	filter := "accountEnabled eq true and userType eq 'Member'"
	count := true

	var ids []uuid.UUID
	for {
		resp, err := builder.Get(
			ctx,
			&msgraphgroups.ItemTransitiveMembersGraphUserRequestBuilderGetRequestConfiguration{
				Headers: advancedQueryHeaders(),
				QueryParameters: &msgraphgroups.ItemTransitiveMembersGraphUserRequestBuilderGetQueryParameters{
					Top:    &top,
					Select: selectFields,
					Filter: &filter,
					Count:  &count,
				},
			},
		)
		if err != nil {
			return nil, fmt.Errorf("list transitive members: %w", err)
		}

		ids = append(ids, extractMemberIDs(resp.GetValue())...)

		next := resp.GetOdataNextLink()
		if next == nil || strings.TrimSpace(*next) == "" {
			break
		}
		builder = builder.WithUrl(*next)
	}

	return ids, nil
}
