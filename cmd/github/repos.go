package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ops-cli/internal/api/github"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newReposCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List GitHub repositories",
		Long: `List GitHub repositories for a user, organization, or search query.

Examples:
  ops-cli github repos --user octocat
  ops-cli github repos --org github --type public
  ops-cli github repos --format json`,
		RunE: runRepos,
	}

	cmd.Flags().String("user", "", "GitHub username")
	cmd.Flags().String("org", "", "GitHub organization")
	cmd.Flags().String("query", "", "Search query")
	cmd.Flags().String("type", "all", "Repository type: all, owner, member")
	cmd.Flags().String("sort", "updated", "Sort field: created, updated, pushed, full_name, name")
	cmd.Flags().String("direction", "desc", "Sort direction: asc, desc")
	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runRepos(cmd *cobra.Command, args []string) error {
	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	user, _ := cmd.Flags().GetString("user")
	org, _ := cmd.Flags().GetString("org")
	query, _ := cmd.Flags().GetString("query")
	repoType, _ := cmd.Flags().GetString("type")
	sort, _ := cmd.Flags().GetString("sort")
	direction, _ := cmd.Flags().GetString("direction")
	perPage, _ := cmd.Flags().GetInt("per-page")
	page, _ := cmd.Flags().GetInt("page")
	format, _ := cmd.Flags().GetString("format")

	options := github.RepositoryListOptions{
		User:      user,
		Org:       org,
		Query:     query,
		Type:      repoType,
		Sort:      sort,
		Direction: direction,
		PerPage:   perPage,
		Page:      page,
	}

	stopSpinner := ui.StartSpinner("Fetching repositories...")
	repos, err := client.ListRepositories(options)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(repos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Found %d repositories:\n\n", len(repos))

	for i, repo := range repos {
		fullName := ""
		if repo.FullName != nil {
			fullName = *repo.FullName
		}
		fmt.Printf("%d. %s\n", i+1, fullName)

		if repo.Description != nil && *repo.Description != "" {
			fmt.Printf("   %s\n", *repo.Description)
		}

		stars := 0
		if repo.StargazersCount != nil {
			stars = *repo.StargazersCount
		}
		forks := 0
		if repo.ForksCount != nil {
			forks = *repo.ForksCount
		}
		fmt.Printf("   ⭐ %d stars | 🍴 %d forks\n", stars, forks)

		if repo.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *repo.HTMLURL)
		}

		if repo.Language != nil {
			fmt.Printf("   💻 %s\n", *repo.Language)
		}

		if repo.UpdatedAt != nil {
			fmt.Printf("   Updated: %s\n", repo.UpdatedAt.Format(time.DateOnly))
		}
		fmt.Println()
	}

	return nil
}
