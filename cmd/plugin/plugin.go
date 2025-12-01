package plugin

import (
	"github.com/spf13/cobra"
)

// Register registers the plugin management commands with the root command
func Register(rootCmd *cobra.Command) {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
		Long: `Manage plugins for ops-cli.

Plugins are dynamically loaded .so files that extend the functionality of ops-cli.
Plugins are installed in the XDG config directory ($XDG_CONFIG_HOME/ops-cli/plugins or ~/.config/ops-cli/plugins).`,
	}

	pluginCmd.AddCommand(newListCmd())
	pluginCmd.AddCommand(newInstallCmd())
	pluginCmd.AddCommand(newUninstallCmd())
	pluginCmd.AddCommand(newUpdateCmd())

	rootCmd.AddCommand(pluginCmd)
}
