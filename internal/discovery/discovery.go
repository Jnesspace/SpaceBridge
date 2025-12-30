// Package discovery handles resource discovery from Spacelift accounts.
package discovery

import (
	"context"
	"fmt"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/models"
)

// Service provides resource discovery capabilities.
type Service struct {
	client *client.Client
}

// New creates a new discovery service.
func New(c *client.Client) *Service {
	return &Service{client: c}
}

// DiscoverAll fetches all resources from the Spacelift account.
func (s *Service) DiscoverAll(ctx context.Context) (*Manifest, error) {
	manifest := &Manifest{
		SourceURL: s.client.URL(),
	}

	// Discover spaces first (foundation)
	spaces, err := s.DiscoverSpaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover spaces: %w", err)
	}
	manifest.Spaces = spaces

	// Discover contexts
	contexts, err := s.DiscoverContexts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover contexts: %w", err)
	}
	manifest.Contexts = contexts

	// Discover policies
	policies, err := s.DiscoverPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover policies: %w", err)
	}
	manifest.Policies = policies

	// Discover stacks
	stacks, err := s.DiscoverStacks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover stacks: %w", err)
	}
	manifest.Stacks = stacks

	// Discover AWS integrations
	awsIntegrations, err := s.DiscoverAWSIntegrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover AWS integrations: %w", err)
	}
	manifest.AWSIntegrations = awsIntegrations

	// Discover Azure integrations
	azureIntegrations, err := s.DiscoverAzureIntegrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover Azure integrations: %w", err)
	}
	manifest.AzureIntegrations = azureIntegrations

	// Discover integration attachments and associate with stacks
	// Build a map of stackID -> stack index for quick lookup
	stackIndex := make(map[string]int)
	for i, stack := range manifest.Stacks {
		stackIndex[stack.ID] = i
	}

	// Discover AWS integration attachments
	for _, integration := range awsIntegrations {
		attachments, err := s.DiscoverAWSIntegrationAttachments(ctx, integration.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to discover AWS integration attachments for %s: %w", integration.Name, err)
		}
		for stackID, attachment := range attachments {
			if idx, ok := stackIndex[stackID]; ok {
				manifest.Stacks[idx].AttachedAWSIntegrations = append(
					manifest.Stacks[idx].AttachedAWSIntegrations,
					attachment,
				)
			}
		}
	}

	// Discover Azure integration attachments
	for _, integration := range azureIntegrations {
		attachments, err := s.DiscoverAzureIntegrationAttachments(ctx, integration.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to discover Azure integration attachments for %s: %w", integration.Name, err)
		}
		for stackID, attachment := range attachments {
			if idx, ok := stackIndex[stackID]; ok {
				manifest.Stacks[idx].AttachedAzureIntegrations = append(
					manifest.Stacks[idx].AttachedAzureIntegrations,
					attachment,
				)
			}
		}
	}

	return manifest, nil
}

// Manifest represents a complete export of all resources.
type Manifest struct {
	SourceURL         string                    `json:"sourceUrl"`
	Spaces            []models.Space            `json:"spaces"`
	Stacks            []models.Stack            `json:"stacks"`
	Contexts          []models.Context          `json:"contexts"`
	Policies          []models.Policy           `json:"policies"`
	AWSIntegrations   []models.AWSIntegration   `json:"awsIntegrations"`
	AzureIntegrations []models.AzureIntegration `json:"azureIntegrations"`
}

// Summary returns a summary of the manifest contents.
func (m *Manifest) Summary() map[string]int {
	return map[string]int{
		"spaces":            len(m.Spaces),
		"stacks":            len(m.Stacks),
		"contexts":          len(m.Contexts),
		"policies":          len(m.Policies),
		"awsIntegrations":   len(m.AWSIntegrations),
		"azureIntegrations": len(m.AzureIntegrations),
	}
}

// SecretsCount returns the number of secrets that will need manual entry.
func (m *Manifest) SecretsCount() int {
	count := 0
	for _, ctx := range m.Contexts {
		for _, cfg := range ctx.Config {
			if cfg.WriteOnly {
				count++
			}
		}
	}
	return count
}
