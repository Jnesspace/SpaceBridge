package discovery

import (
	"context"

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
