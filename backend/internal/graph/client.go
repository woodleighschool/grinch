package graph

import (
	"context"
	"errors"
	"fmt"

	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

// ErrNotConfigured indicates Azure Graph credentials were not provided.
var ErrNotConfigured = errors.New("graph: client not configured")

// Client wraps the Microsoft Graph SDK and tracks if it should be used.
type Client struct {
	graph   *msgraphsdk.GraphServiceClient
	enabled bool
}

// NewClient authenticates with Entra ID and returns a Graph client.
func NewClient(ctx context.Context, tenantID, clientID, clientSecret string) (*Client, error) {
	if tenantID == "" || clientID == "" || clientSecret == "" {
		return &Client{enabled: false}, nil
	}
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("graph credential: %w", err)
	}
	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("graph client: %w", err)
	}
	return &Client{graph: graphClient, enabled: true}, nil
}

// Enabled reports whether the client has been configured with credentials.
func (c *Client) Enabled() bool {
	return c.enabled
}
