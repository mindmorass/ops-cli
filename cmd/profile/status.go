package profile

import (
	"github.com/ops-cli/internal/profile"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show direnv status",
		Long: `Show the status of direnv installation and configuration.

Checks if direnv is installed and shows its current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return profile.ShowDirenvStatus()
		},
	}

	return cmd
}

