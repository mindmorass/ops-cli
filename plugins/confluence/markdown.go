package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newMarkdownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markdown [flags] [markdown-file]",
		Short: "Sync Markdown files to Confluence pages using mark",
		Long: `Sync Markdown files to Confluence pages using the mark tool.

This command wraps the 'mark' tool and automatically configures it with your
Atlassian credentials from the configuration.

Examples:
  # Sync a markdown file (URL from file metadata)
  ops-cli confluence markdown README.md

  # Sync with explicit URL
  ops-cli confluence markdown --file README.md --url https://company.atlassian.net/wiki/spaces/SPACE/pages/12345

  # Sync with title from H1
  ops-cli confluence markdown docs/page.md --title-from-h1

  # Dry run to preview changes
  ops-cli confluence markdown README.md --dry-run`,
		RunE: runMarkdown,
	}

	// Mark command flags
	cmd.Flags().StringP("file", "f", "", "Markdown file(s) to sync (required, supports glob patterns)")
	cmd.Flags().StringP("url", "l", "", "Target Confluence page URL (optional, can be specified in markdown file metadata)")
	cmd.Flags().Bool("continue-on-error", false, "Continue processing remaining files if an error occurs")
	cmd.Flags().Bool("compile-only", false, "Show resulting HTML without updating Confluence page")
	cmd.Flags().Bool("dry-run", false, "Resolve page and ancestry, show resulting HTML and exit")
	cmd.Flags().BoolP("edit-lock", "k", false, "Lock page editing to current user only")
	cmd.Flags().Bool("drop-h1", false, "Don't include the first H1 heading in Confluence output")
	cmd.Flags().BoolP("strip-linebreaks", "L", false, "Remove linebreaks inside of tags")
	cmd.Flags().Bool("title-from-h1", false, "Extract page title from a leading H1 heading")
	cmd.Flags().Bool("title-from-filename", false, "Use filename as the Confluence page title")
	cmd.Flags().Bool("title-append-generated-hash", false, "Append a short hash to the title")
	cmd.Flags().Bool("minor-edit", false, "Don't send notifications while updating page")
	cmd.Flags().String("version-message", "", "Add a message to the page version")
	cmd.Flags().String("log-level", "info", "Set the log level (TRACE, DEBUG, INFO, WARNING, ERROR, FATAL)")

	return cmd
}

func runMarkdown(cmd *cobra.Command, args []string) error {
	// Check if mark is installed
	markPath, err := exec.LookPath("mark")
	if err != nil {
		return fmt.Errorf("mark tool not found. Please install it with: brew install mark")
	}

	// Load config to get credentials
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get Confluence credentials
	baseURL, username, token := cfg.GetConfluenceCredentials()

	// Environment variables take precedence
	if envURL := os.Getenv("CONFLUENCE_BASE_URL"); envURL != "" {
		baseURL = envURL
	}
	if envUser := os.Getenv("CONFLUENCE_USERNAME"); envUser != "" {
		username = envUser
	}
	if envToken := os.Getenv("ATLASSIAN_TOKEN"); envToken != "" {
		token = envToken
	}

	if baseURL == "" || username == "" || token == "" {
		return fmt.Errorf("Confluence credentials not configured. Run 'ops-cli confluence config setup'")
	}

	// Get required flags
	file, _ := cmd.Flags().GetString("file")
	url, _ := cmd.Flags().GetString("url")

	// If file is not provided as flag, check args
	if file == "" && len(args) > 0 {
		file = args[0]
	}

	if file == "" {
		return fmt.Errorf("markdown file is required. Use --file or provide as argument")
	}

	// Expand file path if it's relative
	if !filepath.IsAbs(file) {
		// Check if file exists in current directory
		if _, err := os.Stat(file); err == nil {
			file, _ = filepath.Abs(file)
		}
	}

	// Build mark command
	markCmd := exec.Command(markPath)

	// Add credentials as command-line flags (more reliable than env vars)
	markCmd.Args = append(markCmd.Args, "--username", username)
	markCmd.Args = append(markCmd.Args, "--password", token)

	// Add base URL (required by mark to know which Confluence instance to use)
	markCmd.Args = append(markCmd.Args, "--base-url", baseURL)

	// Add required flags
	markCmd.Args = append(markCmd.Args, "--files", file)

	// Target URL is optional - if provided, use it; otherwise mark will read from file metadata
	if url != "" {
		markCmd.Args = append(markCmd.Args, "--target-url", url)
	}

	// Add optional flags
	if cmd.Flags().Changed("continue-on-error") {
		markCmd.Args = append(markCmd.Args, "--continue-on-error")
	}
	if cmd.Flags().Changed("compile-only") {
		markCmd.Args = append(markCmd.Args, "--compile-only")
	}
	if cmd.Flags().Changed("dry-run") {
		markCmd.Args = append(markCmd.Args, "--dry-run")
	}
	if cmd.Flags().Changed("edit-lock") {
		markCmd.Args = append(markCmd.Args, "--edit-lock")
	}
	if cmd.Flags().Changed("drop-h1") {
		markCmd.Args = append(markCmd.Args, "--drop-h1")
	}
	if cmd.Flags().Changed("strip-linebreaks") {
		markCmd.Args = append(markCmd.Args, "--strip-linebreaks")
	}
	if cmd.Flags().Changed("title-from-h1") {
		markCmd.Args = append(markCmd.Args, "--title-from-h1")
	}
	if cmd.Flags().Changed("title-from-filename") {
		markCmd.Args = append(markCmd.Args, "--title-from-filename")
	}
	if cmd.Flags().Changed("title-append-generated-hash") {
		markCmd.Args = append(markCmd.Args, "--title-append-generated-hash")
	}
	if cmd.Flags().Changed("minor-edit") {
		markCmd.Args = append(markCmd.Args, "--minor-edit")
	}
	if versionMsg, _ := cmd.Flags().GetString("version-message"); versionMsg != "" {
		markCmd.Args = append(markCmd.Args, "--version-message", versionMsg)
	}
	if logLevel, _ := cmd.Flags().GetString("log-level"); logLevel != "" {
		markCmd.Args = append(markCmd.Args, "--log-level", logLevel)
	}

	// Set up command output
	markCmd.Stdout = os.Stdout
	markCmd.Stderr = os.Stderr
	markCmd.Stdin = os.Stdin

	// Show what we're doing
	fmt.Printf("Syncing Markdown to Confluence...\n")
	fmt.Printf("  File: %s\n", file)
	if url != "" {
		fmt.Printf("  URL: %s\n", url)
	} else {
		fmt.Printf("  URL: (will be read from file metadata)\n")
	}
	fmt.Printf("  Base URL: %s\n", baseURL)
	fmt.Println()

	// Run the mark command
	stopSpinner := ui.StartSpinner("Running mark...")
	err = markCmd.Run()
	stopSpinner()

	if err != nil {
		return fmt.Errorf("mark command failed: %w", err)
	}

	fmt.Println("\n✓ Markdown sync completed successfully!")
	return nil
}
