package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newPackagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages <owner>",
		Short: "List GitHub packages",
		Long: `List GitHub packages for a user or organization.

Examples:
  ops-cli github packages octocat
  ops-cli github packages myorg --type npm`,
		Args: cobra.ExactArgs(1),
		RunE: runPackages,
	}

	cmd.Flags().String("type", "", "Package type: npm, maven, rubygems, docker, nuget, container")
	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runPackages(cmd *cobra.Command, args []string) error {
	owner := args[0]

	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	packageType, _ := cmd.Flags().GetString("type")
	perPage, _ := cmd.Flags().GetInt("per-page")
	page, _ := cmd.Flags().GetInt("page")
	format, _ := cmd.Flags().GetString("format")

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching packages for %s...", owner))
	packages, err := client.ListPackages(owner, packageType, perPage, page)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Println("No packages found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(packages, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Packages for %s:\n\n", owner)

	for i, pkg := range packages {
		name := ""
		if pkg.Name != nil {
			name = *pkg.Name
		}
		pkgType := ""
		if pkg.PackageType != nil {
			pkgType = *pkg.PackageType
		}
		visibility := ""
		if pkg.Visibility != nil {
			visibility = string(*pkg.Visibility)
		}
		versionCount := int64(0)
		if pkg.VersionCount != nil {
			versionCount = *pkg.VersionCount
		}

		fmt.Printf("%d. %s\n", i+1, name)
		fmt.Printf("   Type: %s\n", pkgType)
		fmt.Printf("   Visibility: %s\n", visibility)
		fmt.Printf("   Versions: %d\n", versionCount)

		if pkg.UpdatedAt != nil {
			fmt.Printf("   Updated: %s\n", pkg.UpdatedAt.Format(time.DateOnly))
		}

		if pkg.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *pkg.HTMLURL)
		}
		fmt.Println()
	}

	return nil
}
