package models

// Policy represents a Spacelift policy.
type Policy struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Space       string   `json:"space"`
	Type        string   `json:"type"`       // ACCESS, APPROVAL, GIT_PUSH, INITIALIZATION, LOGIN, PLAN, TASK, TRIGGER, NOTIFICATION
	EngineType  string   `json:"engineType"` // OPA or REGO
	Body        string   `json:"body"`       // The policy code
	Labels      []string `json:"labels"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
}

// PolicyType constants for the different policy types.
const (
	PolicyTypeAccess         = "ACCESS"
	PolicyTypeApproval       = "APPROVAL"
	PolicyTypeGitPush        = "GIT_PUSH"
	PolicyTypeInitialization = "INITIALIZATION"
	PolicyTypeLogin          = "LOGIN"
	PolicyTypePlan           = "PLAN"
	PolicyTypeTask           = "TASK"
	PolicyTypeTrigger        = "TRIGGER"
	PolicyTypeNotification   = "NOTIFICATION"
)

// PolicyEngineType constants.
const (
	PolicyEngineOPA  = "OPA"
	PolicyEngineRego = "REGO"
)

// GetPolicyTypeDescription returns a human-readable description of the policy type.
func GetPolicyTypeDescription(policyType string) string {
	descriptions := map[string]string{
		PolicyTypeAccess:         "Controls who can access stacks and what actions they can perform",
		PolicyTypeApproval:       "Determines whether runs require approval before applying",
		PolicyTypeGitPush:        "Filters which git push events trigger runs",
		PolicyTypeInitialization: "Runs during the initialization phase of a run",
		PolicyTypeLogin:          "Controls user login and session management",
		PolicyTypePlan:           "Validates Tofu plan output before approval",
		PolicyTypeTask:           "Controls which tasks can be executed",
		PolicyTypeTrigger:        "Determines when runs should be triggered",
		PolicyTypeNotification:   "Controls notification routing and filtering",
	}

	if desc, ok := descriptions[policyType]; ok {
		return desc
	}
	return "Unknown policy type"
}
