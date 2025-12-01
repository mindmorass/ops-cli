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

func newReleasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases [owner/repo]",
		Short: "List GitHub releases",
		Long: `List GitHub releases for a repository.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github releases octocat/Hello-World
  ops-cli github releases github/docs --format json
  ops-cli github releases  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runReleases,
	}

	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	cmd.AddCommand(newReleasesDeleteCmd())

	return cmd
}

func runReleases(cmd *cobra.Command, args []string) error {
	repoPath, err := getRepoArg(args, 0)
	if err != nil {
		return err
	}

	if !strings.Contains(repoPath, "/") {
		return fmt.Errorf("please provide a repository in format 'owner/repo'")
	}

	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format. Use 'owner/repo'")
	}

	owner := parts[0]
	repo := parts[1]

	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	perPage, _ := cmd.Flags().GetInt("per-page")
	page, _ := cmd.Flags().GetInt("page")
	format, _ := cmd.Flags().GetString("format")

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching releases for %s...", repoPath))
	releases, err := client.ListReleases(owner, repo, perPage, page)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List releases")
	}

	if len(releases) == 0 {
		fmt.Println("No releases found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(releases, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Releases for %s:\n\n", repoPath)

	for i, release := range releases {
		tagName := ""
		if release.TagName != nil {
			tagName = *release.TagName
		}
		name := ""
		if release.Name != nil {
			name = *release.Name
		}
		if name == "" {
			name = tagName
		}

		fmt.Printf("%d. %s\n", i+1, name)
		if release.TagName != nil {
			fmt.Printf("   Tag: %s\n", *release.TagName)
		}

		draft := false
		if release.Draft != nil {
			draft = *release.Draft
		}
		prerelease := false
		if release.Prerelease != nil {
			prerelease = *release.Prerelease
		}

		if draft {
			fmt.Println("   [Draft]")
		} else if prerelease {
			fmt.Println("   [Pre-release]")
		}

		if release.PublishedAt != nil {
			fmt.Printf("   Published: %s\n", release.PublishedAt.Format(time.DateOnly))
		}

		if release.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *release.HTMLURL)
		}

		if len(release.Assets) > 0 {
			fmt.Printf("   Assets: %d\n", len(release.Assets))
		}
		fmt.Println()
	}

	return nil
}

func newReleasesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [owner/repo]",
		Short: "Delete GitHub releases interactively",
		Long: `Delete GitHub releases with interactive selection.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github releases delete octocat/Hello-World
  ops-cli github releases delete  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runReleasesDelete,
	}

	return cmd
}

func runReleasesDelete(cmd *cobra.Command, args []string) error {
	repoPath, err := getRepoArg(args, 0)
	if err != nil {
		return err
	}

	if !strings.Contains(repoPath, "/") {
		return fmt.Errorf("please provide a repository in format 'owner/repo'")
	}

	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format. Use 'owner/repo'")
	}

	owner := parts[0]
	repo := parts[1]

	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	// List all releases
	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching releases for %s...", repoPath))
	releases, err := client.ListReleases(owner, repo, 100, 1)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List releases")
	}

	if len(releases) == 0 {
		fmt.Println("No releases found to delete.")
		return nil
	}

	// Build selection options
	choices := make([]string, len(releases))
	for i, release := range releases {
		tagName := ""
		if release.TagName != nil {
			tagName = *release.TagName
		}
		name := ""
		if release.Name != nil {
			name = *release.Name
		}
		if name == "" {
			name = tagName
		}
		draft := ""
		if release.Draft != nil && *release.Draft {
			draft = " [Draft]"
		}
		prerelease := ""
		if release.Prerelease != nil && *release.Prerelease {
			prerelease = " [Pre-release]"
		}
		choices[i] = fmt.Sprintf("• %s (%s)%s%s", name, tagName, draft, prerelease)
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message: "Select releases to delete:",
		Options: choices,
	}
	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No releases selected for deletion.")
		return nil
	}

	// Confirm deletion
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to delete %d release(s)? This action cannot be undone.", len(selectedIndices)),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirm {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete selected releases
	fmt.Printf("\nDeleting %d release(s)...\n", len(selectedIndices))
	for _, idx := range selectedIndices {
		release := releases[idx]
		releaseID := int64(0)
		if release.ID != nil {
			releaseID = *release.ID
		}
		tagName := ""
		if release.TagName != nil {
			tagName = *release.TagName
		}

		stopSpinner := ui.StartSpinner(fmt.Sprintf("Deleting release %s...", tagName))
		err := client.DeleteRelease(owner, repo, releaseID)
		stopSpinner()

		if err != nil {
			fmt.Printf("✗ Failed to delete release %s: %v\n", tagName, err)
		} else {
			fmt.Printf("✓ Deleted release %s\n", tagName)
		}
	}

	fmt.Println("\n✓ Deletion completed!")
	return nil
}
