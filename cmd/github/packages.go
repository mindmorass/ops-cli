package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newPackagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages [owner]",
		Short: "List GitHub packages",
		Long: `List GitHub packages for a user or organization.

If no owner is provided, it will be detected from the current git repository.

Examples:
  ops-cli github packages octocat
  ops-cli github packages myorg --type npm
  ops-cli github packages  # Uses owner from current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runPackages,
	}

	cmd.Flags().String("type", "", "Package type: npm, maven, rubygems, docker, nuget, container")
	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	cmd.AddCommand(newPackagesDeleteCmd())

	return cmd
}

func runPackages(cmd *cobra.Command, args []string) error {
	var owner string
	var err error

	if len(args) > 0 {
		owner = args[0]
	} else {
		// Detect owner from git repository
		repo, err := getRepoArg(args, 0)
		if err != nil {
			return err
		}
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("could not extract owner from repository: %s", repo)
		}
		owner = parts[0]
	}

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

func newPackagesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [owner]",
		Short: "Delete GitHub packages interactively",
		Long: `Delete GitHub packages with interactive selection.

If no owner is provided, it will be detected from the current git repository.

Note: This deletes entire packages. To delete specific versions, use the GitHub web interface.

Examples:
  ops-cli github packages delete octocat
  ops-cli github packages delete  # Uses owner from current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runPackagesDelete,
	}

	cmd.Flags().String("type", "", "Package type: npm, maven, rubygems, docker, nuget, container")

	return cmd
}

func runPackagesDelete(cmd *cobra.Command, args []string) error {
	var owner string
	var err error

	if len(args) > 0 {
		owner = args[0]
	} else {
		// Detect owner from git repository
		repo, err := getRepoArg(args, 0)
		if err != nil {
			return err
		}
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("could not extract owner from repository: %s", repo)
		}
		owner = parts[0]
	}

	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	packageType, _ := cmd.Flags().GetString("type")

	// List packages
	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching packages for %s...", owner))
	packages, err := client.ListPackages(owner, packageType, 100, 1)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Println("No packages found to delete.")
		return nil
	}

	// Build selection options
	choices := make([]string, len(packages))
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
		choices[i] = fmt.Sprintf("• %s (%s) - %s [%d versions]", name, pkgType, visibility, versionCount)
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message: "Select packages to delete:",
		Options: choices,
	}
	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No packages selected for deletion.")
		return nil
	}

	// Confirm deletion
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to delete %d package(s)? This will delete ALL versions. This action cannot be undone.", len(selectedIndices)),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirm {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Note: Package deletion requires deleting all versions first
	// For now, we'll show a message that this needs to be done via API or web interface
	fmt.Println("\n⚠ Package deletion requires deleting all versions first.")
	fmt.Println("This operation is complex and should be done via the GitHub web interface")
	fmt.Println("or by deleting each version individually using the API.")
	fmt.Println("\nSelected packages:")
	for _, idx := range selectedIndices {
		pkg := packages[idx]
		name := ""
		if pkg.Name != nil {
			name = *pkg.Name
		}
		pkgType := ""
		if pkg.PackageType != nil {
			pkgType = *pkg.PackageType
		}
		fmt.Printf("  • %s (%s)\n", name, pkgType)
	}

	return fmt.Errorf("package deletion not fully implemented - please use GitHub web interface")
}
