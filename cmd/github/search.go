package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ops-cli/internal/api/github"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <owner/repo> <query>",
		Short: "Search for files in a GitHub repository",
		Long: `Search for files within a GitHub repository.

Examples:
  ops-cli github search octocat/Hello-World README
  ops-cli github search github/docs --filename package.json
  ops-cli github search microsoft/vscode --extension ts --path src`,
		Args: cobra.MinimumNArgs(2),
		RunE: runSearch,
	}

	cmd.Flags().String("path", "", "Filter by path")
	cmd.Flags().String("filename", "", "Filter by filename")
	cmd.Flags().String("extension", "", "Filter by file extension")
	cmd.Flags().Int("per-page", 30, "Results per page")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	repoPath := args[0]
	searchQuery := args[1]

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

	path, _ := cmd.Flags().GetString("path")
	filename, _ := cmd.Flags().GetString("filename")
	extension, _ := cmd.Flags().GetString("extension")
	perPage, _ := cmd.Flags().GetInt("per-page")
	page, _ := cmd.Flags().GetInt("page")
	format, _ := cmd.Flags().GetString("format")

	options := github.FileSearchOptions{
		Owner:     owner,
		Repo:      repo,
		Query:     searchQuery,
		Path:      path,
		Filename:  filename,
		Extension: extension,
		PerPage:   perPage,
		Page:      page,
	}

	stopSpinner := ui.StartSpinner(fmt.Sprintf("Searching files in %s for: \"%s\"...", repoPath, searchQuery))
	results, err := client.SearchFiles(options)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to search files: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No files found matching the search criteria.")
		return nil
	}

	if format == "json" {
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table format
	fmt.Printf("Found %d files:\n\n", len(results))

	for i, result := range results {
		path := ""
		if result.Path != nil {
			path = *result.Path
		}
		fmt.Printf("%d. %s\n", i+1, path)

		name := ""
		if result.Name != nil {
			name = *result.Name
		}
		fmt.Printf("   %s\n", name)

		if result.HTMLURL != nil {
			fmt.Printf("   🔗 %s\n", *result.HTMLURL)
		}

		// Score is not directly available in CodeResult, skip it
		fmt.Println()
	}

	return nil
}
