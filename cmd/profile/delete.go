package profile

import (
	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "delete [profile-name]",
		Short: "Delete a workspace profile",
		Long: `Delete a workspace profile and all its files.

Interactive selection is enabled by default if profile name is omitted.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				// Use default if config doesn't exist
				cfg, err = config.GetDefaultConfig()
				if err != nil {
					return err
				}
			}

			opts := commands.DeleteOptions{
				ProfileName: "",
				Force:       force,
				DryRun:      dryRun,
			}

			if len(args) > 0 {
				opts.ProfileName = args[0]
			}

			return commands.DeleteProfile(cfg.ProfilesDir, opts)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt (disables interactive)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without deleting (disables interactive)")

	return cmd
}

