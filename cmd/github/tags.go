package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags [owner/repo]",
		Short: "List GitHub tags",
		Long: `List GitHub tags for a repository.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github tags octocat/Hello-World
  ops-cli github tags github/docs --format json
  ops-cli github tags  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTags,
	}

	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	cmd.AddCommand(newTagsDeleteCmd())

	return cmd
}

func runTags(cmd *cobra.Command, args []string) error {
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

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching tags for %s...", repoPath))
	tags, err := client.ListTags(owner, repo, perPage, page)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List tags")
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

func newTagsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [owner/repo]",
		Short: "Delete GitHub tags interactively",
		Long: `Delete GitHub tags with interactive selection.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github tags delete octocat/Hello-World
  ops-cli github tags delete  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTagsDelete,
	}

	return cmd
}

func runTagsDelete(cmd *cobra.Command, args []string) error {
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

	// List all tags
	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching tags for %s...", repoPath))
	tags, err := client.ListTags(owner, repo, 100, 1)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List tags")
	}

	if len(tags) == 0 {
		fmt.Println("No tags found to delete.")
		return nil
	}

	// Build selection options
	choices := make([]string, len(tags))
	for i, tag := range tags {
		name := ""
		if tag.Name != nil {
			name = *tag.Name
		}
		sha := ""
		if tag.Commit != nil && tag.Commit.SHA != nil {
			sha = *tag.Commit.SHA
			if len(sha) > 7 {
				sha = sha[:7]
			}
		}
		choices[i] = fmt.Sprintf("• %s (commit: %s)", name, sha)
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message: "Select tags to delete:",
		Options: choices,
	}
	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No tags selected for deletion.")
		return nil
	}

	// Confirm deletion
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to delete %d tag(s)? This action cannot be undone.", len(selectedIndices)),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirm {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete selected tags
	fmt.Printf("\nDeleting %d tag(s)...\n", len(selectedIndices))
	for _, idx := range selectedIndices {
		tag := tags[idx]
		tagName := ""
		if tag.Name != nil {
			tagName = *tag.Name
		}

		stopSpinner := ui.StartSpinner(fmt.Sprintf("Deleting tag %s...", tagName))
		err := client.DeleteTag(owner, repo, tagName)
		stopSpinner()

		if err != nil {
			fmt.Printf("✗ Failed to delete tag %s: %v\n", tagName, err)
		} else {
			fmt.Printf("✓ Deleted tag %s\n", tagName)
		}
	}

	fmt.Println("\n✓ Deletion completed!")
	return nil
}
