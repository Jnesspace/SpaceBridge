package discovery

import (
	"context"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/models"
)

// DiscoverStacks fetches all stacks from the Spacelift account.
func (s *Service) DiscoverStacks(ctx context.Context) ([]models.Stack, error) {
	var query client.StacksQuery

	if err := s.client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	stacks := make([]models.Stack, 0, len(query.Stacks))
	for _, st := range query.Stacks {
		stack := models.Stack{
			ID:                         string(st.ID),
			Name:                       string(st.Name),
			Space:                      string(st.Space),
			Branch:                     string(st.Branch),
			Repository:                 string(st.Repository),
			Namespace:                  string(st.Namespace),
			Provider:                   string(st.Provider),
			VendorType:                 string(st.VendorConfig.Typename),
			Administrative:             bool(st.Administrative),
			Autodeploy:                 bool(st.Autodeploy),
			Autoretry:                  bool(st.Autoretry),
			LocalPreviewEnabled:        bool(st.LocalPreviewEnabled),
			ProtectFromDeletion:        bool(st.ProtectFromDeletion),
			IsDisabled:                 bool(st.IsDisabled),
			ManagesStateFile:           bool(st.ManagesStateFile),
			ExternalStateAccessEnabled: bool(st.VendorConfig.Terraform.ExternalStateAccessEnabled),
			Labels:                     toStringSlice(st.Labels),
			AdditionalProjectGlobs:     toStringSlice(st.AdditionalProjectGlobs),
			Hooks: models.Hooks{
				AfterApply:    toStringSlice(st.Hooks.AfterApply),
				BeforeApply:   toStringSlice(st.Hooks.BeforeApply),
				AfterInit:     toStringSlice(st.Hooks.AfterInit),
				BeforeInit:    toStringSlice(st.Hooks.BeforeInit),
				AfterPlan:     toStringSlice(st.Hooks.AfterPlan),
				BeforePlan:    toStringSlice(st.Hooks.BeforePlan),
				AfterPerform:  toStringSlice(st.Hooks.AfterPerform),
				BeforePerform: toStringSlice(st.Hooks.BeforePerform),
				AfterDestroy:  toStringSlice(st.Hooks.AfterDestroy),
				BeforeDestroy: toStringSlice(st.Hooks.BeforeDestroy),
				AfterRun:      toStringSlice(st.Hooks.AfterRun),
			},
		}

		// Optional fields
		if st.Description != nil {
			desc := string(*st.Description)
			stack.Description = &desc
		}
		if st.ProjectRoot != nil {
			pr := string(*st.ProjectRoot)
			stack.ProjectRoot = &pr
		}
		if st.RepositoryURL != nil {
			url := string(*st.RepositoryURL)
			stack.RepositoryURL = &url
		}
		if st.RunnerImage != nil {
			img := string(*st.RunnerImage)
			stack.RunnerImage = &img
		}
		if st.TerraformVersion != nil {
			tv := string(*st.TerraformVersion)
			stack.TerraformVersion = &tv
		}

		// Attached contexts
		for _, ac := range st.AttachedContexts {
			stack.AttachedContexts = append(stack.AttachedContexts, models.ContextAttachment{
				ID:        string(ac.ID),
				ContextID: string(ac.ContextID),
				Priority:  int(ac.Priority),
			})
		}

		// Attached policies
		for _, ap := range st.AttachedPolicies {
			stack.AttachedPolicies = append(stack.AttachedPolicies, models.PolicyAttachment{
				ID:       string(ap.ID),
				PolicyID: string(ap.PolicyID),
			})
		}

		// Stack dependencies
		for _, dep := range st.DependsOn {
			stack.DependsOn = append(stack.DependsOn, models.StackDependency{
				ID:               string(dep.ID),
				DependsOnStackID: string(dep.DependsOnStack.ID),
			})
		}

		stacks = append(stacks, stack)
	}

	return stacks, nil
}

// GetStacksBySpace returns stacks grouped by their space ID.
func GetStacksBySpace(stacks []models.Stack) map[string][]models.Stack {
	result := make(map[string][]models.Stack)
	for _, stack := range stacks {
		result[stack.Space] = append(result[stack.Space], stack)
	}
	return result
}
