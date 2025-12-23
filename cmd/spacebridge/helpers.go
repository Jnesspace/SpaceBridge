package main

import (
	"fmt"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/discovery"
)

// createDiscoveryService creates a new discovery service with the source client.
func createDiscoveryService() (*discovery.Service, error) {
	if err := cfg.ValidateSource(); err != nil {
		return nil, fmt.Errorf("source configuration error: %w", err)
	}

	fmt.Printf("Connecting to: %s\n", cfg.Source.URL)
	if verbose {
		fmt.Printf("[CONFIG] API Key ID: %s\n", cfg.Source.KeyID)
	}

	c, err := client.New(cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return discovery.New(c), nil
}

// friendlyVendorType converts the GraphQL typename to a friendly name.
func friendlyVendorType(vendorType string) string {
	switch vendorType {
	case "StackConfigVendorTofu":
		return "Tofu"
	case "StackConfigVendorTerragrunt":
		return "Terragrunt"
	case "StackConfigVendorAnsible":
		return "Ansible"
	case "StackConfigVendorKubernetes":
		return "Kubernetes"
	case "StackConfigVendorCloudFormation":
		return "CloudFormation"
	case "StackConfigVendorPulumi":
		return "Pulumi"
	default:
		return vendorType
	}
}
