package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ops-cli/internal/core"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

// AppConfig represents the application configuration structure
type AppConfig struct {
	Version         string                 `toml:"version"`
	Atlassian       *AtlassianConfig       `toml:"atlassian,omitempty"`
	Jira            *JiraConfig            `toml:"jira,omitempty"`
	GitHub          *GitHubConfig          `toml:"github,omitempty"`
	Confluence      *ConfluenceConfig      `toml:"confluence,omitempty"`
	NewRelic        *NewRelicConfig        `toml:"newrelic,omitempty"`
	Output          *OutputConfig          `toml:"output,omitempty"`
	DevTools        map[string]interface{} `toml:"development_tools,omitempty"`
	PackageManagers map[string]interface{} `toml:"package_managers,omitempty"`
	Startpage       map[string]interface{} `toml:"startpage,omitempty"`
}

// AtlassianConfig holds shared Atlassian API configuration
// Used by both Jira and Confluence
type AtlassianConfig struct {
	BaseURL        string `toml:"base_url"`
	Username       string `toml:"username"`
	AtlassianToken string `toml:"atlassian_token"`
}

// JiraConfig holds Jira API configuration
// If Atlassian config is set, these fields override the shared config
type JiraConfig struct {
	BaseURL        string `toml:"base_url"`        // Overrides Atlassian base_url if set
	Username       string `toml:"username"`        // Overrides Atlassian username if set
	AtlassianToken string `toml:"atlassian_token"` // Overrides Atlassian token if set
	DefaultProject string `toml:"default_project"` // Jira-specific
}

// GitHubConfig holds GitHub API configuration
type GitHubConfig struct {
	Token        string `toml:"token"`
	DefaultOwner string `toml:"default_owner"`
	APIURL       string `toml:"api_url"`
}

// ConfluenceConfig holds Confluence API configuration
// If Atlassian config is set, these fields override the shared config
type ConfluenceConfig struct {
	BaseURL        string `toml:"base_url"`        // Overrides Atlassian base_url if set
	Username       string `toml:"username"`        // Overrides Atlassian username if set
	AtlassianToken string `toml:"atlassian_token"` // Overrides Atlassian token if set
}

// NewRelicConfig holds New Relic API configuration
type NewRelicConfig struct {
	APIKey       string `toml:"api_key"`
	AccountID    string `toml:"account_id"`
	DefaultQuery string `toml:"default_query,omitempty"`
	LogLevel     string `toml:"log_level,omitempty"`
	Region       string `toml:"region,omitempty"`
}

// OutputConfig holds output formatting configuration
type OutputConfig struct {
	Format  string `toml:"format"` // table, json, yaml
	NoColor bool   `toml:"no_color"`
	Verbose bool   `toml:"verbose"`
}

// ConfigManager manages application configuration
type ConfigManager struct {
	configFile string
	config     *AppConfig
	appName    string
	viper      *viper.Viper
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	appName := getAppName()
	xdg := core.NewXDGPaths(appName)

	configFile, err := xdg.ConfigFile("config.toml")
	if err != nil {
		// Fallback to old method if XDG fails
		configDir := getConfigDir()
		configFile = filepath.Join(configDir, "config.toml")
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(configFile)

	// Set environment variable support
	v.SetEnvPrefix("OPS_CLI")
	v.AutomaticEnv()

	return &ConfigManager{
		configFile: configFile,
		appName:    appName,
		viper:      v,
		config:     &AppConfig{},
	}
}

// getAppName determines the application name for config directory
func getAppName() string {
	if name := os.Getenv("OPS_CLI_APP_NAME"); name != "" {
		return name
	}

	execPath, err := os.Executable()
	if err != nil {
		return "ops-cli"
	}

	execName := filepath.Base(execPath)
	if execName == "ops-cli" || execName == "ops-cli.exe" {
		return "ops-cli"
	}

	// Remove .exe extension on Windows
	if runtime.GOOS == "windows" {
		execName = execName[:len(execName)-4]
	}

	return execName
}

// getConfigDir returns the configuration directory path (fallback method)
// This is kept for backward compatibility, but NewConfigManager uses XDGPaths
func getConfigDir() string {
	xdg := core.NewXDGPaths(getAppName())
	configDir, err := xdg.ConfigDir()
	if err != nil {
		// Last resort: use current directory
		return filepath.Join(".", ".config", getAppName())
	}
	return configDir
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	// Try to read config file
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		// If file doesn't exist, create default
		if os.IsNotExist(err) {
			return cm.CreateDefault()
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal TOML directly using go-toml (more reliable than Viper for nested structs)
	if err := toml.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Also read into Viper for environment variable support and GetValue/SetValue
	if err := cm.viper.ReadInConfig(); err != nil {
		// Ignore errors here - we already loaded the config
	}

	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	// Ensure directory exists
	configDir := filepath.Dir(cm.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cm.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *AppConfig {
	return cm.config
}

// Update updates configuration with provided values
func (cm *ConfigManager) Update(updates *AppConfig) {
	if updates.Jira != nil {
		if cm.config.Jira == nil {
			cm.config.Jira = &JiraConfig{}
		}
		if updates.Jira.BaseURL != "" {
			cm.config.Jira.BaseURL = updates.Jira.BaseURL
		}
		if updates.Jira.Username != "" {
			cm.config.Jira.Username = updates.Jira.Username
		}
		if updates.Jira.AtlassianToken != "" {
			cm.config.Jira.AtlassianToken = updates.Jira.AtlassianToken
		}
		if updates.Jira.DefaultProject != "" {
			cm.config.Jira.DefaultProject = updates.Jira.DefaultProject
		}
	}

	if updates.GitHub != nil {
		if cm.config.GitHub == nil {
			cm.config.GitHub = &GitHubConfig{}
		}
		if updates.GitHub.Token != "" {
			cm.config.GitHub.Token = updates.GitHub.Token
		}
		if updates.GitHub.DefaultOwner != "" {
			cm.config.GitHub.DefaultOwner = updates.GitHub.DefaultOwner
		}
		if updates.GitHub.APIURL != "" {
			cm.config.GitHub.APIURL = updates.GitHub.APIURL
		}
	}

	if updates.Atlassian != nil {
		if cm.config.Atlassian == nil {
			cm.config.Atlassian = &AtlassianConfig{}
		}
		if updates.Atlassian.BaseURL != "" {
			cm.config.Atlassian.BaseURL = updates.Atlassian.BaseURL
		}
		if updates.Atlassian.Username != "" {
			cm.config.Atlassian.Username = updates.Atlassian.Username
		}
		if updates.Atlassian.AtlassianToken != "" {
			cm.config.Atlassian.AtlassianToken = updates.Atlassian.AtlassianToken
		}
	}

	if updates.Output != nil {
		if cm.config.Output == nil {
			cm.config.Output = &OutputConfig{}
		}
		if updates.Output.Format != "" {
			cm.config.Output.Format = updates.Output.Format
		}
		cm.config.Output.NoColor = updates.Output.NoColor
		cm.config.Output.Verbose = updates.Output.Verbose
	}
}

// GetValue gets a configuration value by key path
func (cm *ConfigManager) GetValue(key string) interface{} {
	return cm.viper.Get(key)
}

// SetValue sets a configuration value by key path
func (cm *ConfigManager) SetValue(key string, value interface{}) {
	cm.viper.Set(key, value)
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configFile
}

// Exists checks if the configuration file exists
func (cm *ConfigManager) Exists() bool {
	_, err := os.Stat(cm.configFile)
	return !os.IsNotExist(err)
}

// CreateDefault creates a default configuration file
func (cm *ConfigManager) CreateDefault() error {
	cm.config = &AppConfig{
		Version: "1.0.0",
		Atlassian: &AtlassianConfig{
			BaseURL:        "",
			Username:       "",
			AtlassianToken: "",
		},
		Jira: &JiraConfig{
			DefaultProject: "",
		},
		GitHub: &GitHubConfig{
			Token:        "",
			DefaultOwner: "",
			APIURL:       "https://api.github.com",
		},
		Output: &OutputConfig{
			Format:  "table",
			NoColor: false,
			Verbose: false,
		},
	}

	return cm.Save()
}

// Global config manager instance
var globalConfigManager *ConfigManager

// GetConfigManager returns the global configuration manager instance
func GetConfigManager() *ConfigManager {
	if globalConfigManager == nil {
		globalConfigManager = NewConfigManager()
	}
	return globalConfigManager
}

// LoadConfig loads the global configuration
func LoadConfig() (*AppConfig, error) {
	manager := GetConfigManager()
	if err := manager.Load(); err != nil {
		return nil, err
	}
	return manager.Get(), nil
}

// SaveConfig saves the global configuration
func SaveConfig(cfg *AppConfig) error {
	manager := GetConfigManager()
	manager.config = cfg
	return manager.Save()
}

// GetAtlassianCredentials returns Atlassian credentials with fallback logic
// Checks individual config first, then falls back to shared Atlassian config
func (cfg *AppConfig) GetAtlassianCredentials() (baseURL, username, token string) {
	// Try to get from individual configs first (for backward compatibility)
	// Then fall back to shared Atlassian config

	// For Jira
	if cfg.Jira != nil {
		if cfg.Jira.BaseURL != "" {
			baseURL = cfg.Jira.BaseURL
		}
		if cfg.Jira.Username != "" {
			username = cfg.Jira.Username
		}
		if cfg.Jira.AtlassianToken != "" {
			token = cfg.Jira.AtlassianToken
		}
	}

	// For Confluence
	if cfg.Confluence != nil {
		if baseURL == "" && cfg.Confluence.BaseURL != "" {
			baseURL = cfg.Confluence.BaseURL
		}
		if username == "" && cfg.Confluence.Username != "" {
			username = cfg.Confluence.Username
		}
		if token == "" && cfg.Confluence.AtlassianToken != "" {
			token = cfg.Confluence.AtlassianToken
		}
	}

	// Fall back to shared Atlassian config
	if cfg.Atlassian != nil {
		if baseURL == "" {
			baseURL = cfg.Atlassian.BaseURL
		}
		if username == "" {
			username = cfg.Atlassian.Username
		}
		if token == "" {
			token = cfg.Atlassian.AtlassianToken
		}
	}

	return baseURL, username, token
}

// GetJiraCredentials returns Jira credentials with fallback to Atlassian config
func (cfg *AppConfig) GetJiraCredentials() (baseURL, username, token string) {
	// Check Jira-specific config first
	if cfg.Jira != nil {
		baseURL = cfg.Jira.BaseURL
		username = cfg.Jira.Username
		token = cfg.Jira.AtlassianToken
	}

	// Fall back to shared Atlassian config
	if cfg.Atlassian != nil {
		if baseURL == "" {
			baseURL = cfg.Atlassian.BaseURL
		}
		if username == "" {
			username = cfg.Atlassian.Username
		}
		if token == "" {
			token = cfg.Atlassian.AtlassianToken
		}
	}

	return baseURL, username, token
}

// GetConfluenceCredentials returns Confluence credentials with fallback to Atlassian config
func (cfg *AppConfig) GetConfluenceCredentials() (baseURL, username, token string) {
	// Check Confluence-specific config first
	if cfg.Confluence != nil {
		baseURL = cfg.Confluence.BaseURL
		username = cfg.Confluence.Username
		token = cfg.Confluence.AtlassianToken
	}

	// Fall back to shared Atlassian config
	if cfg.Atlassian != nil {
		if baseURL == "" {
			baseURL = cfg.Atlassian.BaseURL
		}
		if username == "" {
			username = cfg.Atlassian.Username
		}
		if token == "" {
			token = cfg.Atlassian.AtlassianToken
		}
	}

	return baseURL, username, token
}
