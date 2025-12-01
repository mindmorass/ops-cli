package startpage

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all bookmarks in your startpage",
		Long: `List all bookmarks in your startpage.

Examples:
  ops-cli startpage list
  ops-cli startpage list --format json`,
		RunE: runList,
	}

	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
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

	format, _ := cmd.Flags().GetString("format")

	if format == "json" {
		data, err := json.MarshalIndent(config.Bookmarks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal bookmarks: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Bookmarks for \"%s\":\n\n", config.Name)

	// Get all bookmarks from all groups
	var allBookmarks []Bookmark
	for _, group := range config.Bookmarks {
		allBookmarks = append(allBookmarks, group.Bookmarks...)
	}

	if len(allBookmarks) == 0 {
		fmt.Println("   (no bookmarks yet)")
		fmt.Println("   Add your first bookmark with: ops-cli startpage add")
		return nil
	}

	for i, bookmark := range allBookmarks {
		icon := bookmark.Icon
		if icon == "" {
			icon = "🔗"
		}
		fmt.Printf("   %d. %s %s\n", i+1, icon, bookmark.Name)
		fmt.Printf("      %s\n", bookmark.URL)
		if bookmark.Description != "" {
			fmt.Printf("      %s\n", bookmark.Description)
		}
	}
	fmt.Println()

	return nil
}
