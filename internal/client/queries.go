package client

import graphql "github.com/hasura/go-graphql-client"

// SpacesQuery is the GraphQL query for fetching all spaces.
type SpacesQuery struct {
	Spaces []struct {
		ID              graphql.ID       `graphql:"id"`
		Name            graphql.String   `graphql:"name"`
		Description     graphql.String   `graphql:"description"`
		ParentSpace     *graphql.ID      `graphql:"parentSpace"`
		InheritEntities graphql.Boolean  `graphql:"inheritEntities"`
		Labels          []graphql.String `graphql:"labels"`
	} `graphql:"spaces"`
}

// StacksQuery is the GraphQL query for fetching all stacks.
type StacksQuery struct {
	Stacks []struct {
		ID                     graphql.ID       `graphql:"id"`
		Name                   graphql.String   `graphql:"name"`
		Description            *graphql.String  `graphql:"description"`
		Space                  graphql.ID       `graphql:"space"`
		Branch                 graphql.String   `graphql:"branch"`
		Repository             graphql.String   `graphql:"repository"`
		Namespace              graphql.String   `graphql:"namespace"`
		ProjectRoot            *graphql.String  `graphql:"projectRoot"`
		Provider               graphql.String   `graphql:"provider"`
		RepositoryURL          *graphql.String  `graphql:"repositoryURL"`
		RunnerImage            *graphql.String  `graphql:"runnerImage"`
		TerraformVersion       *graphql.String  `graphql:"terraformVersion"`
		Administrative         graphql.Boolean  `graphql:"administrative"`
		Autodeploy             graphql.Boolean  `graphql:"autodeploy"`
		Autoretry              graphql.Boolean  `graphql:"autoretry"`
		LocalPreviewEnabled    graphql.Boolean  `graphql:"localPreviewEnabled"`
		ProtectFromDeletion    graphql.Boolean  `graphql:"protectFromDeletion"`
		IsDisabled             graphql.Boolean  `graphql:"isDisabled"`
		ManagesStateFile       graphql.Boolean  `graphql:"managesStateFile"`
		Labels                 []graphql.String `graphql:"labels"`
		AdditionalProjectGlobs []graphql.String `graphql:"additionalProjectGlobs"`
		VendorConfig           struct {
			Typename  graphql.String `graphql:"__typename"`
			Terraform struct {
				ExternalStateAccessEnabled graphql.Boolean `graphql:"externalStateAccessEnabled"`
			} `graphql:"... on StackConfigVendorTerraform"`
		} `graphql:"vendorConfig"`
		Hooks struct {
			AfterApply    []graphql.String `graphql:"afterApply"`
			BeforeApply   []graphql.String `graphql:"beforeApply"`
			AfterInit     []graphql.String `graphql:"afterInit"`
			BeforeInit    []graphql.String `graphql:"beforeInit"`
			AfterPlan     []graphql.String `graphql:"afterPlan"`
			BeforePlan    []graphql.String `graphql:"beforePlan"`
			AfterPerform  []graphql.String `graphql:"afterPerform"`
			BeforePerform []graphql.String `graphql:"beforePerform"`
			AfterDestroy  []graphql.String `graphql:"afterDestroy"`
			BeforeDestroy []graphql.String `graphql:"beforeDestroy"`
			AfterRun      []graphql.String `graphql:"afterRun"`
		} `graphql:"hooks"`
		AttachedContexts []struct {
			ID        graphql.ID  `graphql:"id"`
			ContextID graphql.ID  `graphql:"contextId"`
			Priority  graphql.Int `graphql:"priority"`
		} `graphql:"attachedContexts"`
		AttachedPolicies []struct {
			ID       graphql.ID `graphql:"id"`
			PolicyID graphql.ID `graphql:"policyId"`
		} `graphql:"attachedPolicies"`
		DependsOn []struct {
			ID             graphql.ID `graphql:"id"`
			DependsOnStack struct {
				ID graphql.ID `graphql:"id"`
			} `graphql:"dependsOnStack"`
		} `graphql:"dependsOn"`
	} `graphql:"stacks"`
}

// ContextsQuery is the GraphQL query for fetching all contexts.
type ContextsQuery struct {
	Contexts []struct {
		ID          graphql.ID       `graphql:"id"`
		Name        graphql.String   `graphql:"name"`
		Description *graphql.String  `graphql:"description"`
		Space       graphql.ID       `graphql:"space"`
		Labels      []graphql.String `graphql:"labels"`
		CreatedAt   graphql.Int      `graphql:"createdAt"`
		UpdatedAt   graphql.Int      `graphql:"updatedAt"`
		Hooks       struct {
			AfterApply    []graphql.String `graphql:"afterApply"`
			BeforeApply   []graphql.String `graphql:"beforeApply"`
			AfterInit     []graphql.String `graphql:"afterInit"`
			BeforeInit    []graphql.String `graphql:"beforeInit"`
			AfterPlan     []graphql.String `graphql:"afterPlan"`
			BeforePlan    []graphql.String `graphql:"beforePlan"`
			AfterPerform  []graphql.String `graphql:"afterPerform"`
			BeforePerform []graphql.String `graphql:"beforePerform"`
			AfterDestroy  []graphql.String `graphql:"afterDestroy"`
			BeforeDestroy []graphql.String `graphql:"beforeDestroy"`
			AfterRun      []graphql.String `graphql:"afterRun"`
		} `graphql:"hooks"`
		Config []struct {
			ID        graphql.ID      `graphql:"id"`
			Type      graphql.String  `graphql:"type"`
			Value     graphql.String  `graphql:"value"`
			WriteOnly graphql.Boolean `graphql:"writeOnly"`
		} `graphql:"config"`
	} `graphql:"contexts"`
}

// PoliciesQuery is the GraphQL query for fetching all policies.
type PoliciesQuery struct {
	Policies []struct {
		ID          graphql.ID       `graphql:"id"`
		Name        graphql.String   `graphql:"name"`
		Description *graphql.String  `graphql:"description"`
		Space       graphql.ID       `graphql:"space"`
		Type        graphql.String   `graphql:"type"`
		Body        graphql.String   `graphql:"body"`
		Labels      []graphql.String `graphql:"labels"`
		CreatedAt   graphql.Int      `graphql:"createdAt"`
		UpdatedAt   graphql.Int      `graphql:"updatedAt"`
	} `graphql:"policies"`
}

// WorkerPoolsQuery is the GraphQL query for fetching all worker pools.
type WorkerPoolsQuery struct {
	WorkerPools []struct {
		ID          graphql.ID       `graphql:"id"`
		Name        graphql.String   `graphql:"name"`
		Description *graphql.String  `graphql:"description"`
		Space       graphql.ID       `graphql:"space"`
		Labels      []graphql.String `graphql:"labels"`
		CreatedAt   graphql.Int      `graphql:"createdAt"`
	} `graphql:"workerPools"`
}

// StackUpdateInput is the input for updating a stack.
type StackUpdateInput struct {
	ExternalStateAccessEnabled *graphql.Boolean `json:"vendorConfig,omitempty"`
}

// StackUpdateMutation is the mutation for updating a stack.
type StackUpdateMutation struct {
	StackUpdate struct {
		ID graphql.ID `graphql:"id"`
	} `graphql:"stackUpdate(id: $id, input: $input)"`
}

// StateDownloadURLMutation gets a pre-signed URL to download stack state.
type StateDownloadURLMutation struct {
	StateDownloadURL struct {
		URL graphql.String `graphql:"url"`
	} `graphql:"stateDownloadUrl(stack: $stack)"`
}

// StateUploadURLMutation gets a pre-signed URL to upload stack state.
type StateUploadURLMutation struct {
	StateUploadURL struct {
		URL graphql.String `graphql:"url"`
	} `graphql:"stateUploadUrl(stack: $stack)"`
}

// StackManagedStateImportMutation imports state into a stack.
type StackManagedStateImportMutation struct {
	StackManagedStateImport graphql.Boolean `graphql:"stackManagedStateImport(id: $id)"`
}
