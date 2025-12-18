package discovery

import (
	"context"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/models"
)

// DiscoverSpaces fetches all spaces from the Spacelift account.
func (s *Service) DiscoverSpaces(ctx context.Context) ([]models.Space, error) {
	var query client.SpacesQuery

	if err := s.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	spaces := make([]models.Space, 0, len(query.Spaces))
	for _, sp := range query.Spaces {
		space := models.Space{
			ID:              string(sp.ID),
			Name:            string(sp.Name),
			Description:     string(sp.Description),
			InheritEntities: bool(sp.InheritEntities),
			Labels:          toStringSlice(sp.Labels),
		}

		if sp.ParentSpace != nil {
			parent := string(*sp.ParentSpace)
			space.ParentSpace = &parent
		}

		spaces = append(spaces, space)
	}

	return spaces, nil
}

// DiscoverSpaceTree fetches spaces and returns them as a hierarchical tree.
func (s *Service) DiscoverSpaceTree(ctx context.Context) ([]*models.SpaceTree, error) {
	spaces, err := s.DiscoverSpaces(ctx)
	if err != nil {
		return nil, err
	}

	return models.BuildSpaceTree(spaces), nil
}

// Helper to convert GraphQL string slice to Go string slice.
func toStringSlice[T ~string](gs []T) []string {
	result := make([]string, len(gs))
	for i, g := range gs {
		result[i] = string(g)
	}
	return result
}
