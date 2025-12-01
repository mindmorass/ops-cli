package profile

import (
	"github.com/ops-cli/internal/profile/commands"
	"github.com/ops-cli/internal/profile/config"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var template string
	var gitName string
	var gitEmail string
	var force bool
	var interactive bool
	var dryRun bool
	var initGit bool
	var gitRemote string

	cmd := &cobra.Command{
		Use:   "create [profile-name]",
		Short: "Create a new workspace profile",
		Long: `Create a new workspace profile with direnv configuration.

Creates a profile directory with .envrc, .gitconfig, SSH config, and other configuration files.
Templates: basic, personal, work, client`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				// Use default if config doesn't exist
				cfg, err = config.GetDefaultConfig()
				if err != nil {
					return err
				}
			}

			opts := commands.CreateOptions{
				ProfileName: args[0],
				Template:    template,
				GitName:     gitName,
				GitEmail:    gitEmail,
				Force:       force,
				Interactive: interactive,
				DryRun:      dryRun,
				InitGit:     initGit,
				GitRemote:   gitRemote,
			}

			// If no non-interactive flags provided, enable interactive mode
			hasNonInteractiveFlags := force || template != "basic" || gitName != "" || gitEmail != "" || dryRun || initGit || gitRemote != ""
			if !hasNonInteractiveFlags {
				opts.Interactive = true
			}

			return commands.CreateProfile(cfg.ProfilesDir, opts)
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "basic", "Use template: personal, work, client, basic")
	cmd.Flags().StringVar(&gitName, "git-name", "", "Set git user.name in .gitconfig")
	cmd.Flags().StringVar(&gitEmail, "git-email", "", "Set git user.email in .gitconfig")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing profile if it exists")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Prompt for all configuration values")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without creating it")
	cmd.Flags().BoolVar(&initGit, "init-git", false, "Initialize git repository after creation")
	cmd.Flags().StringVar(&gitRemote, "git-remote", "", "Initialize git repository with remote URL")

	return cmd
}

