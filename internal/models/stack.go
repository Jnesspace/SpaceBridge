package models

// Stack represents a Spacelift stack.
type Stack struct {
	ID                         string              `json:"id"`
	Name                       string              `json:"name"`
	Description                *string             `json:"description,omitempty"`
	Space                      string              `json:"space"`
	Branch                     string              `json:"branch"`
	Repository                 string              `json:"repository"`
	Namespace                  string              `json:"namespace"`
	ProjectRoot                *string             `json:"projectRoot,omitempty"`
	Provider                   string              `json:"provider"`   // VCS provider (GITHUB, GITLAB, etc.)
	VendorType                 string              `json:"vendorType"` // Stack type (StackConfigVendorTerraform, StackConfigVendorOpenTofu, etc.)
	RepositoryURL              *string             `json:"repositoryURL,omitempty"`
	RunnerImage                *string             `json:"runnerImage,omitempty"`
	TerraformVersion           *string             `json:"terraformVersion,omitempty"`
	TerragruntVersion          *string             `json:"terragruntVersion,omitempty"`
	WorkflowTool               *string             `json:"workflowTool,omitempty"` // TERRAFORM, OPEN_TOFU, TERRAGRUNT, CUSTOM
	Administrative             bool                `json:"administrative"`
	Autodeploy                 bool                `json:"autodeploy"`
	Autoretry                  bool                `json:"autoretry"`
	LocalPreviewEnabled        bool                `json:"localPreviewEnabled"`
	ProtectFromDeletion        bool                `json:"protectFromDeletion"`
	IsDisabled                 bool                `json:"isDisabled"`
	ManagesStateFile           bool                `json:"managesStateFile"`
	ExternalStateAccessEnabled bool                `json:"externalStateAccessEnabled"`
	Labels                     []string            `json:"labels"`
	AdditionalProjectGlobs     []string            `json:"additionalProjectGlobs"`
	Hooks                      Hooks               `json:"hooks"`
	AttachedContexts           []ContextAttachment `json:"attachedContexts,omitempty"`
	AttachedPolicies           []PolicyAttachment  `json:"attachedPolicies,omitempty"`
	DependsOn                  []StackDependency   `json:"dependsOn,omitempty"`
}

// IsTerraform returns true if the stack is a Terraform/OpenTofu/Terragrunt stack.
func (s *Stack) IsTerraform() bool {
	switch s.VendorType {
	case "StackConfigVendorTerraform", "StackConfigVendorOpenTofu", "StackConfigVendorTerragrunt":
		return true
	default:
		return false
	}
}

// IsOpenTofu returns true if the stack uses OpenTofu.
func (s *Stack) IsOpenTofu() bool {
	return s.WorkflowTool != nil && *s.WorkflowTool == "OPEN_TOFU"
}

// IsTerragrunt returns true if the stack uses Terragrunt.
func (s *Stack) IsTerragrunt() bool {
	return s.VendorType == "StackConfigVendorTerragrunt" ||
		(s.WorkflowTool != nil && *s.WorkflowTool == "TERRAGRUNT")
}

// Hooks represents the hooks configured on a stack or context.
type Hooks struct {
	AfterApply    []string `json:"afterApply"`
	BeforeApply   []string `json:"beforeApply"`
	AfterInit     []string `json:"afterInit"`
	BeforeInit    []string `json:"beforeInit"`
	AfterPlan     []string `json:"afterPlan"`
	BeforePlan    []string `json:"beforePlan"`
	AfterPerform  []string `json:"afterPerform"`
	BeforePerform []string `json:"beforePerform"`
	AfterDestroy  []string `json:"afterDestroy"`
	BeforeDestroy []string `json:"beforeDestroy"`
	AfterRun      []string `json:"afterRun"`
}

// ContextAttachment represents a context attached to a stack.
type ContextAttachment struct {
	ID        string `json:"id"`
	ContextID string `json:"contextId"`
	Priority  int    `json:"priority"`
}

// PolicyAttachment represents a policy attached to a stack.
type PolicyAttachment struct {
	ID       string `json:"id"`
	PolicyID string `json:"policyId"`
}

// StackDependency represents a dependency between stacks.
type StackDependency struct {
	ID               string `json:"id"`
	DependsOnStackID string `json:"dependsOnStackId"`
}
