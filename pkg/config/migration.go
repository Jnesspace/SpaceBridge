package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MigrationConfig holds the configuration for migration transformations.
type MigrationConfig struct {
	Destination DestinationConfig `yaml:"destination"`
}

// DestinationConfig holds destination-specific configuration.
type DestinationConfig struct {
	VCS VCSConfig `yaml:"vcs"`
}

// VCSConfig holds VCS integration configuration for the destination.
// Only one of these should be set - the one that matches your destination's VCS setup.
type VCSConfig struct {
	// GithubEnterprise configures a GitHub App integration
	GithubEnterprise *GithubEnterpriseConfig `yaml:"github_enterprise,omitempty"`

	// Gitlab configures a GitLab integration
	Gitlab *GitlabConfig `yaml:"gitlab,omitempty"`

	// BitbucketDatacenter configures a Bitbucket Data Center integration
	BitbucketDatacenter *BitbucketDatacenterConfig `yaml:"bitbucket_datacenter,omitempty"`

	// BitbucketCloud configures a Bitbucket Cloud integration
	BitbucketCloud *BitbucketCloudConfig `yaml:"bitbucket_cloud,omitempty"`

	// AzureDevops configures an Azure DevOps integration
	AzureDevops *AzureDevopsConfig `yaml:"azure_devops,omitempty"`
}

// GithubEnterpriseConfig configures a GitHub App integration.
type GithubEnterpriseConfig struct {
	ID        string `yaml:"id"`        // The ID of the GitHub App integration
	Namespace string `yaml:"namespace"` // The GitHub organization/user
}

// GitlabConfig configures a GitLab integration.
type GitlabConfig struct {
	ID        string `yaml:"id"`        // The ID of the GitLab integration
	Namespace string `yaml:"namespace"` // The GitLab group/user
}

// BitbucketDatacenterConfig configures a Bitbucket Data Center integration.
type BitbucketDatacenterConfig struct {
	ID        string `yaml:"id"`        // The ID of the Bitbucket DC integration
	Namespace string `yaml:"namespace"` // The Bitbucket project key
}

// BitbucketCloudConfig configures a Bitbucket Cloud integration.
type BitbucketCloudConfig struct {
	ID        string `yaml:"id"`        // The ID of the Bitbucket Cloud integration
	Namespace string `yaml:"namespace"` // The Bitbucket workspace
}

// AzureDevopsConfig configures an Azure DevOps integration.
type AzureDevopsConfig struct {
	ID      string `yaml:"id"`      // The ID of the Azure DevOps integration
	Project string `yaml:"project"` // The Azure DevOps project name
}

// HasVCSOverride returns true if any VCS override is configured.
func (v *VCSConfig) HasVCSOverride() bool {
	return v.GithubEnterprise != nil ||
		v.Gitlab != nil ||
		v.BitbucketDatacenter != nil ||
		v.BitbucketCloud != nil ||
		v.AzureDevops != nil
}

// LoadMigrationConfig loads the migration configuration from a YAML file.
func LoadMigrationConfig(path string) (*MigrationConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg MigrationConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid.
func (c *MigrationConfig) Validate() error {
	vcs := &c.Destination.VCS
	count := 0
	if vcs.GithubEnterprise != nil {
		count++
		if vcs.GithubEnterprise.ID == "" {
			return fmt.Errorf("github_enterprise.id is required")
		}
		if vcs.GithubEnterprise.Namespace == "" {
			return fmt.Errorf("github_enterprise.namespace is required")
		}
	}
	if vcs.Gitlab != nil {
		count++
		if vcs.Gitlab.ID == "" {
			return fmt.Errorf("gitlab.id is required")
		}
		if vcs.Gitlab.Namespace == "" {
			return fmt.Errorf("gitlab.namespace is required")
		}
	}
	if vcs.BitbucketDatacenter != nil {
		count++
		if vcs.BitbucketDatacenter.ID == "" {
			return fmt.Errorf("bitbucket_datacenter.id is required")
		}
		if vcs.BitbucketDatacenter.Namespace == "" {
			return fmt.Errorf("bitbucket_datacenter.namespace is required")
		}
	}
	if vcs.BitbucketCloud != nil {
		count++
		if vcs.BitbucketCloud.ID == "" {
			return fmt.Errorf("bitbucket_cloud.id is required")
		}
		if vcs.BitbucketCloud.Namespace == "" {
			return fmt.Errorf("bitbucket_cloud.namespace is required")
		}
	}
	if vcs.AzureDevops != nil {
		count++
		if vcs.AzureDevops.ID == "" {
			return fmt.Errorf("azure_devops.id is required")
		}
		if vcs.AzureDevops.Project == "" {
			return fmt.Errorf("azure_devops.project is required")
		}
	}

	if count > 1 {
		return fmt.Errorf("only one VCS integration type can be configured")
	}

	return nil
}
