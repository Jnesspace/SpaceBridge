package models

// AWSIntegration represents a Spacelift AWS cloud integration.
type AWSIntegration struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name"`
	RoleARN                     string   `json:"roleArn"`
	DurationSeconds             int      `json:"durationSeconds"`
	GenerateCredentialsInWorker bool     `json:"generateCredentialsInWorker"`
	ExternalID                  *string  `json:"externalId,omitempty"`
	Space                       string   `json:"space"`
	Labels                      []string `json:"labels"`
}

// AzureIntegration represents a Spacelift Azure cloud integration.
type AzureIntegration struct {
	ID                        string   `json:"id"`
	Name                      string   `json:"name"`
	TenantID                  string   `json:"tenantId"`
	DefaultSubscriptionID     *string  `json:"defaultSubscriptionId,omitempty"`
	ApplicationID             string   `json:"applicationId"`
	DisplayName               string   `json:"displayName"`
	Space                     string   `json:"space"`
	Labels                    []string `json:"labels"`
}
