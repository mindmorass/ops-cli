package profile

import (
	"github.com/ops-cli/internal/profile"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show information about the current profile",
		Long: `Show information about the current workspace profile.

Displays the active profile name, location, git configuration, and environment variables.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				// Use default if config doesn't exist
				cfg, err = config.GetDefaultConfig()
				if err != nil {
					return err
				}
			}

			pm := profile.NewManager(cfg.ProfilesDir)
			return pm.ShowInfo()
		},
	}

	return cmd
}

