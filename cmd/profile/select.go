package profile

import (
	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newSelectCmd() *cobra.Command {
	var allowDirenv bool

	cmd := &cobra.Command{
		Use:   "select [profile-name]",
		Short: "Select and switch to a profile",
		Long: `Select and switch to a workspace profile.

This command helps you select a profile and provides instructions on how to activate it.
The profile is activated by changing to its directory, which automatically loads the profile's environment via direnv.`,
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

			opts := commands.SelectOptions{
				ProfileName: "",
				AllowDirenv: allowDirenv,
			}

			if len(args) > 0 {
				opts.ProfileName = args[0]
			}

			return commands.SelectProfile(cfg.ProfilesDir, opts)
		},
	}

	cmd.Flags().BoolVar(&allowDirenv, "allow-direnv", false, "Automatically allow direnv for the selected profile")

	return cmd
}

