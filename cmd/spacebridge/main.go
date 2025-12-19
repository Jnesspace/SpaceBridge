// SpaceBridge - Spacelift Enterprise Migration Kit
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/discovery"
	"github.com/jnesspace/spacebridge/internal/generator"
	"github.com/jnesspace/spacebridge/internal/models"
	"github.com/jnesspace/spacebridge/internal/ui"
	"github.com/jnesspace/spacebridge/pkg/config"
)

var (
	outputFile    string
	verbose       bool
	cfg           *config.Config
	generateDir   string
	manifestInput string
	disableStacks bool
	filterSpace   string
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	// Load configuration
	var err error
	cfg, err = config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "spacebridge",
		Short: "SpaceBridge - Spacelift Enterprise Migration Kit",
		Long: `SpaceBridge is a professional-grade CLI tool for migrating and
cloning Spacelift resources between accounts.

It provides safe, validated migrations with full dry-run support.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			client.Verbose = verbose
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Discover command group
	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover resources in the source Spacelift account",
	}

	discoverCmd.AddCommand(
		newDiscoverSpacesCmd(),
		newDiscoverStacksCmd(),
		newDiscoverContextsCmd(),
		newDiscoverPoliciesCmd(),
		newDiscoverAllCmd(),
	)

	// Export command
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export all resources to a manifest file",
		RunE:  runExport,
	}
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "manifest.json", "Output file path")

	// Generate command
	generateCmd := &cobra.Command{
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
  spacebridge generate -o ./Tofu/ --disabled

  # Generate from existing manifest
  spacebridge generate -m manifest.json -o ./Tofu/`,
		RunE: runGenerate,
	}
	generateCmd.Flags().StringVarP(&generateDir, "output", "o", "./generated", "Output directory for Tofu files")
	generateCmd.Flags().StringVarP(&manifestInput, "manifest", "m", "", "Input manifest file (optional, discovers fresh if not provided)")
	generateCmd.Flags().BoolVarP(&disableStacks, "disabled", "d", false, "Create stacks as disabled for safe state migration")
	generateCmd.Flags().StringVarP(&filterSpace, "space", "s", "", "Only include resources from this space (and its children)")

	// State command group
	stateCmd := &cobra.Command{
		Use:   "state",
		Short: "Manage Tofu state migration",
	}
	stateCmd.AddCommand(
		newStatePlanCmd(),
		newStateEnableAccessCmd(),
		newStateMigrateCmd(),
	)

	// Stacks command group
	stacksCmd := &cobra.Command{
		Use:   "stacks",
		Short: "Manage stacks in destination account",
	}
	stacksCmd.AddCommand(
		newStacksEnableCmd(),
	)

	rootCmd.AddCommand(discoverCmd, exportCmd, generateCmd, stateCmd, stacksCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// newDiscoverSpacesCmd creates the discover spaces command.
func newDiscoverSpacesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spaces",
		Short: "Discover all spaces with hierarchy",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := createDiscoveryService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			spaces, err := svc.DiscoverSpaces(ctx)
			if err != nil {
				return fmt.Errorf("failed to discover spaces: %w", err)
			}

			ui.PrintSpaces(spaces)
			fmt.Printf("\nTotal: %d spaces\n", len(spaces))
			return nil
		},
	}
}

// newDiscoverStacksCmd creates the discover stacks command.
func newDiscoverStacksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stacks",
		Short: "Discover all stacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := createDiscoveryService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			stacks, err := svc.DiscoverStacks(ctx)
			if err != nil {
				return fmt.Errorf("failed to discover stacks: %w", err)
			}

			ui.PrintStacks(stacks)
			return nil
		},
	}
}

// newDiscoverContextsCmd creates the discover contexts command.
func newDiscoverContextsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "contexts",
		Short: "Discover all contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := createDiscoveryService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			contexts, err := svc.DiscoverContexts(ctx)
			if err != nil {
				return fmt.Errorf("failed to discover contexts: %w", err)
			}

			ui.PrintContexts(contexts)
			ui.PrintSecretsWarning(contexts)
			return nil
		},
	}
}

// newDiscoverPoliciesCmd creates the discover policies command.
func newDiscoverPoliciesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "policies",
		Short: "Discover all policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := createDiscoveryService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			policies, err := svc.DiscoverPolicies(ctx)
			if err != nil {
				return fmt.Errorf("failed to discover policies: %w", err)
			}

			ui.PrintPolicies(policies)
			return nil
		},
	}
}

// newDiscoverAllCmd creates the discover all command.
func newDiscoverAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Discover all resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := createDiscoveryService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			fmt.Println("Discovering all resources...")

			manifest, err := svc.DiscoverAll(ctx)
			if err != nil {
				return fmt.Errorf("failed to discover resources: %w", err)
			}

			ui.PrintSpaces(manifest.Spaces)
			ui.PrintStacks(manifest.Stacks)
			ui.PrintContexts(manifest.Contexts)
			ui.PrintPolicies(manifest.Policies)
			ui.PrintSecretsWarning(manifest.Contexts)
			ui.PrintSummary(manifest)

			return nil
		},
	}
}

// runExport exports all resources to a JSON file.
func runExport(cmd *cobra.Command, args []string) error {
	svc, err := createDiscoveryService()
	if err != nil {
		return err
	}

	ctx := context.Background()
	fmt.Println("Discovering all resources for export...")

	manifest, err := svc.DiscoverAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	fmt.Printf("Manifest exported to: %s\n", outputFile)
	ui.PrintSummary(manifest)

	return nil
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
	fmt.Printf("  - Spaces:    %d\n", len(manifest.Spaces)-1) // -1 for root
	fmt.Printf("  - Contexts:  %d\n", len(manifest.Contexts))
	fmt.Printf("  - Policies:  %d\n", len(manifest.Policies))
	fmt.Printf("  - Stacks:    %d\n", len(manifest.Stacks))

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

// filterManifestBySpace filters a manifest to only include resources in the given space and its children.
func filterManifestBySpace(manifest *discovery.Manifest, spaceID string) *discovery.Manifest {
	if spaceID == "" {
		return manifest
	}

	// Build a set of spaces to include (the target space and all its descendants)
	includedSpaces := make(map[string]bool)
	includedSpaces[spaceID] = true

	// Find all child spaces recursively
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

	// Filter spaces
	var filteredSpaces []models.Space
	for _, space := range manifest.Spaces {
		if includedSpaces[space.ID] {
			filteredSpaces = append(filteredSpaces, space)
		}
	}

	// Filter stacks
	var filteredStacks []models.Stack
	for _, stack := range manifest.Stacks {
		if includedSpaces[stack.Space] {
			filteredStacks = append(filteredStacks, stack)
		}
	}

	// Filter contexts
	var filteredContexts []models.Context
	for _, ctx := range manifest.Contexts {
		if includedSpaces[ctx.Space] {
			filteredContexts = append(filteredContexts, ctx)
		}
	}

	// Filter policies
	var filteredPolicies []models.Policy
	for _, policy := range manifest.Policies {
		if includedSpaces[policy.Space] {
			filteredPolicies = append(filteredPolicies, policy)
		}
	}

	return &discovery.Manifest{
		SourceURL: manifest.SourceURL,
		Spaces:    filteredSpaces,
		Stacks:    filteredStacks,
		Contexts:  filteredContexts,
		Policies:  filteredPolicies,
	}
}

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

// newStatePlanCmd creates the state plan command.
func newStatePlanCmd() *cobra.Command {
	var spaceFilter string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Preview state migration for all stacks",
		Long: `Shows which stacks can have their Tofu state migrated.

Stacks are categorized as:
  ✓ Ready:    Managed state + external access enabled (can migrate)
  ⚠ Blocked:  Managed state but external access disabled (needs enabling)
  ○ Skipped:  Self-managed state (migrate via external backend)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatePlan(spaceFilter)
		},
	}
	cmd.Flags().StringVarP(&spaceFilter, "space", "s", "", "Only include stacks from this space")
	return cmd
}

// runStatePlan shows the state migration plan.
func runStatePlan(spaceFilter string) error {
	svc, err := createDiscoveryService()
	if err != nil {
		return err
	}

	ctx := context.Background()
	fmt.Println("Analyzing stacks for state migration...")

	stacks, err := svc.DiscoverStacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover stacks: %w", err)
	}

	// Filter by space if specified
	if spaceFilter != "" {
		fmt.Printf("Filtering to space: %s\n", spaceFilter)
		var filtered []models.Stack
		for _, stack := range stacks {
			if stack.Space == spaceFilter {
				filtered = append(filtered, stack)
			}
		}
		stacks = filtered
	}

	var ready, blocked, skipped, nonTofu []string
	readyStacks := make(map[string]bool)

	for _, stack := range stacks {
		if !stack.ManagesStateFile {
			skipped = append(skipped, stack.Name)
		} else if !stack.IsTerraform() {
			// Non-Terraform stacks (Ansible, Kubernetes, etc.) don't have TF state
			nonTofu = append(nonTofu, fmt.Sprintf("%s (%s)", stack.Name, friendlyVendorType(stack.VendorType)))
		} else if stack.ExternalStateAccessEnabled {
			ready = append(ready, stack.Name)
			readyStacks[stack.ID] = true
		} else {
			blocked = append(blocked, stack.Name)
		}
	}

	fmt.Println("\n┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    STATE MIGRATION PLAN                     │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")

	// Ready stacks
	fmt.Printf("\n✓ READY TO MIGRATE (%d stacks)\n", len(ready))
	if len(ready) > 0 {
		fmt.Println("  These stacks have managed state with external access enabled:")
		for _, name := range ready {
			fmt.Printf("    • %s\n", name)
		}
	} else {
		fmt.Println("  No stacks ready for migration")
	}

	// Blocked stacks
	fmt.Printf("\n⚠ BLOCKED - External State Access Disabled (%d stacks)\n", len(blocked))
	if len(blocked) > 0 {
		fmt.Println("  Enable external state access on these stacks first:")
		for _, name := range blocked {
			fmt.Printf("    • %s\n", name)
		}
		fmt.Println("\n  To enable, go to Stack Settings > Backend > Enable 'External State Access'")
		fmt.Println("  Or use the Spacelift API/Tofu to enable it")
	} else {
		fmt.Println("  All managed-state stacks have external access enabled")
	}

	// Skipped stacks
	fmt.Printf("\n○ SKIPPED - Self-Managed State (%d stacks)\n", len(skipped))
	if len(skipped) > 0 {
		fmt.Println("  These stacks use external backends (S3, GCS, etc.):")
		for _, name := range skipped {
			fmt.Printf("    • %s\n", name)
		}
		fmt.Println("\n  Migrate state via your external backend directly")
	} else {
		fmt.Println("  All stacks use Spacelift-managed state")
	}

	// Non-Tofu stacks
	if len(nonTofu) > 0 {
		fmt.Printf("\n○ N/A - Non-Tofu Stacks (%d stacks)\n", len(nonTofu))
		fmt.Println("  These stacks don't use Tofu state:")
		for _, name := range nonTofu {
			fmt.Printf("    • %s\n", name)
		}
	}

	// Summary
	fmt.Println("\n─────────────────────────────────────────────────────────────")
	fmt.Printf("Total: %d stacks | Ready: %d | Blocked: %d | Skipped: %d | N/A: %d\n",
		len(stacks), len(ready), len(blocked), len(skipped), len(nonTofu))

	if len(blocked) > 0 {
		fmt.Println("\n⚠️  Run: spacebridge state enable-access")
	} else if len(ready) > 0 {
		fmt.Println("\n✓ Ready to migrate! Run: spacebridge state migrate")
	}

	return nil
}

// newStateEnableAccessCmd creates the state enable-access command.
func newStateEnableAccessCmd() *cobra.Command {
	var spaceFilter string
	cmd := &cobra.Command{
		Use:   "enable-access",
		Short: "Enable external state access on all managed-state stacks",
		Long: `Enables external state access on all stacks that have Spacelift-managed state.
This is required before state can be downloaded for migration.

This command will:
  1. Find all stacks with managed state but external access disabled
  2. Enable external state access via the Spacelift API
  3. Report success/failure for each stack`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateEnableAccess(spaceFilter)
		},
	}
	cmd.Flags().StringVarP(&spaceFilter, "space", "s", "", "Only include stacks from this space")
	return cmd
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

// runStateEnableAccess enables external state access on blocked stacks.
func runStateEnableAccess(spaceFilter string) error {
	svc, err := createDiscoveryService()
	if err != nil {
		return err
	}

	ctx := context.Background()
	fmt.Println("Finding stacks that need external state access enabled...")

	stacks, err := svc.DiscoverStacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover stacks: %w", err)
	}

	// Filter by space if specified
	if spaceFilter != "" {
		fmt.Printf("Filtering to space: %s\n", spaceFilter)
		var filtered []models.Stack
		for _, stack := range stacks {
			if stack.Space == spaceFilter {
				filtered = append(filtered, stack)
			}
		}
		stacks = filtered
	}

	// Find blocked Terraform stacks (need full stack info for update)
	var blocked []models.Stack
	for _, stack := range stacks {
		// Only Terraform stacks have externalStateAccessEnabled
		if stack.ManagesStateFile && !stack.ExternalStateAccessEnabled && stack.IsTerraform() {
			blocked = append(blocked, stack)
		}
	}

	if len(blocked) == 0 {
		fmt.Println("\n✓ All managed-state stacks already have external access enabled!")
		return nil
	}

	fmt.Printf("\nEnabling external state access on %d stacks...\n\n", len(blocked))

	// Get the client directly for mutations
	c, err := client.New(cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	successCount := 0
	failCount := 0

	for _, stack := range blocked {
		fmt.Printf("  • %s ... ", stack.Name)
		if err := c.EnableExternalStateAccess(ctx, stack); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failCount++
		} else {
			fmt.Printf("✓ Enabled\n")
			successCount++
		}
	}

	fmt.Println("\n─────────────────────────────────────────────────────────────")
	fmt.Printf("Results: %d enabled, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d stacks failed to update", failCount)
	}

	fmt.Println("\n✓ All stacks ready for state migration!")
	fmt.Println("  Run: spacebridge state plan   # Verify all stacks are ready")
	fmt.Println("  Run: spacebridge state migrate # Migrate state")

	return nil
}

// newStateMigrateCmd creates the state migrate command.
func newStateMigrateCmd() *cobra.Command {
	var dryRun bool
	var spaceFilter string
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Tofu state from source to destination",
		Long: `Migrates Tofu state files from source Spacelift account to destination.

This command:
  1. Gets download URLs from source stacks (with external state access)
  2. Gets upload URLs from destination stacks (matched by stack name)
  3. Streams state directly between accounts (no local disk storage)
  4. Triggers state import on destination stacks

Prerequisites:
  - Destination stacks must already exist (run: Tofu apply on generated code)
  - Source stacks must have external state access enabled (run: spacebridge state enable-access)
  - Both SOURCE_* and DESTINATION_* environment variables must be configured

Use --dry-run to see what would be migrated without making changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateMigrate(dryRun, spaceFilter)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be migrated without making changes")
	cmd.Flags().StringVarP(&spaceFilter, "space", "s", "", "Only include stacks from this space")
	return cmd
}

// runStateMigrate performs the state migration.
func runStateMigrate(dryRun bool, spaceFilter string) error {
	// Validate both source and destination configs
	if err := cfg.ValidateSource(); err != nil {
		return fmt.Errorf("source configuration error: %w", err)
	}
	if err := cfg.ValidateDestination(); err != nil {
		return fmt.Errorf("destination configuration error: %w\n\nPlease set DESTINATION_SPACELIFT_URL, DESTINATION_SPACELIFT_KEY_ID, and DESTINATION_SPACELIFT_SECRET_KEY", err)
	}

	ctx := context.Background()

	// Create clients
	sourceClient, err := client.New(cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to create source client: %w", err)
	}

	destClient, err := client.New(cfg.Destination)
	if err != nil {
		return fmt.Errorf("failed to create destination client: %w", err)
	}

	fmt.Printf("Source:      %s\n", cfg.Source.URL)
	fmt.Printf("Destination: %s\n", cfg.Destination.URL)
	if spaceFilter != "" {
		fmt.Printf("Space:       %s\n", spaceFilter)
	}

	// Discover stacks from both accounts
	fmt.Println("\nDiscovering stacks...")
	sourceSvc := discovery.New(sourceClient)
	destSvc := discovery.New(destClient)

	sourceStacks, err := sourceSvc.DiscoverStacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover source stacks: %w", err)
	}

	// Filter source stacks by space if specified
	if spaceFilter != "" {
		var filtered []models.Stack
		for _, stack := range sourceStacks {
			if stack.Space == spaceFilter {
				filtered = append(filtered, stack)
			}
		}
		sourceStacks = filtered
	}

	destStacks, err := destSvc.DiscoverStacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover destination stacks: %w", err)
	}

	// Build map of destination stacks by name
	destStackMap := make(map[string]models.Stack)
	for _, stack := range destStacks {
		destStackMap[stack.Name] = stack
	}

	// Find stacks eligible for migration
	type migrationCandidate struct {
		Source models.Stack
		Dest   models.Stack
	}
	var candidates []migrationCandidate
	var skipped []string
	var notInDest []string
	var noAccess []string

	for _, stack := range sourceStacks {
		// Only Tofu stacks with managed state
		if !stack.ManagesStateFile {
			skipped = append(skipped, stack.Name+" (self-managed state)")
			continue
		}
		if !stack.IsTerraform() {
			skipped = append(skipped, stack.Name+" ("+friendlyVendorType(stack.VendorType)+")")
			continue
		}
		if !stack.ExternalStateAccessEnabled {
			noAccess = append(noAccess, stack.Name)
			continue
		}

		// Find matching destination stack
		destStack, exists := destStackMap[stack.Name]
		if !exists {
			notInDest = append(notInDest, stack.Name)
			continue
		}

		candidates = append(candidates, migrationCandidate{
			Source: stack,
			Dest:   destStack,
		})
	}

	// Print migration plan
	fmt.Println("\n┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    STATE MIGRATION                          │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")

	if len(candidates) == 0 {
		fmt.Println("\n⚠ No stacks eligible for migration.")
		if len(noAccess) > 0 {
			fmt.Printf("\n  %d stacks need external state access enabled:\n", len(noAccess))
			for _, name := range noAccess {
				fmt.Printf("    • %s\n", name)
			}
			fmt.Println("\n  Run: spacebridge state enable-access")
		}
		if len(notInDest) > 0 {
			fmt.Printf("\n  %d stacks not found in destination:\n", len(notInDest))
			for _, name := range notInDest {
				fmt.Printf("    • %s\n", name)
			}
			fmt.Println("\n  Apply Tofu to create destination stacks first")
		}
		return nil
	}

	fmt.Printf("\n✓ WILL MIGRATE (%d stacks)\n", len(candidates))
	for _, c := range candidates {
		fmt.Printf("    • %s\n", c.Source.Name)
	}

	if len(skipped) > 0 {
		fmt.Printf("\n○ SKIPPED (%d stacks)\n", len(skipped))
		for _, name := range skipped {
			fmt.Printf("    • %s\n", name)
		}
	}

	if len(notInDest) > 0 {
		fmt.Printf("\n⚠ NOT IN DESTINATION (%d stacks)\n", len(notInDest))
		for _, name := range notInDest {
			fmt.Printf("    • %s\n", name)
		}
	}

	if len(noAccess) > 0 {
		fmt.Printf("\n⚠ NO EXTERNAL ACCESS (%d stacks)\n", len(noAccess))
		for _, name := range noAccess {
			fmt.Printf("    • %s\n", name)
		}
	}

	if dryRun {
		fmt.Println("\n─────────────────────────────────────────────────────────────")
		fmt.Println("DRY RUN - No changes made")
		fmt.Println("Remove --dry-run flag to perform migration")
		return nil
	}

	// Perform migration
	fmt.Println("\n─────────────────────────────────────────────────────────────")
	fmt.Println("Starting state migration...")

	successCount := 0
	failCount := 0

	for _, c := range candidates {
		fmt.Printf("\n  Migrating: %s\n", c.Source.Name)

		// Get download URL from source
		fmt.Print("    Getting download URL... ")
		downloadURL, err := sourceClient.GetStateDownloadURL(ctx, c.Source.ID)
		if err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failCount++
			continue
		}
		fmt.Println("✓")

		// Get upload URL from destination
		fmt.Print("    Getting upload URL... ")
		uploadURL, err := destClient.GetStateUploadURL(ctx, c.Dest.ID)
		if err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failCount++
			continue
		}
		fmt.Println("✓")

		// Stream state from source to destination
		fmt.Print("    Streaming state... ")
		stateReader, contentLength, err := client.StreamStateFromURL(ctx, downloadURL)
		if err != nil {
			fmt.Printf("✗ Failed to download: %v\n", err)
			failCount++
			continue
		}

		err = client.UploadStateToURL(ctx, uploadURL, stateReader, contentLength)
		stateReader.Close()
		if err != nil {
			fmt.Printf("✗ Failed to upload: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("✓ (%d bytes)\n", contentLength)

		// Trigger state import
		fmt.Print("    Importing state... ")
		if err := destClient.ImportManagedState(ctx, c.Dest.ID); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failCount++
			continue
		}
		fmt.Println("✓")

		successCount++
	}

	// Print summary
	fmt.Println("\n─────────────────────────────────────────────────────────────")
	fmt.Printf("Migration complete: %d succeeded, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d stacks failed to migrate", failCount)
	}

	fmt.Println("\n✓ All states migrated successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Verify state in destination stacks (Spacelift UI > Stack > State)")
	fmt.Println("  2. Enable stacks: spacebridge stacks enable")
	fmt.Println("  3. Trigger runs to verify infrastructure matches")

	return nil
}

// newStacksEnableCmd creates the stacks enable command.
func newStacksEnableCmd() *cobra.Command {
	var dryRun bool
	var spaceFilter string
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable all disabled stacks in destination",
		Long: `Enables all disabled stacks in the destination Spacelift account.

Use this command after:
  1. Creating stacks with is_disabled = true (Tofu apply)
  2. Migrating state (spacebridge state migrate)

This command will:
  1. Find all disabled stacks in the destination account
  2. Enable each stack (set is_disabled = false)
  3. Report success/failure for each stack

Note: This command operates on the DESTINATION account.

Use --dry-run to see what would be enabled without making changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStacksEnable(dryRun, spaceFilter)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be enabled without making changes")
	cmd.Flags().StringVarP(&spaceFilter, "space", "s", "", "Only include stacks from this space")
	return cmd
}

// runStacksEnable enables all disabled stacks in the destination.
func runStacksEnable(dryRun bool, spaceFilter string) error {
	// Validate destination config
	if err := cfg.ValidateDestination(); err != nil {
		return fmt.Errorf("destination configuration error: %w\n\nPlease set DESTINATION_SPACELIFT_URL, DESTINATION_SPACELIFT_KEY_ID, and DESTINATION_SPACELIFT_SECRET_KEY", err)
	}

	ctx := context.Background()

	// Create destination client
	destClient, err := client.New(cfg.Destination)
	if err != nil {
		return fmt.Errorf("failed to create destination client: %w", err)
	}

	fmt.Printf("Destination: %s\n", cfg.Destination.URL)
	if spaceFilter != "" {
		fmt.Printf("Space:       %s\n", spaceFilter)
	}
	fmt.Println("\nDiscovering disabled stacks...")

	// Discover stacks from destination
	destSvc := discovery.New(destClient)
	stacks, err := destSvc.DiscoverStacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover stacks: %w", err)
	}

	// Filter by space if specified
	if spaceFilter != "" {
		var filtered []models.Stack
		for _, stack := range stacks {
			if stack.Space == spaceFilter {
				filtered = append(filtered, stack)
			}
		}
		stacks = filtered
	}

	// Find disabled stacks
	var disabled []models.Stack
	for _, stack := range stacks {
		if stack.IsDisabled {
			disabled = append(disabled, stack)
		}
	}

	if len(disabled) == 0 {
		fmt.Println("\n✓ No disabled stacks found!")
		return nil
	}

	fmt.Printf("\nFound %d disabled stacks:\n", len(disabled))
	for _, stack := range disabled {
		fmt.Printf("    • %s\n", stack.Name)
	}

	if dryRun {
		fmt.Println("\n─────────────────────────────────────────────────────────────")
		fmt.Println("DRY RUN - No changes made")
		fmt.Println("Remove --dry-run flag to enable stacks")
		return nil
	}

	// Enable stacks
	fmt.Println("\n─────────────────────────────────────────────────────────────")
	fmt.Println("Enabling stacks...")

	successCount := 0
	failCount := 0

	for _, stack := range disabled {
		fmt.Printf("  • %s ... ", stack.Name)
		if err := destClient.EnableStack(ctx, stack); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failCount++
		} else {
			fmt.Printf("✓ Enabled\n")
			successCount++
		}
	}

	// Print summary
	fmt.Println("\n─────────────────────────────────────────────────────────────")
	fmt.Printf("Results: %d enabled, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d stacks failed to enable", failCount)
	}

	fmt.Println("\n✓ All stacks enabled!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Trigger runs on enabled stacks to verify state matches infrastructure")
	fmt.Println("  2. Monitor runs for any drift or issues")

	return nil
}
