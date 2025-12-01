package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newWorkflowsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflows [owner/repo]",
		Short: "List GitHub workflows",
		Long: `List GitHub workflows for a repository.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github workflows octocat/Hello-World
  ops-cli github workflows github/docs --format json
  ops-cli github workflows  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runWorkflows,
	}

	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runWorkflows(cmd *cobra.Command, args []string) error {
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

	format, _ := cmd.Flags().GetString("format")

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching workflows for %s...", repoPath))
	workflows, err := client.ListWorkflows(owner, repo)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List workflows")
	}

	if len(workflows) == 0 {
		fmt.Println("No workflows found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(workflows, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Workflows for %s:\n\n", repoPath)

	for i, workflow := range workflows {
		name := ""
		if workflow.Name != nil {
			name = *workflow.Name
		}
		path := ""
		if workflow.Path != nil {
			path = *workflow.Path
		}
		state := ""
		if workflow.State != nil {
			state = *workflow.State
		}

		fmt.Printf("%d. %s\n", i+1, name)
		fmt.Printf("   Path: %s\n", path)
		fmt.Printf("   State: %s\n", state)

		if workflow.UpdatedAt != nil {
			fmt.Printf("   Updated: %s\n", workflow.UpdatedAt.Format(time.DateOnly))
		}

		if workflow.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *workflow.HTMLURL)
		}
		fmt.Println()
	}

	return nil
}
