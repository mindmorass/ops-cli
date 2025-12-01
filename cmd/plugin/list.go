package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ops-cli/internal/plugin"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available and installed plugins",
		Long: `List all available plugins (source code) and installed plugins (.so files).

Available plugins are those with source code in the plugins/ directory.
Installed plugins are compiled .so files in the plugin directory.`,
		RunE: runList,
	}

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	loader, err := plugin.NewLoader()
	if err != nil {
		return fmt.Errorf("failed to create plugin loader: %w", err)
	}

	// List available plugins (source code in plugins/ directory)
	fmt.Println("Available plugins:")
	pluginsDir := "plugins"
	availablePlugins := []string{}

	// Check if plugins directory exists
	if entries, err := os.ReadDir(pluginsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				pluginPath := filepath.Join(pluginsDir, entry.Name(), "plugin.go")
				if _, err := os.Stat(pluginPath); err == nil {
					availablePlugins = append(availablePlugins, entry.Name())
					fmt.Printf("  • %s\n", entry.Name())
				}
			}
		}
	}

	if len(availablePlugins) == 0 {
		fmt.Println("  (none found in plugins/ directory)")
	}

	// List installed plugins
	fmt.Println("\nInstalled plugins:")
	registry := plugin.NewRegistry(loader.GetPluginDir())
	installedPlugins, err := registry.ListInstalledPlugins()
	if err != nil {
		return fmt.Errorf("failed to list installed plugins: %w", err)
	}

	if len(installedPlugins) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, name := range installedPlugins {
			// Try to load plugin to get version info
			p, err := loader.LoadPlugin(name)
			if err == nil {
				fmt.Printf("  ✓ %s (v%s)\n", p.Name(), p.Version())
			} else {
				fmt.Printf("  ✓ %s (error loading: %v)\n", name, err)
			}
		}
	}

	return nil
}
