package main

import (
	"github.com/spf13/cobra"
)

func newStaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stale",
		Short: "Find and comment on stale Jira issues",
		Long: `Find and comment on stale Jira issues.

Examples:
  ops-cli jira stale --days 7
  ops-cli jira stale --days 5 --comment --color`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement stale command
			return nil
		},
	}

	cmd.Flags().Int("days", 7, "Number of days to consider stale")
	cmd.Flags().Bool("comment", false, "Add comment to stale issues")
	cmd.Flags().String("project", "", "Filter by project key")

	return cmd
}
