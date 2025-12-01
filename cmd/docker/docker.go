package docker

import (
	"github.com/spf13/cobra"
)

// Register registers the docker module with the root command
func Register(rootCmd *cobra.Command) {
	dockerCmd := &cobra.Command{
		Use:   "docker",
		Short: "Docker container and image management",
		Long: `Docker Commands

Usage: ops-cli docker <subcommand> [options] [args]

The Docker module provides comprehensive Docker container and image management capabilities.`,
	}

	// Register subcommands
	dockerCmd.AddCommand(newInfoCmd())
	dockerCmd.AddCommand(newPsCmd())
	dockerCmd.AddCommand(newImagesCmd())
	dockerCmd.AddCommand(newStatsCmd())
	dockerCmd.AddCommand(newLogsCmd())
	dockerCmd.AddCommand(newCleanupCmd())
	dockerCmd.AddCommand(newStartCmd())
	dockerCmd.AddCommand(newStopCmd())

	rootCmd.AddCommand(dockerCmd)
}
