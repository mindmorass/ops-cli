package github

import (
	"fmt"
	"strings"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file [owner/repo] <file-path>",
		Short: "Get file content from a GitHub repository",
		Long: `Get the content of a specific file from a GitHub repository.

If no repository is provided, it will be detected from the current git directory.

Examples:
  ops-cli github file octocat/Hello-World README.md
  ops-cli github file microsoft/vscode package.json --ref main
  ops-cli github file README.md  # Uses current git repository`,
		Args: cobra.MinimumNArgs(1),
		RunE: runFile,
	}

	cmd.Flags().String("ref", "", "Git reference (branch, tag, or commit SHA)")

	return cmd
}

func runFile(cmd *cobra.Command, args []string) error {
	var repoPath, filePath string
	var err error

	// Determine if first arg is repo or file path
	if len(args) > 1 && strings.Contains(args[0], "/") {
		// First arg is a repository
		repoPath = args[0]
		filePath = args[1]
	} else {
		// First arg is file path, detect repo from git
		repoPath, err = getRepoArg(args, 0)
		if err != nil {
			return err
		}
		filePath = args[0]
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

	ref, _ := cmd.Flags().GetString("ref")

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching file: %s...", filePath))
	content, err := client.GetFileContent(owner, repo, filePath, ref)
	stopSpinner()
	if err != nil {
		return handleGitHubError(err, owner, repo, "Get file content")
	}

	fmt.Println(content)
	return nil
}
