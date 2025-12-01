package atlassian

import (
	"github.com/spf13/cobra"
)

// Register registers the atlassian module with the root command
func Register(rootCmd *cobra.Command) {
	atlassianCmd := &cobra.Command{
		Use:   "atlassian",
		Short: "Manage shared Atlassian configuration",
		Long: `Atlassian Configuration Commands

Manage shared Atlassian API configuration used by both Jira and Confluence.
This shared configuration can be used by both services, with individual service
configs acting as overrides when needed.`,
	}

	atlassianCmd.AddCommand(newConfigCmd())

	rootCmd.AddCommand(atlassianCmd)
}

