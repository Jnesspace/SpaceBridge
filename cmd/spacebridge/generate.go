package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/discovery"
	"github.com/jnesspace/spacebridge/internal/generator"
	"github.com/jnesspace/spacebridge/internal/models"
	"github.com/jnesspace/spacebridge/pkg/config"
)

var (
	generateDir     string
	manifestInput   string
	disableStacks   bool
	filterSpace     string
	migrationConfig string
)

// newGenerateCmd creates the generate command.
func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Tofu code from discovered resources",
		Long: `Generate Tofu code using the Spacelift provider from a manifest file
or directly from the source Spacelift account.

The generated code includes:
  - main.tf:      All Spacelift resources (spaces, stacks, contexts, policies)
  - variables.tf: Variable declarations for secrets
  - secrets.auto.tfvars.template: Template for secret values
  - provider.tf:  Spacelift provider configuration

Example usage:
  # Generate from live discovery (stacks disabled for safe migration)
  spacebridge generate -o ./tofu/ --disabled

  # Generate from existing manifest
  spacebridge generate -m manifest.json -o ./tofu/

  # Generate with VCS override (e.g., use GitHub App instead of built-in)
  spacebridge generate -o ./tofu/ -c spacebridge.yaml`,
		RunE: runGenerate,
	}
	cmd.Flags().StringVarP(&generateDir, "output", "o", "./generated", "Output directory for Tofu files")
	cmd.Flags().StringVarP(&manifestInput, "manifest", "m", "", "Input manifest file (optional, discovers fresh if not provided)")
	cmd.Flags().BoolVarP(&disableStacks, "disabled", "d", false, "Create stacks as disabled for safe state migration")
	cmd.Flags().StringVarP(&filterSpace, "space", "s", "", "Only include resources from this space (and its children)")
	cmd.Flags().StringVarP(&migrationConfig, "config", "c", "", "Migration config YAML file for VCS overrides")
	return cmd
}

// runGenerate generates Tofu code from a manifest.
func runGenerate(cmd *cobra.Command, args []string) error {
	var manifest *discovery.Manifest

	if manifestInput != "" {
		// Load from file
		fmt.Printf("Loading manifest from: %s\n", manifestInput)
		data, err := os.ReadFile(manifestInput)
		if err != nil {
			return fmt.Errorf("failed to read manifest file: %w", err)
		}

		manifest = &discovery.Manifest{}
		if err := json.Unmarshal(data, manifest); err != nil {
			return fmt.Errorf("failed to parse manifest file: %w", err)
		}
	} else {
		// Discover fresh
		svc, err := createDiscoveryService()
		if err != nil {
			return err
		}

		ctx := context.Background()
		fmt.Println("Discovering resources for code generation...")

		manifest, err = svc.DiscoverAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to discover resources: %w", err)
		}
	}

	// Apply space filter if specified
	if filterSpace != "" {
		fmt.Printf("Filtering to space: %s (and children)\n", filterSpace)
		manifest = filterManifestBySpace(manifest, filterSpace)
		if len(manifest.Stacks) == 0 && len(manifest.Contexts) == 0 && len(manifest.Policies) == 0 {
			return fmt.Errorf("no resources found in space '%s'", filterSpace)
		}
	}

	// Count secrets for summary
	secretCount := 0
	for _, ctx := range manifest.Contexts {
		for _, cfg := range ctx.Config {
			if cfg.WriteOnly {
				secretCount++
			}
		}
	}

	// Generate Tofu code
	fmt.Printf("\nGenerating Tofu code to: %s\n", generateDir)
	gen := generator.New(manifest, generateDir).WithSafeMode(disableStacks)

	// Use destination config if available for provider.tf
	if cfg.HasDestination() {
		gen.WithDestinationConfig(&cfg.Destination)
	}

	// Load migration config if provided
	if migrationConfig != "" {
		fmt.Printf("Loading migration config from: %s\n", migrationConfig)
		migCfg, err := config.LoadMigrationConfig(migrationConfig)
		if err != nil {
			return fmt.Errorf("failed to load migration config: %w", err)
		}
		if err := migCfg.Validate(); err != nil {
			return fmt.Errorf("invalid migration config: %w", err)
		}
		gen.WithMigrationConfig(migCfg)
		if migCfg.Destination.VCS.HasVCSOverride() {
			fmt.Println("VCS override configured - stacks will use custom VCS integration")
		}
	}

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate Tofu code: %w", err)
	}

	// Count stacks with managed state, autodeploy, and external state access
	managedStateCount := 0
	autodeployCount := 0
	needsAccessCount := 0
	var needsAccessStacks []string
	for _, stack := range manifest.Stacks {
		if stack.ManagesStateFile && stack.IsTerraform() {
			managedStateCount++
			if !stack.ExternalStateAccessEnabled {
				needsAccessCount++
				needsAccessStacks = append(needsAccessStacks, stack.Name)
			}
		}
		if stack.Autodeploy {
			autodeployCount++
		}
	}

	// Print summary
	fmt.Println("\n✓ Tofu code generated successfully!")
	fmt.Println("\nGenerated resources:")
	// Count non-root spaces (root is not generated as a resource)
	nonRootSpaces := 0
	for _, space := range manifest.Spaces {
		if space.ID != "root" {
			nonRootSpaces++
		}
	}
	fmt.Printf("  - Spaces:             %d\n", nonRootSpaces)
	fmt.Printf("  - Contexts:           %d\n", len(manifest.Contexts))
	fmt.Printf("  - Policies:           %d\n", len(manifest.Policies))
	fmt.Printf("  - Stacks:             %d\n", len(manifest.Stacks))
	fmt.Printf("  - AWS Integrations:   %d\n", len(manifest.AWSIntegrations))
	fmt.Printf("  - Azure Integrations: %d\n", len(manifest.AzureIntegrations))

	if disableStacks {
		fmt.Println("\n🔒 Safe migration mode enabled:")
		fmt.Printf("   - All stacks created with autodeploy = false\n")
		if autodeployCount > 0 {
			fmt.Printf("   - %d stacks need autodeploy re-enabled after migration\n", autodeployCount)
			fmt.Printf("   - See autodeploy_re_enable.tf.disabled\n")
		}
		fmt.Printf("   - %d stacks with Spacelift-managed state can be migrated\n", managedStateCount)

		if needsAccessCount > 0 {
			fmt.Printf("\n⚠️  %d stacks need external state access enabled before migration:\n", needsAccessCount)
			for _, name := range needsAccessStacks {
				fmt.Printf("   - %s\n", name)
			}
			fmt.Println("   Run: spacebridge state enable-access")
		} else if managedStateCount > 0 {
			fmt.Println("\n✓ All managed-state stacks already have external state access enabled")
		}
	}

	if secretCount > 0 {
		fmt.Printf("\n⚠️  %d secret values require manual entry.\n", secretCount)
		fmt.Println("   Edit secrets.auto.tfvars.template and rename to secrets.auto.tfvars")
	}

	fmt.Println("\nNext steps:")
	fmt.Printf("  1. cd %s\n", generateDir)
	fmt.Println("  2. Review and modify generated code as needed")
	if secretCount > 0 {
		fmt.Println("  3. Fill in secret values in secrets.auto.tfvars")
		fmt.Println("  4. Tofu init && Tofu plan && Tofu apply")
	} else {
		fmt.Println("  3. Tofu init && Tofu plan && Tofu apply")
	}
	if disableStacks && managedStateCount > 0 {
		step := 5
		if secretCount > 0 {
			step = 5
		} else {
			step = 4
		}
		fmt.Printf("  %d. spacebridge state enable-access  # Enable external state access on source\n", step)
		step++
		fmt.Printf("  %d. spacebridge state plan           # Preview state migration\n", step)
		step++
		fmt.Printf("  %d. spacebridge state migrate        # Migrate state to new stacks\n", step)
		step++
		if autodeployCount > 0 {
			fmt.Printf("  %d. Rename autodeploy_re_enable.tf.disabled → .tf\n", step)
			step++
			fmt.Printf("  %d. tofu apply                       # Re-enable autodeploy\n", step)
		}
	}

	return nil
}

// filterManifestBySpace filters a manifest to only include resources in the given space,
// its ancestors (so the hierarchy can be created), and its descendants.
func filterManifestBySpace(manifest *discovery.Manifest, spaceID string) *discovery.Manifest {
	if spaceID == "" {
		return manifest
	}

	// Build a map of space ID -> space for quick lookups
	spaceMap := make(map[string]models.Space)
	for _, space := range manifest.Spaces {
		spaceMap[space.ID] = space
	}

	// Build a set of spaces to include
	includedSpaces := make(map[string]bool)
	includedSpaces[spaceID] = true

	// Track ancestor spaces separately (for inheriting integrations from upstream)
	ancestorSpaces := make(map[string]bool)
	ancestorSpaces["root"] = true // Root is always an ancestor

	// Include ancestor spaces (walk UP the tree) so the hierarchy can be created
	currentSpaceID := spaceID
	for currentSpaceID != "" && currentSpaceID != "root" {
		if space, exists := spaceMap[currentSpaceID]; exists {
			includedSpaces[currentSpaceID] = true
			ancestorSpaces[currentSpaceID] = true
			if space.ParentSpace != nil {
				currentSpaceID = *space.ParentSpace
				ancestorSpaces[currentSpaceID] = true
			} else {
				break
			}
		} else {
			break
		}
	}

	// Find all child spaces recursively (walk DOWN the tree)
	changed := true
	for changed {
		changed = false
		for _, space := range manifest.Spaces {
			if space.ParentSpace != nil && includedSpaces[*space.ParentSpace] && !includedSpaces[space.ID] {
				includedSpaces[space.ID] = true
				changed = true
			}
		}
	}

	// Filter stacks and collect attached context/policy IDs
	var filteredStacks []models.Stack
	requiredContextIDs := make(map[string]bool)
	requiredPolicyIDs := make(map[string]bool)
	for _, stack := range manifest.Stacks {
		if includedSpaces[stack.Space] {
			filteredStacks = append(filteredStacks, stack)
			// Collect context IDs from attachments
			for _, attachment := range stack.AttachedContexts {
				requiredContextIDs[attachment.ContextID] = true
			}
			// Collect policy IDs from attachments
			for _, attachment := range stack.AttachedPolicies {
				requiredPolicyIDs[attachment.PolicyID] = true
			}
		}
	}

	// Build a map of context/policy ID -> space for lookups
	contextSpaceMap := make(map[string]string)
	for _, ctx := range manifest.Contexts {
		contextSpaceMap[ctx.ID] = ctx.Space
	}
	policySpaceMap := make(map[string]string)
	for _, policy := range manifest.Policies {
		policySpaceMap[policy.ID] = policy.Space
	}

	// Include spaces for required contexts/policies and their ancestors
	for contextID := range requiredContextIDs {
		if spaceID, exists := contextSpaceMap[contextID]; exists {
			// Walk up the tree to include all ancestors
			currentSpaceID := spaceID
			for currentSpaceID != "" && currentSpaceID != "root" {
				if space, exists := spaceMap[currentSpaceID]; exists {
					includedSpaces[currentSpaceID] = true
					if space.ParentSpace != nil {
						currentSpaceID = *space.ParentSpace
					} else {
						break
					}
				} else {
					break
				}
			}
		}
	}
	for policyID := range requiredPolicyIDs {
		if spaceID, exists := policySpaceMap[policyID]; exists {
			// Walk up the tree to include all ancestors
			currentSpaceID := spaceID
			for currentSpaceID != "" && currentSpaceID != "root" {
				if space, exists := spaceMap[currentSpaceID]; exists {
					includedSpaces[currentSpaceID] = true
					if space.ParentSpace != nil {
						currentSpaceID = *space.ParentSpace
					} else {
						break
					}
				} else {
					break
				}
			}
		}
	}

	// Filter spaces (now includes spaces needed for attached contexts/policies)
	var filteredSpaces []models.Space
	for _, space := range manifest.Spaces {
		if includedSpaces[space.ID] {
			filteredSpaces = append(filteredSpaces, space)
		}
	}

	// Filter contexts: include if in filtered space OR attached to a filtered stack
	var filteredContexts []models.Context
	for _, ctx := range manifest.Contexts {
		if includedSpaces[ctx.Space] || requiredContextIDs[ctx.ID] {
			filteredContexts = append(filteredContexts, ctx)
		}
	}

	// Filter policies: include if in filtered space OR attached to a filtered stack
	var filteredPolicies []models.Policy
	for _, policy := range manifest.Policies {
		if includedSpaces[policy.Space] || requiredPolicyIDs[policy.ID] {
			filteredPolicies = append(filteredPolicies, policy)
		}
	}

	// Filter AWS integrations: include if in filtered space OR in an ancestor space (inherited)
	var filteredAWSIntegrations []models.AWSIntegration
	for _, integration := range manifest.AWSIntegrations {
		if includedSpaces[integration.Space] || ancestorSpaces[integration.Space] {
			filteredAWSIntegrations = append(filteredAWSIntegrations, integration)
		}
	}

	// Filter Azure integrations: include if in filtered space OR in an ancestor space (inherited)
	var filteredAzureIntegrations []models.AzureIntegration
	for _, integration := range manifest.AzureIntegrations {
		if includedSpaces[integration.Space] || ancestorSpaces[integration.Space] {
			filteredAzureIntegrations = append(filteredAzureIntegrations, integration)
		}
	}

	return &discovery.Manifest{
		SourceURL:         manifest.SourceURL,
		Spaces:            filteredSpaces,
		Stacks:            filteredStacks,
		Contexts:          filteredContexts,
		Policies:          filteredPolicies,
		AWSIntegrations:   filteredAWSIntegrations,
		AzureIntegrations: filteredAzureIntegrations,
	}
}
