package github

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
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
  get     Get details of a specific workflow run
  delete  Delete workflow runs interactively`,
	}

	cmd.AddCommand(newRunsListCmd())
	cmd.AddCommand(newRunsGetCmd())
	cmd.AddCommand(newRunsDeleteCmd())

	return cmd
}

func newRunsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [owner/repo]",
		Short: "List workflow runs for a repository",
		Long: `List workflow runs for a repository.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github runs list octocat/Hello-World
  ops-cli github runs list github/docs --format json
  ops-cli github runs list  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRunsList,
	}

	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runRunsList(cmd *cobra.Command, args []string) error {
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

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching workflow runs for %s...", repoPath))
	runs, err := client.ListWorkflowRuns(owner, repo, perPage, page)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List workflow runs")
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
		Use:   "get [owner/repo] <run-id>",
		Short: "Get details of a specific workflow run",
		Long: `Get details of a specific workflow run.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github runs get octocat/Hello-World 12345
  ops-cli github runs get github/docs 67890 --format json
  ops-cli github runs get 12345  # Uses current git repository`,
		Args: cobra.MinimumNArgs(1),
		RunE: runRunsGet,
	}

	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runRunsGet(cmd *cobra.Command, args []string) error {
	var repoPath, runIDStr string
	var err error

	// Determine if first arg is repo or run-id
	if len(args) > 1 && strings.Contains(args[0], "/") {
		// First arg is a repository
		repoPath = args[0]
		runIDStr = args[1]
	} else {
		// First arg is run-id, detect repo from git
		repoPath, err = getRepoArg(args, 0)
		if err != nil {
			return err
		}
		runIDStr = args[0]
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
		return handleGitHubError(err, owner, repo, "Get workflow run")
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

func newRunsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [owner/repo]",
		Short: "Delete GitHub workflow runs interactively",
		Long: `Delete GitHub workflow runs with interactive selection.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github runs delete octocat/Hello-World
  ops-cli github runs delete  # Uses current git repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRunsDelete,
	}

	return cmd
}

func runRunsDelete(cmd *cobra.Command, args []string) error {
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

	// List all workflow runs
	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching workflow runs for %s...", repoPath))
	runs, err := client.ListWorkflowRuns(owner, repo, 100, 1)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "List workflow runs")
	}

	if len(runs) == 0 {
		fmt.Println("No workflow runs found to delete.")
		return nil
	}

	// Build selection options
	choices := make([]string, len(runs))
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
		statusStr := status
		if conclusion != "" {
			statusStr = fmt.Sprintf("%s (%s)", status, conclusion)
		}
		choices[i] = fmt.Sprintf("• Run #%d: %s - %s [%s]", runNumber, name, headBranch, statusStr)
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message: "Select workflow runs to delete:",
		Options: choices,
	}
	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No workflow runs selected for deletion.")
		return nil
	}

	// Confirm deletion
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to delete %d workflow run(s)? This action cannot be undone.", len(selectedIndices)),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirm {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete selected runs
	fmt.Printf("\nDeleting %d workflow run(s)...\n", len(selectedIndices))
	for _, idx := range selectedIndices {
		run := runs[idx]
		runID := int64(0)
		if run.ID != nil {
			runID = *run.ID
		}
		runNumber := 0
		if run.RunNumber != nil {
			runNumber = *run.RunNumber
		}

		stopSpinner := ui.StartSpinner(fmt.Sprintf("Deleting workflow run #%d...", runNumber))
		err := client.DeleteWorkflowRun(owner, repo, runID)
		stopSpinner()

		if err != nil {
			fmt.Printf("✗ Failed to delete workflow run #%d: %v\n", runNumber, err)
		} else {
			fmt.Printf("✓ Deleted workflow run #%d\n", runNumber)
		}
	}

	fmt.Println("\n✓ Deletion completed!")
	return nil
}
