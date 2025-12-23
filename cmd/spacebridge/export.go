package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/ui"
)

var outputFile string

// newExportCmd creates the export command.
func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all resources to a manifest file",
		RunE:  runExport,
	}
	cmd.Flags().StringVarP(&outputFile, "output", "o", "manifest.json", "Output file path")
	return cmd
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
