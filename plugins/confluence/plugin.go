package main

import (
	"github.com/spf13/cobra"
)

// Plugin must be exported for plugin loading
var Plugin = &ConfluencePlugin{}

type ConfluencePlugin struct{}

func (p *ConfluencePlugin) Name() string {
	return "confluence"
}

func (p *ConfluencePlugin) Version() string {
	return "1.0.0"
}

func (p *ConfluencePlugin) Register(rootCmd *cobra.Command) error {
	confluenceCmd := &cobra.Command{
		Use:   "confluence",
		Short: "Confluence documentation management",
		Long: `Confluence Documentation Commands

Manage Confluence documentation with page operations and API configuration.
Supports getting pages, managing configuration, and more.`,
	}

	confluenceCmd.AddCommand(newConfigCmd())
	confluenceCmd.AddCommand(newGetCmd())
	confluenceCmd.AddCommand(newMarkdownCmd())

	rootCmd.AddCommand(confluenceCmd)
	return nil
}

