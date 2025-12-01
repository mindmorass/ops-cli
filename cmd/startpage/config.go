package startpage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// getStartpageDirectory returns the startpage directory path
func getStartpageDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, "startpage"), nil
}

// getConfigPath returns the path to the startpage config file
func getConfigPath() (string, error) {
	dir, err := getStartpageDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "startpage-config.json"), nil
}

// ensureStartpageDirectory ensures the startpage directory exists
func ensureStartpageDirectory() error {
	dir, err := getStartpageDirectory()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// loadConfig loads the startpage configuration
func loadConfig() (*StartpageConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config StartpageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// saveConfig saves the startpage configuration
func saveConfig(config *StartpageConfig) error {
	if err := ensureStartpageDirectory(); err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

// expandTilde expands ~ to home directory
func expandTilde(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if len(path) == 1 {
			return homeDir, nil
		}
		if path[1] == '/' || path[1] == '\\' {
			return filepath.Join(homeDir, path[2:]), nil
		}
	}
	return path, nil
}
