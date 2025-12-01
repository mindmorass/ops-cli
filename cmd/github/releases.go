package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newReleasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases <owner/repo>",
		Short: "List GitHub releases",
		Long: `List GitHub releases for a repository.

Examples:
  ops-cli github releases octocat/Hello-World
  ops-cli github releases github/docs --format json`,
		Args: cobra.ExactArgs(1),
		RunE: runReleases,
	}

	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runReleases(cmd *cobra.Command, args []string) error {
	repoPath := args[0]
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
		return fmt.Errorf("failed to list releases: %w", err)
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
