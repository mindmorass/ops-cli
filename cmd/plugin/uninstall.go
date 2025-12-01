package plugin

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/plugin"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uninstall <plugin-name>",
		Short:   "Uninstall a plugin",
		Long:    `Uninstall a plugin by removing its .so file from the plugin directory.`,
		Args:    cobra.ExactArgs(1),
		RunE:    runUninstall,
		Aliases: []string{"remove"},
	}

	return cmd
}

func runUninstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	loader, err := plugin.NewLoader()
	if err != nil {
		return fmt.Errorf("failed to create plugin loader: %w", err)
	}

	registry := plugin.NewRegistry(loader.GetPluginDir())

	// Check if plugin is installed
	if !registry.IsInstalled(pluginName) {
		return fmt.Errorf("plugin %s is not installed", pluginName)
	}

	// Remove plugin file
	pluginPath := registry.GetPluginPath(pluginName)
	if err := os.Remove(pluginPath); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}

	fmt.Printf("✓ Plugin %s uninstalled successfully\n", pluginName)
	return nil
}
