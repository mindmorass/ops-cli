package profile

import (
	"fmt"

	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newDotfilesCmd() *cobra.Command {
	var profileName string
	var fileName string
	var editor string

	cmd := &cobra.Command{
		Use:   "dotfiles [command]",
		Short: "Manage profile dotfiles",
		Long: `Manage dotfiles in workspace profiles.

Commands: list, edit`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				// Use default if config doesn't exist
				cfg, err = config.GetDefaultConfig()
				if err != nil {
					return err
				}
			}

			subcommand := args[0]

			opts := commands.DotfilesOptions{
				ProfileName: profileName,
				FileName:    fileName,
				Editor:      editor,
			}

			switch subcommand {
			case "list", "ls":
				return commands.ListDotfiles(cfg.ProfilesDir, opts)
			case "edit", "e":
				return commands.EditDotfile(cfg.ProfilesDir, opts)
			default:
				return fmt.Errorf("unknown dotfiles command: %s", subcommand)
			}
		},
	}

	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Profile name (interactive selection if omitted)")
	cmd.Flags().StringVarP(&fileName, "file", "f", "", "File name (interactive selection if omitted)")
	cmd.Flags().StringVarP(&editor, "editor", "e", "", "Editor to use (default: $EDITOR, $VISUAL, or vim)")

	return cmd
}

