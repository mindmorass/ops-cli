package main

import (
	"github.com/spf13/cobra"
)

// Plugin must be exported for plugin loading
var Plugin = &TeamsPlugin{}

type TeamsPlugin struct{}

func (p *TeamsPlugin) Name() string {
	return "teams"
}

func (p *TeamsPlugin) Version() string {
	return "1.0.0"
}

func (p *TeamsPlugin) Register(rootCmd *cobra.Command) error {
	teamsCmd := &cobra.Command{
		Use:   "teams",
		Short: "Microsoft Teams integration",
		Long: `Microsoft Teams Integration

Usage: ops-cli teams <subcommand> [options] [args]

The Teams module provides Microsoft Teams integration capabilities via Microsoft Graph API:

• Authentication & Configuration
  - Support for Graph Explorer tokens
  - Token validation and management
  - Interactive authentication setup

• Team & Channel Management
  - List your Microsoft Teams
  - Browse channels within teams
  - Create new channels

• Messaging & Communication
  - Send text messages to Teams channels
  - Send rich HTML formatted messages
  - Send Adaptive Cards for notifications

• Profile & User Information
  - View your Teams profile
  - Check authentication status

Setup:
  Before using Teams features, you need to authenticate with Microsoft Graph API.
  You can get an access token from Graph Explorer (developer.microsoft.com/graph/graph-explorer)
  or use the interactive device code flow.`,
	}

	// Register subcommands
	registerTeamsCommands(teamsCmd)

	rootCmd.AddCommand(teamsCmd)
	return nil
}

