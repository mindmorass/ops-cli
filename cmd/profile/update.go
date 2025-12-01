package profile

import (
	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var force bool
	var dryRun bool
	var noBackup bool

	cmd := &cobra.Command{
		Use:   "update [profile-name]",
		Short: "Update an existing profile with new features",
		Long: `Update an existing profile with new features and configurations.

This command adds missing directories, environment variables, and configuration files to existing profiles.
Useful when new features are added to the profile manager.`,
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

			opts := commands.UpdateOptions{
				ProfileName: "",
				Force:       force,
				DryRun:      dryRun,
				NoBackup:    noBackup,
			}

			if len(args) > 0 {
				opts.ProfileName = args[0]
			}

			return commands.UpdateProfile(cfg.ProfilesDir, opts)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files without prompting")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Skip creating backup before updating")

	return cmd
}

