package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newContentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contents <owner/repo> [path]",
		Short: "List repository contents",
		Long: `List the contents of a GitHub repository directory.

Examples:
  ops-cli github contents octocat/Hello-World
  ops-cli github contents github/docs src
  ops-cli github contents microsoft/vscode --ref main`,
		Args: cobra.MinimumNArgs(1),
		RunE: runContents,
	}

	cmd.Flags().String("ref", "", "Git reference (branch, tag, or commit SHA)")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runContents(cmd *cobra.Command, args []string) error {
	repoPath := args[0]
	path := ""
	if len(args) > 1 {
		path = args[1]
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
	format, _ := cmd.Flags().GetString("format")

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching contents for %s/%s...", repoPath, path))
	contents, err := client.ListRepositoryContents(owner, repo, path, ref)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list contents: %w", err)
	}

	if len(contents) == 0 {
		fmt.Println("No contents found.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(contents, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Contents of %s/%s:\n\n", repoPath, path)

	for _, item := range contents {
		name := ""
		if item.Name != nil {
			name = *item.Name
		}
		itemType := ""
		if item.Type != nil {
			itemType = *item.Type
		}
		size := int64(0)
		if item.Size != nil {
			size = int64(*item.Size)
		}

		icon := "📄"
		if itemType == "dir" {
			icon = "📁"
		}

		fmt.Printf("%s %s", icon, name)
		if itemType == "file" && size > 0 {
			fmt.Printf(" (%s)", formatFileSize(size))
		}
		fmt.Println()

		if item.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *item.HTMLURL)
		}
		fmt.Println()
	}

	return nil
}

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB"}

	i := 0
	size := float64(bytes)
	for size >= unit && i < len(sizes)-1 {
		size /= unit
		i++
	}

	return fmt.Sprintf("%.1f %s", size, sizes[i])
}
