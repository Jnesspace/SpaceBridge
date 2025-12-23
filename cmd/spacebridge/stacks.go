package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/internal/discovery"
	"github.com/jnesspace/spacebridge/internal/models"
)

// newStacksCmd creates the stacks command group.
func newStacksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stacks",
		Short: "Manage stacks in destination account",
	}
	cmd.AddCommand(
		newStacksEnableCmd(),
	)
	return cmd
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
