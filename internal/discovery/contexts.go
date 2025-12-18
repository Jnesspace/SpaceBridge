package discovery

import (
	"context"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/models"
)

// DiscoverContexts fetches all contexts from the Spacelift account.
func (s *Service) DiscoverContexts(ctx context.Context) ([]models.Context, error) {
	var query client.ContextsQuery

	if err := s.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	contexts := make([]models.Context, 0, len(query.Contexts))
	for _, c := range query.Contexts {
		context := models.Context{
			ID:        string(c.ID),
			Name:      string(c.Name),
			Space:     string(c.Space),
			Labels:    toStringSlice(c.Labels),
			CreatedAt: int64(c.CreatedAt),
			UpdatedAt: int64(c.UpdatedAt),
			Hooks: models.Hooks{
				AfterApply:    toStringSlice(c.Hooks.AfterApply),
				BeforeApply:   toStringSlice(c.Hooks.BeforeApply),
				AfterInit:     toStringSlice(c.Hooks.AfterInit),
				BeforeInit:    toStringSlice(c.Hooks.BeforeInit),
				AfterPlan:     toStringSlice(c.Hooks.AfterPlan),
				BeforePlan:    toStringSlice(c.Hooks.BeforePlan),
				AfterPerform:  toStringSlice(c.Hooks.AfterPerform),
				BeforePerform: toStringSlice(c.Hooks.BeforePerform),
				AfterDestroy:  toStringSlice(c.Hooks.AfterDestroy),
				BeforeDestroy: toStringSlice(c.Hooks.BeforeDestroy),
				AfterRun:      toStringSlice(c.Hooks.AfterRun),
			},
		}

		// Optional description
		if c.Description != nil {
			desc := string(*c.Description)
			context.Description = &desc
		}

		// Config elements
		for _, cfg := range c.Config {
			context.Config = append(context.Config, models.ConfigElement{
				ID:        string(cfg.ID),
				Type:      string(cfg.Type),
				Value:     string(cfg.Value),
				WriteOnly: bool(cfg.WriteOnly),
			})
		}

		contexts = append(contexts, context)
	}

	return contexts, nil
}

// GetContextsBySpace returns contexts grouped by their space ID.
func GetContextsBySpace(contexts []models.Context) map[string][]models.Context {
	result := make(map[string][]models.Context)
	for _, ctx := range contexts {
		result[ctx.Space] = append(result[ctx.Space], ctx)
	}
	return result
}

// GetContextsWithSecrets returns only contexts that have secret config elements.
func GetContextsWithSecrets(contexts []models.Context) []models.Context {
	var result []models.Context
	for _, ctx := range contexts {
		if ctx.HasSecrets() {
			result = append(result, ctx)
		}
	}
	return result
}
