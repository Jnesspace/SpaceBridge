package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/ui"
)

// newDiscoverCmd creates the discover command group.
func newDiscoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover resources in the source Spacelift account",
	}

	cmd.AddCommand(
		newDiscoverSpacesCmd(),
		newDiscoverStacksCmd(),
		newDiscoverContextsCmd(),
		newDiscoverPoliciesCmd(),
		newDiscoverAllCmd(),
	)

	return cmd
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
