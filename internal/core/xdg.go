package core

import (
	"os"
	"path/filepath"
)

// XDGPaths provides XDG Base Directory Specification compliant paths
type XDGPaths struct {
	appName string
}

// NewXDGPaths creates a new XDG paths helper
func NewXDGPaths(appName string) *XDGPaths {
	return &XDGPaths{appName: appName}
}

// ConfigDir returns the XDG config directory
// Uses WORKSPACE_HOME if set, then XDG_CONFIG_HOME if set, otherwise defaults to $HOME/.config
func (x *XDGPaths) ConfigDir() (string, error) {
	// Check WORKSPACE_HOME first (for workspace profiles)
	// This matches the Makefile behavior
	if workspaceHome := os.Getenv("WORKSPACE_HOME"); workspaceHome != "" {
		return filepath.Join(workspaceHome, ".config", x.appName), nil
	}

	// Then check XDG_CONFIG_HOME
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, x.appName), nil
	}

	// Default to $HOME/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", x.appName), nil
}

// DataDir returns the XDG data directory
// Uses XDG_DATA_HOME if set, otherwise defaults to $HOME/.local/share
func (x *XDGPaths) DataDir() (string, error) {
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, x.appName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", x.appName), nil
}

// CacheDir returns the XDG cache directory
// Uses XDG_CACHE_HOME if set, otherwise defaults to $HOME/.cache
func (x *XDGPaths) CacheDir() (string, error) {
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, x.appName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".cache", x.appName), nil
}

// ConfigFile returns the path to the config file
func (x *XDGPaths) ConfigFile(filename string) (string, error) {
	configDir, err := x.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, filename), nil
}

// PluginDir returns the directory for plugins (uses config directory)
func (x *XDGPaths) PluginDir() (string, error) {
	configDir, err := x.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "plugins"), nil
}
