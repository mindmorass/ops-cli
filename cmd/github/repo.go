package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo <owner/repo>",
		Short: "Get a specific GitHub repository",
		Long: `Get detailed information about a specific GitHub repository.

Examples:
  ops-cli github repo octocat/Hello-World
  ops-cli github repo github/docs --format json`,
		Args: cobra.ExactArgs(1),
		RunE: runRepo,
	}

	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runRepo(cmd *cobra.Command, args []string) error {
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

	format, _ := cmd.Flags().GetString("format")

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching repository: %s...", repoPath))
	repository, err := client.GetRepository(owner, repo)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	if format == "json" {
		output, err := json.MarshalIndent(repository, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Formatted output
	if repository.FullName != nil {
		fmt.Println(*repository.FullName)
	}
	if repository.Description != nil && *repository.Description != "" {
		fmt.Println(*repository.Description)
	}

	stars := 0
	if repository.StargazersCount != nil {
		stars = *repository.StargazersCount
	}
	forks := 0
	if repository.ForksCount != nil {
		forks = *repository.ForksCount
	}
	fmt.Printf("⭐ %d stars | 🍴 %d forks\n", stars, forks)

	if repository.HTMLURL != nil {
		fmt.Printf("🔗 %s\n", *repository.HTMLURL)
	}

	if repository.Language != nil {
		fmt.Printf("💻 Primary language: %s\n", *repository.Language)
	}

	private := false
	if repository.Private != nil {
		private = *repository.Private
	}
	if private {
		fmt.Println("🔒 Private repository")
	} else {
		fmt.Println("🔓 Public repository")
	}

	if repository.CreatedAt != nil {
		fmt.Printf("Created: %s\n", repository.CreatedAt.Format(time.DateOnly))
	}
	if repository.UpdatedAt != nil {
		fmt.Printf("Updated: %s\n", repository.UpdatedAt.Format(time.DateOnly))
	}

	if repository.CloneURL != nil {
		fmt.Println("\nClone URLs:")
		fmt.Printf("   HTTPS: %s\n", *repository.CloneURL)
		if repository.SSHURL != nil {
			fmt.Printf("   SSH:   %s\n", *repository.SSHURL)
		}
	}

	return nil
}
