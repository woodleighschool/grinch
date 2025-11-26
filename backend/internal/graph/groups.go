package graph

import (
	"context"
	"fmt"

	msgraphgroups "github.com/microsoftgraph/msgraph-sdk-go/groups"
)

// DirectoryGroup represents the subset of Entra ID fields we care about.
type DirectoryGroup struct {
	ObjectID    string
	DisplayName string
	Description string
	Members     []string
}

// FetchGroups enumerates Entra ID security groups plus their members.
func (c *Client) FetchGroups(ctx context.Context) ([]DirectoryGroup, error) {
	if !c.enabled {
		return nil, ErrNotConfigured
	}
	if c.graph == nil {
		return nil, fmt.Errorf("graph client missing")
	}
	builder := c.graph.Groups()
	adapter := c.graph.GetAdapter()
	top := int32(100)
	selectFields := []string{"id", "displayName", "description"}
	var groups []DirectoryGroup
	for {
		resp, err := builder.Get(ctx, &msgraphgroups.GroupsRequestBuilderGetRequestConfiguration{
			QueryParameters: &msgraphgroups.GroupsRequestBuilderGetQueryParameters{
				Top:    &top,
				Select: selectFields,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list groups: %w", err)
		}
		for _, item := range resp.GetValue() {
			if item == nil {
				continue
			}
			groupID := deref(item.GetId())
			members, err := c.fetchGroupMembers(ctx, groupID)
			if err != nil {
				return nil, fmt.Errorf("fetch members for %s: %w", groupID, err)
			}
			groups = append(groups, DirectoryGroup{
				ObjectID:    groupID,
				DisplayName: deref(item.GetDisplayName()),
				Description: deref(item.GetDescription()),
				Members:     members,
			})
		}
		next := resp.GetOdataNextLink()
		if next == nil || len(*next) == 0 {
			break
		}
		builder = msgraphgroups.NewGroupsRequestBuilder(*next, adapter)
	}
	return groups, nil
}

// fetchGroupMembers walks the paginated graph API to collect member IDs.
func (c *Client) fetchGroupMembers(ctx context.Context, groupID string) ([]string, error) {
	if groupID == "" {
		return nil, nil
	}
	builder := c.graph.Groups().ByGroupId(groupID).TransitiveMembers()
	adapter := c.graph.GetAdapter()
	top := int32(100)
	selectFields := []string{"id"}
	var members []string
	for {
		resp, err := builder.Get(ctx, &msgraphgroups.ItemTransitiveMembersRequestBuilderGetRequestConfiguration{
			QueryParameters: &msgraphgroups.ItemTransitiveMembersRequestBuilderGetQueryParameters{
				Top:    &top,
				Select: selectFields,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, member := range resp.GetValue() {
			if member == nil {
				continue
			}
			if id := deref(member.GetId()); id != "" {
				members = append(members, id)
			}
		}
		next := resp.GetOdataNextLink()
		if next == nil || len(*next) == 0 {
			break
		}
		builder = msgraphgroups.NewItemTransitiveMembersRequestBuilder(*next, adapter)
	}
	return members, nil
}
