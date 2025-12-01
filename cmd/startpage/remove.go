package startpage

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove bookmarks from your startpage",
		Long: `Remove bookmarks from your startpage interactively.

Examples:
  ops-cli startpage remove`,
		RunE: runRemove,
	}

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
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

	// Get all bookmarks from all groups
	var allBookmarks []Bookmark
	for _, group := range config.Bookmarks {
		allBookmarks = append(allBookmarks, group.Bookmarks...)
	}

	if len(allBookmarks) == 0 {
		fmt.Println("No bookmarks to remove.")
		fmt.Println("Add some bookmarks first with: ops-cli startpage add")
		return nil
	}

	// Create choices for survey multiselect
	choices := make([]string, len(allBookmarks))
	for i, bookmark := range allBookmarks {
		icon := bookmark.Icon
		if icon == "" {
			icon = "🔗"
		}
		choices[i] = fmt.Sprintf("%s %s (%s)", icon, bookmark.Name, bookmark.URL)
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message: "Select bookmarks to remove:",
		Options: choices,
	}
	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No bookmarks selected for removal.")
		return nil
	}

	// Confirm deletion
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to remove %d bookmark(s)?", len(selectedIndices)),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil || !confirm {
		fmt.Println("Removal cancelled.")
		return nil
	}

	// Collect IDs to remove
	idsToRemove := make(map[string]bool)
	for _, idx := range selectedIndices {
		if idx >= 0 && idx < len(allBookmarks) {
			idsToRemove[allBookmarks[idx].ID] = true
		}
	}

	// Remove selected bookmarks from all groups
	removedNames := []string{}
	for i := range config.Bookmarks {
		filtered := []Bookmark{}
		for _, bookmark := range config.Bookmarks[i].Bookmarks {
			if !idsToRemove[bookmark.ID] {
				filtered = append(filtered, bookmark)
			} else {
				removedNames = append(removedNames, bookmark.Name)
			}
		}
		config.Bookmarks[i].Bookmarks = filtered
	}

	// Save updated config
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if err := generateStartpage(config); err != nil {
		return fmt.Errorf("failed to generate startpage: %w", err)
	}

	fmt.Printf("Removed %d bookmark(s):\n", len(removedNames))
	for _, name := range removedNames {
		fmt.Printf("   • %s\n", name)
	}
	fmt.Println("Rebuilding startpage...")

	if err := rebuildStartpage(config); err != nil {
		fmt.Println("Bookmarks removed but rebuild failed. Run 'ops-cli startpage build' manually.")
		return nil
	}

	fmt.Println("Startpage rebuilt successfully!")
	fmt.Println("Changes are now visible on your site!")

	return nil
}
