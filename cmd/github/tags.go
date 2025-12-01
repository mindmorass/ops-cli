package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags <owner/repo>",
		Short: "List GitHub tags",
		Long: `List GitHub tags for a repository.

Examples:
  ops-cli github tags octocat/Hello-World
  ops-cli github tags github/docs --format json`,
		Args: cobra.ExactArgs(1),
		RunE: runTags,
	}

	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runTags(cmd *cobra.Command, args []string) error {
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

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching tags for %s...", repoPath))
	tags, err := client.ListTags(owner, repo, perPage, page)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(tags, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Tags for %s:\n\n", repoPath)

	for i, tag := range tags {
		name := ""
		if tag.Name != nil {
			name = *tag.Name
		}
		fmt.Printf("%d. %s\n", i+1, name)

		if tag.Commit != nil && tag.Commit.SHA != nil {
			sha := *tag.Commit.SHA
			if len(sha) > 7 {
				sha = sha[:7]
			}
			fmt.Printf("   Commit: %s\n", sha)
		}

		if tag.ZipballURL != nil {
			fmt.Printf("   🔗 %s\n", *tag.ZipballURL)
		}
		fmt.Println()
	}

	return nil
}
