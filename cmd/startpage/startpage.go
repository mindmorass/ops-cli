package startpage

import (
	"github.com/spf13/cobra"
)

// Register registers the startpage module with the root command
func Register(rootCmd *cobra.Command) {
	startpageCmd := &cobra.Command{
		Use:   "startpage",
		Short: "Setup and manage Astro-based browser startpage",
		Long: `Startpage Commands

Setup and manage an Astro-based browser startpage using the OpenLinks theme.
Manage bookmarks, themes, and serve your personalized startpage.`,
	}

	startpageCmd.AddCommand(newInitCmd())
	startpageCmd.AddCommand(newAddCmd())
	startpageCmd.AddCommand(newListCmd())
	startpageCmd.AddCommand(newRemoveCmd())
	startpageCmd.AddCommand(newDevCmd())
	startpageCmd.AddCommand(newBuildCmd())
	startpageCmd.AddCommand(newServeCmd())
	startpageCmd.AddCommand(newStopCmd())
	startpageCmd.AddCommand(newThemeCmd())

	rootCmd.AddCommand(startpageCmd)
}
