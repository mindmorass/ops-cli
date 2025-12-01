package github

import (
	"github.com/spf13/cobra"
)

// Register registers the github module with the root command
func Register(rootCmd *cobra.Command) {
	githubCmd := &cobra.Command{
		Use:   "github",
		Short: "GitHub API operations and repository management",
		Long: `GitHub Commands

Usage: ops-cli github <subcommand> [options] [args]

The GitHub module provides comprehensive GitHub repository and issue management capabilities.`,
	}

	githubCmd.AddCommand(newConfigCmd())
	githubCmd.AddCommand(newReposCmd())
	githubCmd.AddCommand(newRepoCmd())
	githubCmd.AddCommand(newSearchCmd())
	githubCmd.AddCommand(newContentsCmd())
	githubCmd.AddCommand(newFileCmd())
	githubCmd.AddCommand(newReleasesCmd())
	githubCmd.AddCommand(newTagsCmd())
	githubCmd.AddCommand(newWorkflowsCmd())
	githubCmd.AddCommand(newRunsCmd())
	githubCmd.AddCommand(newPackagesCmd())

	rootCmd.AddCommand(githubCmd)
}
