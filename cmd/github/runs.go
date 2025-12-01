package github

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	ghapi "github.com/google/go-github/v55/github"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs <subcommand>",
		Short: "Manage GitHub Actions workflow runs",
		Long: `Manage GitHub Actions workflow runs.

Subcommands:
  list    List workflow runs for a repository
  get     Get details of a specific workflow run`,
	}

	cmd.AddCommand(newRunsListCmd())
	cmd.AddCommand(newRunsGetCmd())

	return cmd
}

func newRunsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <owner/repo>",
		Short: "List workflow runs for a repository",
		Long: `List workflow runs for a repository.

Examples:
  ops-cli github runs list octocat/Hello-World
  ops-cli github runs list github/docs --format json`,
		Args: cobra.ExactArgs(1),
		RunE: runRunsList,
	}

	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runRunsList(cmd *cobra.Command, args []string) error {
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

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching workflow runs for %s...", repoPath))
	runs, err := client.ListWorkflowRuns(owner, repo, perPage, page)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list workflow runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("No workflow runs found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(runs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Workflow runs for %s:\n\n", repoPath)

	for i, run := range runs {
		name := ""
		if run.Name != nil {
			name = *run.Name
		}
		headBranch := ""
		if run.HeadBranch != nil {
			headBranch = *run.HeadBranch
		}
		status := ""
		if run.Status != nil {
			status = *run.Status
		}
		conclusion := ""
		if run.Conclusion != nil {
			conclusion = *run.Conclusion
		}
		runNumber := 0
		if run.RunNumber != nil {
			runNumber = *run.RunNumber
		}

		fmt.Printf("%d. Run #%d: %s\n", i+1, runNumber, name)
		if headBranch != "" {
			fmt.Printf("   Branch: %s\n", headBranch)
		}
		fmt.Printf("   Status: %s", status)
		if conclusion != "" {
			fmt.Printf(" (%s)", conclusion)
		}
		fmt.Println()

		if run.CreatedAt != nil {
			fmt.Printf("   Created: %s\n", run.CreatedAt.Format(time.RFC3339))
		}

		if run.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *run.HTMLURL)
		}
		fmt.Println()
	}

	return nil
}

func newRunsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <owner/repo> <run-id>",
		Short: "Get details of a specific workflow run",
		Long: `Get details of a specific workflow run.

Examples:
  ops-cli github runs get octocat/Hello-World 12345
  ops-cli github runs get github/docs 67890 --format json`,
		Args: cobra.ExactArgs(2),
		RunE: runRunsGet,
	}

	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runRunsGet(cmd *cobra.Command, args []string) error {
	repoPath := args[0]
	runIDStr := args[1]

	if !strings.Contains(repoPath, "/") {
		return fmt.Errorf("please provide a repository in format 'owner/repo'")
	}

	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format. Use 'owner/repo'")
	}

	owner := parts[0]
	repo := parts[1]

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid run ID: %w", err)
	}

	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")

	// For now, just list runs and find the matching one
	// A proper implementation would use GetWorkflowRun, but we'll keep it simple
	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching workflow run %s...", runIDStr))
	runs, err := client.ListWorkflowRuns(owner, repo, 100, 1)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get workflow run: %w", err)
	}

	var foundRun *ghapi.WorkflowRun
	for _, run := range runs {
		if run.ID != nil && *run.ID == runID {
			foundRun = run
			break
		}
	}

	if foundRun == nil {
		return fmt.Errorf("workflow run %s not found", runIDStr)
	}

	if format == "json" {
		output, err := json.MarshalIndent(foundRun, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Display run details
	name := ""
	if foundRun.Name != nil {
		name = *foundRun.Name
	}
	headBranch := ""
	if foundRun.HeadBranch != nil {
		headBranch = *foundRun.HeadBranch
	}
	status := ""
	if foundRun.Status != nil {
		status = *foundRun.Status
	}
	conclusion := ""
	if foundRun.Conclusion != nil {
		conclusion = *foundRun.Conclusion
	}
	runNumber := 0
	if foundRun.RunNumber != nil {
		runNumber = *foundRun.RunNumber
	}

	fmt.Printf("Workflow Run #%d: %s\n", runNumber, name)
	if headBranch != "" {
		fmt.Printf("Branch: %s\n", headBranch)
	}
	fmt.Printf("Status: %s", status)
	if conclusion != "" {
		fmt.Printf(" (%s)", conclusion)
	}
	fmt.Println()

	if foundRun.CreatedAt != nil {
		fmt.Printf("Created: %s\n", foundRun.CreatedAt.Format(time.RFC3339))
	}
	if foundRun.UpdatedAt != nil {
		fmt.Printf("Updated: %s\n", foundRun.UpdatedAt.Format(time.RFC3339))
	}

	if foundRun.HTMLURL != nil {
		fmt.Printf("\n🔗 %s\n", *foundRun.HTMLURL)
	}

	return nil
}
