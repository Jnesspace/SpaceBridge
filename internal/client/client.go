// Package client provides a GraphQL client for the Spacelift API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	graphql "github.com/hasura/go-graphql-client"
	"github.com/jnesspace/spacebridge/internal/models"
	"github.com/jnesspace/spacebridge/pkg/config"
)

// Verbose controls whether verbose output is enabled.
var Verbose bool

// Client wraps the GraphQL client with Spacelift-specific functionality.
type Client struct {
	graphql *graphql.Client
	config  config.AccountConfig
}

// spaceliftTransport handles authentication for Spacelift API requests.
type spaceliftTransport struct {
	baseURL   string
	keyID     string
	secretKey string
	token     string
	tokenExp  time.Time
	base      http.RoundTripper
}

// tokenResponse represents the JWT token response from Spacelift.
type tokenResponse struct {
	Token string `json:"jwt"`
}

// RoundTrip implements http.RoundTripper with automatic token refresh.
func (t *spaceliftTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Refresh token if expired or not set
	if t.token == "" || time.Now().After(t.tokenExp) {
		if Verbose {
			fmt.Printf("[AUTH] Authenticating with Spacelift at %s...\n", t.baseURL)
		}
		if err := t.refreshToken(); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
		if Verbose {
			fmt.Printf("[AUTH] Successfully authenticated! Token expires in ~55 minutes\n")
		}
	}

	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

// refreshToken obtains a new JWT token from Spacelift.
func (t *spaceliftTransport) refreshToken() error {
	url := fmt.Sprintf("%s/graphql", t.baseURL)

	// Build the token mutation
	payload := map[string]interface{}{
		"query": `mutation GetToken($keyId: ID!, $keySecret: String!) {
			apiKeyUser(id: $keyId, secret: $keySecret) {
				jwt
			}
		}`,
		"variables": map[string]string{
			"keyId":     t.keyID,
			"keySecret": t.secretKey,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = &readCloser{data: body}
	req.ContentLength = int64(len(body))

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			APIKeyUser tokenResponse `json:"apiKeyUser"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("API error: %s", result.Errors[0].Message)
	}

	t.token = result.Data.APIKeyUser.Token
	t.tokenExp = time.Now().Add(55 * time.Minute) // Tokens valid for ~1 hour

	return nil
}

// readCloser is a helper to create an io.ReadCloser from bytes.
type readCloser struct {
	data []byte
	pos  int
}

func (r *readCloser) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *readCloser) Close() error {
	return nil
}

// New creates a new Spacelift GraphQL client.
func New(cfg config.AccountConfig) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	transport := &spaceliftTransport{
		baseURL:   cfg.URL,
		keyID:     cfg.KeyID,
		secretKey: cfg.SecretKey,
		base:      http.DefaultTransport,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	graphqlURL := fmt.Sprintf("%s/graphql", cfg.URL)
	client := graphql.NewClient(graphqlURL, httpClient)

	return &Client{
		graphql: client,
		config:  cfg,
	}, nil
}

// Query executes a GraphQL query.
func (c *Client) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	return c.graphql.Query(ctx, q, variables)
}

// Mutate executes a GraphQL mutation.
func (c *Client) Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error {
	return c.graphql.Mutate(ctx, m, variables)
}

// URL returns the Spacelift URL this client is connected to.
func (c *Client) URL() string {
	return c.config.URL
}

// EnableExternalStateAccess enables external state access on a stack.
func (c *Client) EnableExternalStateAccess(ctx context.Context, stack models.Stack) error {
	// The mutation needs all required fields from StackInput
	mutation := `mutation EnableExternalState(
		$id: ID!,
		$administrative: Boolean!,
		$branch: String!,
		$name: String!,
		$repository: String!
	) {
		stackUpdate(id: $id, input: {
			administrative: $administrative
			branch: $branch
			name: $name
			repository: $repository
			vendorConfig: {
				terraform: {
					externalStateAccessEnabled: true
				}
			}
		}) {
			id
		}
	}`

	var result struct {
		StackUpdate struct {
			ID string `json:"id"`
		} `json:"stackUpdate"`
	}

	variables := map[string]interface{}{
		"id":             stack.ID,
		"administrative": stack.Administrative,
		"branch":         stack.Branch,
		"name":           stack.Name,
		"repository":     stack.Repository,
	}

	if err := c.rawMutate(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("failed to enable external state access: %w", err)
	}

	return nil
}

// GetStateDownloadURL gets a pre-signed URL to download stack state.
func (c *Client) GetStateDownloadURL(ctx context.Context, stackID string) (string, error) {
	mutation := `mutation GetStateDownloadURL($stackId: ID!) {
		stateDownloadUrl(input: { stackId: $stackId }) {
			url
		}
	}`

	var result struct {
		StateDownloadURL struct {
			URL string `json:"url"`
		} `json:"stateDownloadUrl"`
	}

	variables := map[string]interface{}{
		"stackId": stackID,
	}

	if err := c.rawMutate(ctx, mutation, variables, &result); err != nil {
		return "", fmt.Errorf("failed to get state download URL: %w", err)
	}

	return result.StateDownloadURL.URL, nil
}

// StateUploadResult contains the upload URL and object ID.
type StateUploadResult struct {
	URL      string
	ObjectID string
}

// GetStateUploadURL gets a pre-signed URL to upload stack state.
func (c *Client) GetStateUploadURL(ctx context.Context, stackID string) (*StateUploadResult, error) {
	mutation := `mutation GetStateUploadURL {
		stateUploadUrl {
			url
			objectId
		}
	}`

	var result struct {
		StateUploadURL struct {
			URL      string `json:"url"`
			ObjectID string `json:"objectId"`
		} `json:"stateUploadUrl"`
	}

	if err := c.rawMutate(ctx, mutation, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get state upload URL: %w", err)
	}

	return &StateUploadResult{
		URL:      result.StateUploadURL.URL,
		ObjectID: result.StateUploadURL.ObjectID,
	}, nil
}

// LockStack locks a stack for exclusive access.
func (c *Client) LockStack(ctx context.Context, stackID string) error {
	mutation := `mutation LockStack($id: ID!) {
		stackLock(id: $id) {
			id
		}
	}`

	variables := map[string]interface{}{
		"id": stackID,
	}

	if err := c.rawMutate(ctx, mutation, variables, nil); err != nil {
		return fmt.Errorf("failed to lock stack: %w", err)
	}

	return nil
}

// UnlockStack unlocks a previously locked stack.
func (c *Client) UnlockStack(ctx context.Context, stackID string) error {
	mutation := `mutation UnlockStack($id: ID!) {
		stackUnlock(id: $id) {
			id
		}
	}`

	variables := map[string]interface{}{
		"id": stackID,
	}

	if err := c.rawMutate(ctx, mutation, variables, nil); err != nil {
		return fmt.Errorf("failed to unlock stack: %w", err)
	}

	return nil
}

// ImportManagedState triggers the state import on a stack.
func (c *Client) ImportManagedState(ctx context.Context, stackID string, objectID string) error {
	mutation := `mutation ImportManagedState($stackId: ID!, $state: String!) {
		stackManagedStateImport(stackId: $stackId, state: $state)
	}`

	variables := map[string]interface{}{
		"stackId": stackID,
		"state":   objectID,
	}

	if err := c.rawMutate(ctx, mutation, variables, nil); err != nil {
		return fmt.Errorf("failed to import managed state: %w", err)
	}

	return nil
}

// StreamStateFromURL downloads state from a URL and returns an io.ReadCloser.
// The caller is responsible for closing the reader.
func StreamStateFromURL(ctx context.Context, downloadURL string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download state: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

// UploadStateToURL uploads state data to a pre-signed URL.
func UploadStateToURL(ctx context.Context, uploadURL string, data io.Reader, contentLength int64) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, data)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.ContentLength = contentLength
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// EnableStack enables a disabled stack.
func (c *Client) EnableStack(ctx context.Context, stack models.Stack) error {
	mutation := `mutation EnableStack(
		$id: ID!,
		$administrative: Boolean!,
		$branch: String!,
		$name: String!,
		$repository: String!
	) {
		stackUpdate(id: $id, input: {
			administrative: $administrative
			branch: $branch
			name: $name
			repository: $repository
			isDisabled: false
		}) {
			id
		}
	}`

	var result struct {
		StackUpdate struct {
			ID string `json:"id"`
		} `json:"stackUpdate"`
	}

	variables := map[string]interface{}{
		"id":             stack.ID,
		"administrative": stack.Administrative,
		"branch":         stack.Branch,
		"name":           stack.Name,
		"repository":     stack.Repository,
	}

	if err := c.rawMutate(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("failed to enable stack: %w", err)
	}

	return nil
}

// rawMutate executes a raw GraphQL mutation string.
func (c *Client) rawMutate(ctx context.Context, mutation string, variables map[string]interface{}, result interface{}) error {
	payload := map[string]interface{}{
		"query":     mutation,
		"variables": variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal mutation: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/graphql", c.config.URL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = &readCloser{data: body}
	req.ContentLength = int64(len(body))

	// Use the graphql client's underlying http client for auth
	httpClient := &http.Client{
		Transport: &spaceliftTransport{
			baseURL:   c.config.URL,
			keyID:     c.config.KeyID,
			secretKey: c.config.SecretKey,
			base:      http.DefaultTransport,
		},
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mutation request failed: %w", err)
	}
	defer resp.Body.Close()

	var graphqlResult struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphqlResult); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(graphqlResult.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", graphqlResult.Errors[0].Message)
	}

	if result != nil {
		if err := json.Unmarshal(graphqlResult.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}
