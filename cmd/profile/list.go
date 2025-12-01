package profile

import (
	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var verbose bool
	var showConfig bool
	var noInteractive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workspace profiles",
		Long: `List all workspace profiles with their configurations.

Interactive mode is enabled by default. Use flags to disable it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				// Use default if config doesn't exist
				cfg, err = config.GetDefaultConfig()
				if err != nil {
					return err
				}
			}

			opts := commands.ListOptions{
				Verbose:     verbose,
				ShowConfig:  showConfig,
				Interactive: !noInteractive && !verbose && !showConfig, // Interactive by default unless flags disable it
			}

			return commands.ListProfiles(cfg.ProfilesDir, opts)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information (disables interactive)")
	cmd.Flags().BoolVarP(&showConfig, "config", "c", false, "Show git configuration (disables interactive)")
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Disable interactive mode")

	return cmd
}

