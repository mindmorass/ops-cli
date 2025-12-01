package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// Registry manages plugin metadata and state
type Registry struct {
	pluginDir string
}

// NewRegistry creates a new plugin registry
func NewRegistry(pluginDir string) *Registry {
	return &Registry{
		pluginDir: pluginDir,
	}
}

// ListInstalledPlugins returns a list of installed plugin names
func (r *Registry) ListInstalledPlugins() ([]string, error) {
	var plugins []string

	// Ensure plugin directory exists
	if err := os.MkdirAll(r.pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Find all .so files
	pattern := filepath.Join(r.pluginDir, "*.so")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan plugin directory: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		name = name[:len(name)-3] // Remove .so extension
		plugins = append(plugins, name)
	}

	return plugins, nil
}

// IsInstalled checks if a plugin is installed
func (r *Registry) IsInstalled(name string) bool {
	pluginPath := filepath.Join(r.pluginDir, fmt.Sprintf("%s.so", name))
	_, err := os.Stat(pluginPath)
	return !os.IsNotExist(err)
}

// GetPluginPath returns the path to a plugin file
func (r *Registry) GetPluginPath(name string) string {
	return filepath.Join(r.pluginDir, fmt.Sprintf("%s.so", name))
}
