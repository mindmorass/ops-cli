package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ops-cli/internal/plugin"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <plugin-name>",
		Short: "Install a plugin by compiling it to .so",
		Long: `Install a plugin by compiling its source code to a .so file.

The plugin source must exist in the plugins/<plugin-name>/plugin.go file.
The compiled .so file will be installed to the plugin directory.`,
		Args: cobra.ExactArgs(1),
		RunE: runInstall,
	}

	cmd.Flags().Bool("force", false, "Force reinstall even if plugin already exists")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	loader, err := plugin.NewLoader()
	if err != nil {
		return fmt.Errorf("failed to create plugin loader: %w", err)
	}

	registry := plugin.NewRegistry(loader.GetPluginDir())

	// Check if plugin already installed
	if !force && registry.IsInstalled(pluginName) {
		return fmt.Errorf("plugin %s is already installed. Use --force to reinstall", pluginName)
	}

	// Check if plugin source exists
	pluginPath := filepath.Join("plugins", pluginName)
	pluginMain := filepath.Join(pluginPath, "plugin.go")

	if _, err := os.Stat(pluginMain); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found in plugins/%s/plugin.go", pluginName, pluginName)
	}

	// Find repo root (directory containing go.mod)
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Build plugin
	stopSpinner := ui.StartSpinner(fmt.Sprintf("Building plugin %s...", pluginName))
	defer stopSpinner()

	outputPath := registry.GetPluginPath(pluginName)

	// Ensure plugin directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	buildCmd := exec.Command("go", "build",
		"-buildmode=plugin",
		"-o", outputPath,
		pluginMain,
	)

	buildCmd.Dir = repoRoot
	buildCmd.Env = os.Environ()

	// macOS requires CGO for plugins
	buildCmd.Env = append(buildCmd.Env, "CGO_ENABLED=1")

	output, err := buildCmd.CombinedOutput()
	stopSpinner()

	if err != nil {
		return fmt.Errorf("failed to build plugin: %w\n%s", err, output)
	}

	fmt.Printf("✓ Plugin %s installed successfully to %s\n", pluginName, outputPath)
	return nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found (not in a Go module)")
}
