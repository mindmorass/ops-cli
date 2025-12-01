package plugin

import "github.com/spf13/cobra"

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Register registers the plugin's commands with the root command
	Register(rootCmd *cobra.Command) error
}

// PluginMetadata contains metadata about a plugin
type PluginMetadata struct {
	Name        string
	Version     string
	GoVersion   string
	Platform    string
	BuildDate   string
	Description string
}

// PluginSymbol is the exported symbol name that plugins must export
const PluginSymbol = "Plugin"
