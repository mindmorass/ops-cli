package profile

import (
	"github.com/spf13/cobra"
)

// Register registers the profile module with the root command
func Register(rootCmd *cobra.Command) {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage workspace profiles with direnv",
		Long: `Profile Commands

Manage workspace profiles with direnv for environment-specific configurations.
Each profile can have its own git config, SSH keys, AWS credentials, and more.`,
	}

	profileCmd.AddCommand(newInitCmd())
	profileCmd.AddCommand(newCreateCmd())
	profileCmd.AddCommand(newListCmd())
	profileCmd.AddCommand(newSelectCmd())
	profileCmd.AddCommand(newDeleteCmd())
	profileCmd.AddCommand(newUpdateCmd())
	profileCmd.AddCommand(newInfoCmd())
	profileCmd.AddCommand(newStatusCmd())
	profileCmd.AddCommand(newSyncCmd())
	profileCmd.AddCommand(newDotfilesCmd())

	rootCmd.AddCommand(profileCmd)
}

