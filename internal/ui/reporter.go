package ui

import (
	"fmt"
	"strings"

	"github.com/jnesspace/spacebridge/internal/discovery"
	"github.com/jnesspace/spacebridge/internal/models"
)

// PrintSummary prints a summary of the discovered resources.
func PrintSummary(manifest *discovery.Manifest) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("DISCOVERY SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Source: %s\n\n", manifest.SourceURL)

	summary := manifest.Summary()
	fmt.Printf("  Spaces:   %d\n", summary["spaces"])
	fmt.Printf("  Stacks:   %d\n", summary["stacks"])
	fmt.Printf("  Contexts: %d\n", summary["contexts"])
	fmt.Printf("  Policies: %d\n", summary["policies"])
	fmt.Println()

	secretsCount := manifest.SecretsCount()
	if secretsCount > 0 {
		fmt.Printf("  ⚠️  Secrets requiring manual entry: %d\n", secretsCount)
	}

	fmt.Println(strings.Repeat("=", 50))
}

// PrintSpaces prints spaces in a formatted way.
func PrintSpaces(spaces []models.Space) {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("SPACES")
	fmt.Println(strings.Repeat("-", 40))

	trees := models.BuildSpaceTree(spaces)
	fmt.Print(RenderSpaceTree(trees))
}

// PrintStacks prints stacks in a formatted table.
func PrintStacks(stacks []models.Stack) {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Printf("STACKS (%d total)\n", len(stacks))
	fmt.Println(strings.Repeat("-", 40))

	if len(stacks) == 0 {
		fmt.Println("No stacks found.")
		return
	}

	headers := []string{"ID", "Name", "Space", "Repository", "Branch"}
	rows := make([][]string, 0, len(stacks))

	for _, stack := range stacks {
		rows = append(rows, []string{
			truncate(stack.ID, 25),
			truncate(stack.Name, 25),
			truncate(stack.Space, 15),
			truncate(stack.Repository, 30),
			truncate(stack.Branch, 15),
		})
	}

	fmt.Print(RenderTable(headers, rows))
}

// PrintContexts prints contexts in a formatted table.
func PrintContexts(contexts []models.Context) {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Printf("CONTEXTS (%d total)\n", len(contexts))
	fmt.Println(strings.Repeat("-", 40))

	if len(contexts) == 0 {
		fmt.Println("No contexts found.")
		return
	}

	headers := []string{"ID", "Name", "Space", "Config Items", "Secrets"}
	rows := make([][]string, 0, len(contexts))

	for _, ctx := range contexts {
		secretCount := len(ctx.GetSecretConfigs())
		secretsStr := fmt.Sprintf("%d", secretCount)
		if secretCount > 0 {
			secretsStr += " ⚠️"
		}

		rows = append(rows, []string{
			truncate(ctx.ID, 25),
			truncate(ctx.Name, 25),
			truncate(ctx.Space, 15),
			fmt.Sprintf("%d", len(ctx.Config)),
			secretsStr,
		})
	}

	fmt.Print(RenderTable(headers, rows))
}

// PrintPolicies prints policies in a formatted table.
func PrintPolicies(policies []models.Policy) {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Printf("POLICIES (%d total)\n", len(policies))
	fmt.Println(strings.Repeat("-", 40))

	if len(policies) == 0 {
		fmt.Println("No policies found.")
		return
	}

	headers := []string{"ID", "Name", "Type", "Space"}
	rows := make([][]string, 0, len(policies))

	for _, pol := range policies {
		rows = append(rows, []string{
			truncate(pol.ID, 25),
			truncate(pol.Name, 25),
			truncate(pol.Type, 15),
			truncate(pol.Space, 15),
		})
	}

	fmt.Print(RenderTable(headers, rows))
}

// PrintSecretsWarning prints a warning about secrets that need manual entry.
func PrintSecretsWarning(contexts []models.Context) {
	secretContexts := make([]models.Context, 0)
	for _, ctx := range contexts {
		if ctx.HasSecrets() {
			secretContexts = append(secretContexts, ctx)
		}
	}

	if len(secretContexts) == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("!", 50))
	fmt.Println("SECRETS REQUIRING MANUAL ENTRY")
	fmt.Println(strings.Repeat("!", 50))
	fmt.Println("The following contexts contain secrets that cannot")
	fmt.Println("be exported via the API. You will need to manually")
	fmt.Println("re-enter these values in the destination account.")
	fmt.Println()

	for _, ctx := range secretContexts {
		secrets := ctx.GetSecretConfigs()
		fmt.Printf("  Context: %s\n", ctx.ID)
		for _, secret := range secrets {
			fmt.Printf("    - %s (%s)\n", secret.ID, secret.Type)
		}
		fmt.Println()
	}
}

// truncate truncates a string to a maximum length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
