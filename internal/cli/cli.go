package cli

import (
	"github.com/ops-cli/cmd/atlassian"
	"github.com/ops-cli/cmd/devtools"
	"github.com/ops-cli/cmd/docker"
	"github.com/ops-cli/cmd/github"
	"github.com/ops-cli/cmd/profile"
	pluginCmd "github.com/ops-cli/cmd/plugin"
	"github.com/ops-cli/cmd/startpage"
	"github.com/ops-cli/internal/plugin"
	"github.com/spf13/cobra"
)

// RegisterModules registers all CLI modules with the root command
func RegisterModules(rootCmd *cobra.Command) {
	// Register all core modules
	atlassian.Register(rootCmd)
	github.Register(rootCmd)
	startpage.Register(rootCmd)
	devtools.Register(rootCmd)
	docker.Register(rootCmd)
	profile.Register(rootCmd)

	// Register plugin management commands
	pluginCmd.Register(rootCmd)

	// Load and register plugins dynamically
	loader, err := plugin.NewLoader()
	if err == nil {
		// Ignore errors - plugins are optional
		loader.RegisterPlugins(rootCmd)
	}
}
