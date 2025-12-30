package discovery

import (
	"context"

	graphql "github.com/hasura/go-graphql-client"
	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/models"
)

// DiscoverAWSIntegrations fetches all AWS integrations from the Spacelift account.
func (s *Service) DiscoverAWSIntegrations(ctx context.Context) ([]models.AWSIntegration, error) {
	var query client.AWSIntegrationsQuery

	if err := s.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	integrations := make([]models.AWSIntegration, 0, len(query.AWSIntegrations))
	for _, i := range query.AWSIntegrations {
		integration := models.AWSIntegration{
			ID:                          string(i.ID),
			Name:                        string(i.Name),
			RoleARN:                     string(i.RoleARN),
			DurationSeconds:             int(i.DurationSeconds),
			GenerateCredentialsInWorker: bool(i.GenerateCredentialsInWorker),
			Space:                       string(i.Space),
			Labels:                      toStringSlice(i.Labels),
		}

		if i.ExternalID != nil {
			extID := string(*i.ExternalID)
			integration.ExternalID = &extID
		}

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

// DiscoverAzureIntegrations fetches all Azure integrations from the Spacelift account.
func (s *Service) DiscoverAzureIntegrations(ctx context.Context) ([]models.AzureIntegration, error) {
	var query client.AzureIntegrationsQuery

	if err := s.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	integrations := make([]models.AzureIntegration, 0, len(query.AzureIntegrations))
	for _, i := range query.AzureIntegrations {
		integration := models.AzureIntegration{
			ID:            string(i.ID),
			Name:          string(i.Name),
			TenantID:      string(i.TenantID),
			ApplicationID: string(i.ApplicationID),
			DisplayName:   string(i.DisplayName),
			Space:         string(i.Space),
			Labels:        toStringSlice(i.Labels),
		}

		if i.DefaultSubscriptionID != nil {
			subID := string(*i.DefaultSubscriptionID)
			integration.DefaultSubscriptionID = &subID
		}

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

// DiscoverAWSIntegrationAttachments fetches all stack attachments for an AWS integration.
// Returns a map of stackID -> attachment info.
func (s *Service) DiscoverAWSIntegrationAttachments(ctx context.Context, integrationID string) (map[string]models.AWSIntegrationAttachment, error) {
	var query client.AWSIntegrationAttachmentsQuery
	vars := map[string]interface{}{
		"id": graphql.ID(integrationID),
	}

	if err := s.client.Query(ctx, &query, vars); err != nil {
		return nil, err
	}

	attachments := make(map[string]models.AWSIntegrationAttachment)
	if query.AWSIntegration == nil {
		return attachments, nil
	}

	for _, a := range query.AWSIntegration.AttachedStacks {
		// Skip module attachments - we only care about stacks
		if bool(a.IsModule) {
			continue
		}
		stackID := string(a.StackID)
		attachments[stackID] = models.AWSIntegrationAttachment{
			IntegrationID: integrationID,
			Read:          bool(a.Read),
			Write:         bool(a.Write),
		}
	}

	return attachments, nil
}

// DiscoverAzureIntegrationAttachments fetches all stack attachments for an Azure integration.
// Returns a map of stackID -> attachment info.
func (s *Service) DiscoverAzureIntegrationAttachments(ctx context.Context, integrationID string) (map[string]models.AzureIntegrationAttachment, error) {
	var query client.AzureIntegrationAttachmentsQuery
	vars := map[string]interface{}{
		"id": graphql.ID(integrationID),
	}

	if err := s.client.Query(ctx, &query, vars); err != nil {
		return nil, err
	}

	attachments := make(map[string]models.AzureIntegrationAttachment)
	if query.AzureIntegration == nil {
		return attachments, nil
	}

	for _, a := range query.AzureIntegration.AttachedStacks {
		// Skip module attachments - we only care about stacks
		if bool(a.IsModule) {
			continue
		}
		stackID := string(a.StackID)
		attachment := models.AzureIntegrationAttachment{
			IntegrationID: integrationID,
			Read:          bool(a.Read),
			Write:         bool(a.Write),
		}
		if a.SubscriptionID != nil {
			subID := string(*a.SubscriptionID)
			attachment.SubscriptionID = &subID
		}
		attachments[stackID] = attachment
	}

	return attachments, nil
}
