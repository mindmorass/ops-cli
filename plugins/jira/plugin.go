package main

import (
	"github.com/spf13/cobra"
)

// Plugin must be exported for plugin loading
var Plugin = &JiraPlugin{}

type JiraPlugin struct{}

func (p *JiraPlugin) Name() string {
	return "jira"
}

func (p *JiraPlugin) Version() string {
	return "1.0.0"
}

func (p *JiraPlugin) Register(rootCmd *cobra.Command) error {
	jiraCmd := &cobra.Command{
		Use:   "jira",
		Short: "Jira API operations and issue management",
		Long: `Jira Commands

Usage: ops-cli jira <subcommand> [options] [args]

The Jira module provides comprehensive Jira issue management capabilities.`,
	}

	// Register subcommands
	jiraCmd.AddCommand(newListCmd())
	jiraCmd.AddCommand(newGetCmd())
	jiraCmd.AddCommand(newCreateCmd())
	jiraCmd.AddCommand(newStaleCmd())
	jiraCmd.AddCommand(newTimeCmd())
	jiraCmd.AddCommand(newConfigCmd())

	rootCmd.AddCommand(jiraCmd)
	return nil
}

