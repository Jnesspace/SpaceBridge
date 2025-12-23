// SpaceBridge - Spacelift Enterprise Migration Kit
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/jnesspace/spacebridge/internal/client"
	"github.com/jnesspace/spacebridge/pkg/config"
)

var (
	verbose bool
	cfg     *config.Config
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

	// Add command groups
	rootCmd.AddCommand(
		newDiscoverCmd(),
		newExportCmd(),
		newGenerateCmd(),
		newStateCmd(),
		newStacksCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
