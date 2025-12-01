package devtools

import (
	"github.com/spf13/cobra"
)

// Register registers the devtools module with the root command
func Register(rootCmd *cobra.Command) {
	devtoolsCmd := &cobra.Command{
		Use:   "devtools",
		Short: "Development tools management and installation (macOS/Homebrew only)",
		Long: `DevTools Commands (macOS/Homebrew only)

Usage: ops-cli devtools <subcommand> [options] [args]

The DevTools module provides development tools management using Homebrew on macOS.`,
	}

	// Register subcommands
	devtoolsCmd.AddCommand(newCheckCmd())
	devtoolsCmd.AddCommand(newInstallCmd())

	rootCmd.AddCommand(devtoolsCmd)
}
