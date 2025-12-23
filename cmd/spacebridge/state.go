package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/discovery"
	"github.com/jnesspace/spacebridge/internal/models"
)


// newStateCmd creates the state command group.
func newStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage Tofu state migration",
	}
	cmd.AddCommand(
		newStatePlanCmd(),
		newStateEnableAccessCmd(),
		newStateMigrateCmd(),
	)
	return cmd
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

// resolveSpaceFilter resolves a space filter (ID, name, or name-ID format) to a space ID.
// Returns the resolved space ID and the display name for output.
// Supports formats: "01ABC..." (ID), "migration" (name), "migration-01ABC..." (name-ID)
func resolveSpaceFilter(ctx context.Context, svc *discovery.Service, filter string) (spaceID string, displayName string, err error) {
	spaces, err := svc.DiscoverSpaces(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to discover spaces: %w", err)
	}

	// Build a map for quick lookup
	spaceByID := make(map[string]models.Space)
	for _, space := range spaces {
		spaceByID[space.ID] = space
	}

	// First, try to match by exact ID
	if space, ok := spaceByID[filter]; ok {
		return space.ID, space.Name, nil
	}

	// Then, try to match by name (case-sensitive)
	for _, space := range spaces {
		if space.Name == filter {
			return space.ID, space.Name, nil
		}
	}

	// Finally, try to extract ID from "name-ID" format (e.g., "migration-01ABC...")
	// Find the last hyphen and check if the suffix is a valid space ID
	for i := len(filter) - 1; i >= 0; i-- {
		if filter[i] == '-' {
			potentialID := filter[i+1:]
			if space, ok := spaceByID[potentialID]; ok {
				return space.ID, space.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("space not found: %s", filter)
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
		spaceID, spaceName, err := resolveSpaceFilter(ctx, svc, spaceFilter)
		if err != nil {
			return err
		}
		fmt.Printf("Filtering to space: %s (ID: %s)\n", spaceName, spaceID)
		var filtered []models.Stack
		for _, stack := range stacks {
			if stack.Space == spaceID {
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
		spaceID, spaceName, err := resolveSpaceFilter(ctx, svc, spaceFilter)
		if err != nil {
			return err
		}
		fmt.Printf("Filtering to space: %s (ID: %s)\n", spaceName, spaceID)
		var filtered []models.Stack
		for _, stack := range stacks {
			if stack.Space == spaceID {
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

	// Discover stacks from both accounts
	fmt.Println("\nDiscovering stacks...")
	sourceSvc := discovery.New(sourceClient)
	destSvc := discovery.New(destClient)

	// Resolve space filter if specified (using source account spaces)
	var resolvedSpaceID string
	if spaceFilter != "" {
		spaceID, spaceName, err := resolveSpaceFilter(ctx, sourceSvc, spaceFilter)
		if err != nil {
			return err
		}
		resolvedSpaceID = spaceID
		fmt.Printf("Space:       %s (ID: %s)\n", spaceName, spaceID)
	}

	sourceStacks, err := sourceSvc.DiscoverStacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover source stacks: %w", err)
	}

	// Filter source stacks by space if specified
	if resolvedSpaceID != "" {
		var filtered []models.Stack
		for _, stack := range sourceStacks {
			if stack.Space == resolvedSpaceID {
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
		uploadResult, err := destClient.GetStateUploadURL(ctx, c.Dest.ID)
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

		err = client.UploadStateToURL(ctx, uploadResult.URL, stateReader, contentLength)
		stateReader.Close()
		if err != nil {
			fmt.Printf("✗ Failed to upload: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("✓ (%d bytes)\n", contentLength)

		// Lock stack, import state, then unlock
		fmt.Print("    Locking stack... ")
		if err := destClient.LockStack(ctx, c.Dest.ID); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failCount++
			continue
		}
		fmt.Println("✓")

		fmt.Print("    Importing state... ")
		if err := destClient.ImportManagedState(ctx, c.Dest.ID, uploadResult.ObjectID); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			// Try to unlock even if import failed
			destClient.UnlockStack(ctx, c.Dest.ID)
			failCount++
			continue
		}
		fmt.Println("✓")

		fmt.Print("    Unlocking stack... ")
		if err := destClient.UnlockStack(ctx, c.Dest.ID); err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			// Don't count as failure since state was imported
		} else {
			fmt.Println("✓")
		}

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
