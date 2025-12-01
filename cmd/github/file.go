package github

import (
	"fmt"
	"strings"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file <owner/repo> <file-path>",
		Short: "Get file content from a GitHub repository",
		Long: `Get the content of a specific file from a GitHub repository.

Examples:
  ops-cli github file octocat/Hello-World README.md
  ops-cli github file microsoft/vscode package.json --ref main`,
		Args: cobra.ExactArgs(2),
		RunE: runFile,
	}

	cmd.Flags().String("ref", "", "Git reference (branch, tag, or commit SHA)")

	return cmd
}

func runFile(cmd *cobra.Command, args []string) error {
	repoPath := args[0]
	filePath := args[1]

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
		return fmt.Errorf("failed to get file content: %w", err)
	}

	fmt.Println(content)
	return nil
}
