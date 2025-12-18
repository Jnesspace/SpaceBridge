package discovery

import (
	"context"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/models"
)

// DiscoverPolicies fetches all policies from the Spacelift account.
func (s *Service) DiscoverPolicies(ctx context.Context) ([]models.Policy, error) {
	var query client.PoliciesQuery

	if err := s.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	policies := make([]models.Policy, 0, len(query.Policies))
	for _, p := range query.Policies {
		policy := models.Policy{
			ID:        string(p.ID),
			Name:      string(p.Name),
			Space:     string(p.Space),
			Type:      string(p.Type),
			Body:      string(p.Body),
			Labels:    toStringSlice(p.Labels),
			CreatedAt: int64(p.CreatedAt),
			UpdatedAt: int64(p.UpdatedAt),
		}

		// Optional description
		if p.Description != nil {
			desc := string(*p.Description)
			policy.Description = &desc
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

// GetPoliciesBySpace returns policies grouped by their space ID.
func GetPoliciesBySpace(policies []models.Policy) map[string][]models.Policy {
	result := make(map[string][]models.Policy)
	for _, pol := range policies {
		result[pol.Space] = append(result[pol.Space], pol)
	}
	return result
}

// GetPoliciesByType returns policies grouped by their type.
func GetPoliciesByType(policies []models.Policy) map[string][]models.Policy {
	result := make(map[string][]models.Policy)
	for _, pol := range policies {
		result[pol.Type] = append(result[pol.Type], pol)
	}
	return result
}
