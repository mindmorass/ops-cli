package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/ops-cli/internal/profile/ui"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var force bool
	var remote string
	var noInteractive bool

	cmd := &cobra.Command{
		Use:   "sync [command] [profile-name]",
		Short: "Sync operations for profiles",
		Long: `Sync operations for profiles.

Commands: init, pull, push, sync, remote, status`,
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

			syncCommand := args[0]
			profileName := ""
			if len(args) > 1 {
				profileName = args[1]
			}

			opts := commands.GitOptions{
				ProfileName: profileName,
				Remote:      remote,
				Force:       force,
			}

			// If no profile name provided and not --no-interactive, show interactive selection
			if opts.ProfileName == "" && !noInteractive && syncCommand != "status" {
				// Get list of profiles
				entries, err := os.ReadDir(cfg.ProfilesDir)
				if err != nil {
					return fmt.Errorf("failed to read profiles directory: %w", err)
				}

				var profiles []string
				for _, entry := range entries {
					if entry.IsDir() && entry.Name() != ".git" {
						profilePath := filepath.Join(cfg.ProfilesDir, entry.Name())
						envrcPath := filepath.Join(profilePath, ".envrc")
						if _, err := os.Stat(envrcPath); err == nil {
							profiles = append(profiles, entry.Name())
						}
					}
				}

				if len(profiles) == 0 {
					return fmt.Errorf("no profiles found")
				}

				selected, err := ui.SelectProfile(profiles, fmt.Sprintf("Select profile for sync %s:", syncCommand))
				if err != nil {
					return err
				}
				opts.ProfileName = selected
			}

			switch syncCommand {
			case "init":
				return commands.InitGit(cfg.ProfilesDir, opts)
			case "pull":
				return commands.PullGit(cfg.ProfilesDir, opts)
			case "push":
				return commands.PushGit(cfg.ProfilesDir, opts)
			case "sync":
				return commands.SyncGit(cfg.ProfilesDir, opts)
			case "remote":
				// For remote command, the URL might be the last argument
				if opts.Remote == "" && len(args) > 2 {
					opts.Remote = args[2]
				}
				return commands.SetRemote(cfg.ProfilesDir, opts)
			case "status":
				return commands.GetGitStatus(cfg.ProfilesDir, opts)
			default:
				return fmt.Errorf("unknown sync command: %s", syncCommand)
			}
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force push (use with caution)")
	cmd.Flags().StringVar(&remote, "remote", "", "Remote URL")
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Disable interactive profile selection")

	return cmd
}

