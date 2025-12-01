package main

import (
	"github.com/spf13/cobra"
)

// Plugin must be exported for plugin loading
// Export the concrete type, not the interface
var Plugin = &NewRelicPlugin{}

type NewRelicPlugin struct{}

func (p *NewRelicPlugin) Name() string {
	return "newrelic"
}

func (p *NewRelicPlugin) Version() string {
	return "1.0.0"
}

func (p *NewRelicPlugin) Register(rootCmd *cobra.Command) error {
	newrelicCmd := &cobra.Command{
		Use:   "newrelic",
		Short: "New Relic logs streaming and monitoring",
		Long: `New Relic Commands

Usage: ops-cli newrelic <subcommand> [options] [args]

Manage New Relic logs, entities, and metrics.`,
	}

	newrelicCmd.AddCommand(newConfigCmd())
	newrelicCmd.AddCommand(newLogsCmd())
	newrelicCmd.AddCommand(newEntitiesCmd())
	newrelicCmd.AddCommand(newSummaryCmd())

	rootCmd.AddCommand(newrelicCmd)
	return nil
}
