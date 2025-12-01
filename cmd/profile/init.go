package profile

import (
	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool
	var interactive bool
	var profilesDir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the profile manager configuration",
		Long: `Initialize the profile manager configuration.

This command creates a configuration file that stores the path to your profiles directory.
If not initialized, the tool will use the default path: ~/workspaces/profiles`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				// Use default if config doesn't exist
				cfg, err = config.GetDefaultConfig()
				if err != nil {
					return err
				}
			}

			opts := commands.InitOptions{
				ProfilesDir: profilesDir,
				Force:       force,
				Interactive: interactive,
			}

			// If profilesDir not provided, use current config or default
			if opts.ProfilesDir == "" {
				opts.ProfilesDir = cfg.ProfilesDir
			}

			return commands.InitConfig(opts)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive setup (prompt for paths)")
	cmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Set profiles directory path")

	return cmd
}

