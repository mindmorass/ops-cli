package startpage

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new bookmark to your startpage",
		Long: `Add a new bookmark to your startpage.

Examples:
  ops-cli startpage add
  ops-cli startpage add --name "GitHub" --url "https://github.com"`,
		RunE: runAdd,
	}

	cmd.Flags().String("name", "", "Bookmark name")
	cmd.Flags().String("url", "", "Bookmark URL")
	cmd.Flags().String("icon", "", "Bookmark icon (emoji or SVG path)")
	cmd.Flags().String("description", "", "Bookmark description")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	if err := ensureStartpageDirectory(); err != nil {
		return fmt.Errorf("failed to create startpage directory: %w", err)
	}

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no startpage found. Run 'ops-cli startpage init' first")
	}

	// Get flags
	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	icon, _ := cmd.Flags().GetString("icon")
	description, _ := cmd.Flags().GetString("description")

	// Interactive prompts if not provided
	if name == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "Bookmark name:",
		}, &name, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("bookmark name is required")
		}
	}

	if url == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "Bookmark URL:",
		}, &url, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("bookmark URL is required")
		}
	}

	if icon == "" {
		// Auto-detect icon based on URL (simplified)
		icon = getIconForURL(url)
		fmt.Printf("Auto-detected icon: %s\n", icon)
	}

	if description == "" {
		survey.AskOne(&survey.Input{
			Message: "Description (optional):",
		}, &description)
	}

	// Ensure there's at least one group
	if len(config.Bookmarks) == 0 {
		config.Bookmarks = append(config.Bookmarks, BookmarkGroup{
			ID:        "general",
			Name:      "General",
			Icon:      "📁",
			Bookmarks: []Bookmark{},
		})
	}

	// Add bookmark to first group
	targetGroup := &config.Bookmarks[0]
	newBookmark := Bookmark{
		ID:          strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		Name:        name,
		URL:         url,
		Icon:        icon,
		Description: description,
	}

	targetGroup.Bookmarks = append(targetGroup.Bookmarks, newBookmark)

	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if err := generateStartpage(config); err != nil {
		return fmt.Errorf("failed to generate startpage: %w", err)
	}

	fmt.Printf("Added bookmark \"%s\"\n", newBookmark.Name)
	fmt.Println("Rebuilding startpage...")

	if err := rebuildStartpage(config); err != nil {
		fmt.Println("Bookmark added but rebuild failed. Run 'ops-cli startpage build' manually.")
		return nil
	}

	fmt.Println("Startpage rebuilt successfully!")
	fmt.Println("Your new bookmark is now visible on the site!")

	return nil
}

// getIconForURL returns an icon for a URL (simplified)
func getIconForURL(url string) string {
	// Simple icon mapping based on domain
	urlLower := strings.ToLower(url)
	if strings.Contains(urlLower, "github") {
		return "🐙"
	}
	if strings.Contains(urlLower, "gmail") || strings.Contains(urlLower, "mail") {
		return "📧"
	}
	if strings.Contains(urlLower, "vscode") {
		return "💻"
	}
	return "🔗"
}

// rebuildStartpage rebuilds the startpage (shared function)
func rebuildStartpage(config *StartpageConfig) error {
	if err := generateStartpage(config); err != nil {
		return err
	}

	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = config.Path
	buildCmd.Stdout = nil // Suppress output
	buildCmd.Stderr = nil
	return buildCmd.Run()
}
