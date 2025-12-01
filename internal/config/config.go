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
	Jira            *JiraConfig            `toml:"jira,omitempty"`
	GitHub          *GitHubConfig          `toml:"github,omitempty"`
	Confluence      *ConfluenceConfig      `toml:"confluence,omitempty"`
	NewRelic        *NewRelicConfig        `toml:"newrelic,omitempty"`
	Output          *OutputConfig          `toml:"output,omitempty"`
	DevTools        map[string]interface{} `toml:"development_tools,omitempty"`
	PackageManagers map[string]interface{} `toml:"package_managers,omitempty"`
	Startpage       map[string]interface{} `toml:"startpage,omitempty"`
}

// JiraConfig holds Jira API configuration
type JiraConfig struct {
	BaseURL        string `toml:"base_url"`
	Username       string `toml:"username"`
	AtlassianToken string `toml:"atlassian_token"`
	DefaultProject string `toml:"default_project"` // Jira-specific
}

// GitHubConfig holds GitHub API configuration
type GitHubConfig struct {
	Token        string `toml:"token"`
	DefaultOwner string `toml:"default_owner"`
	APIURL       string `toml:"api_url"`
}

// ConfluenceConfig holds Confluence API configuration
type ConfluenceConfig struct {
	BaseURL        string `toml:"base_url"`
	Username       string `toml:"username"`
	AtlassianToken string `toml:"atlassian_token"`
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

	if updates.Confluence != nil {
		if cm.config.Confluence == nil {
			cm.config.Confluence = &ConfluenceConfig{}
		}
		if updates.Confluence.BaseURL != "" {
			cm.config.Confluence.BaseURL = updates.Confluence.BaseURL
		}
		if updates.Confluence.Username != "" {
			cm.config.Confluence.Username = updates.Confluence.Username
		}
		if updates.Confluence.AtlassianToken != "" {
			cm.config.Confluence.AtlassianToken = updates.Confluence.AtlassianToken
		}
	}

	if updates.NewRelic != nil {
		if cm.config.NewRelic == nil {
			cm.config.NewRelic = &NewRelicConfig{}
		}
		if updates.NewRelic.APIKey != "" {
			cm.config.NewRelic.APIKey = updates.NewRelic.APIKey
		}
		if updates.NewRelic.AccountID != "" {
			cm.config.NewRelic.AccountID = updates.NewRelic.AccountID
		}
		if updates.NewRelic.DefaultQuery != "" {
			cm.config.NewRelic.DefaultQuery = updates.NewRelic.DefaultQuery
		}
		if updates.NewRelic.LogLevel != "" {
			cm.config.NewRelic.LogLevel = updates.NewRelic.LogLevel
		}
		if updates.NewRelic.Region != "" {
			cm.config.NewRelic.Region = updates.NewRelic.Region
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

	// Merge map-based configs (DevTools, PackageManagers, Startpage)
	if updates.DevTools != nil {
		if cm.config.DevTools == nil {
			cm.config.DevTools = make(map[string]interface{})
		}
		for k, v := range updates.DevTools {
			cm.config.DevTools[k] = v
		}
	}

	if updates.PackageManagers != nil {
		if cm.config.PackageManagers == nil {
			cm.config.PackageManagers = make(map[string]interface{})
		}
		for k, v := range updates.PackageManagers {
			cm.config.PackageManagers[k] = v
		}
	}

	if updates.Startpage != nil {
		if cm.config.Startpage == nil {
			cm.config.Startpage = make(map[string]interface{})
		}
		for k, v := range updates.Startpage {
			cm.config.Startpage[k] = v
		}
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

// NewCommandConfigManager creates a config manager for a command-specific config file
func NewCommandConfigManager(commandName string) *ConfigManager {
	appName := getAppName()
	xdg := core.NewXDGPaths(appName)

	configFile, err := xdg.ConfigFile(commandName + ".toml")
	if err != nil {
		// Fallback to old method if XDG fails
		configDir := getConfigDir()
		configFile = filepath.Join(configDir, commandName+".toml")
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

// LoadCommandConfig loads a command-specific configuration
func LoadCommandConfig(commandName string) (*AppConfig, error) {
	manager := NewCommandConfigManager(commandName)
	// Try to read config file
	data, err := os.ReadFile(manager.configFile)
	if err != nil {
		// If file doesn't exist, return empty config (not an error)
		if os.IsNotExist(err) {
			return &AppConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal TOML directly
	if err := toml.Unmarshal(data, manager.config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return manager.Get(), nil
}

// SaveCommandConfig saves a command-specific configuration
func SaveCommandConfig(commandName string, cfg *AppConfig) error {
	manager := NewCommandConfigManager(commandName)
	manager.config = cfg
	return manager.Save()
}

// LoadConfig loads the global configuration from both shared and command-specific files
func LoadConfig() (*AppConfig, error) {
	// Load shared config (main config.toml)
	sharedManager := GetConfigManager()
	if err := sharedManager.Load(); err != nil {
		// If shared config doesn't exist, create default
		if os.IsNotExist(err) {
			if err := sharedManager.CreateDefault(); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return nil, err
		}
	}
	sharedConfig := sharedManager.Get()

	// Merge command-specific configs
	// Load each command config and merge into shared config
	commands := []string{"github", "jira", "confluence", "newrelic"}
	for _, cmd := range commands {
		cmdConfig, err := LoadCommandConfig(cmd)
		if err != nil {
			// Skip if file doesn't exist
			continue
		}

		// Merge command-specific config into shared config
		switch cmd {
		case "github":
			if cmdConfig.GitHub != nil {
				if sharedConfig.GitHub == nil {
					sharedConfig.GitHub = &GitHubConfig{}
				}
				if cmdConfig.GitHub.Token != "" {
					sharedConfig.GitHub.Token = cmdConfig.GitHub.Token
				}
				if cmdConfig.GitHub.DefaultOwner != "" {
					sharedConfig.GitHub.DefaultOwner = cmdConfig.GitHub.DefaultOwner
				}
				if cmdConfig.GitHub.APIURL != "" {
					sharedConfig.GitHub.APIURL = cmdConfig.GitHub.APIURL
				}
			}
		case "jira":
			if cmdConfig.Jira != nil {
				if sharedConfig.Jira == nil {
					sharedConfig.Jira = &JiraConfig{}
				}
				if cmdConfig.Jira.BaseURL != "" {
					sharedConfig.Jira.BaseURL = cmdConfig.Jira.BaseURL
				}
				if cmdConfig.Jira.Username != "" {
					sharedConfig.Jira.Username = cmdConfig.Jira.Username
				}
				if cmdConfig.Jira.AtlassianToken != "" {
					sharedConfig.Jira.AtlassianToken = cmdConfig.Jira.AtlassianToken
				}
				if cmdConfig.Jira.DefaultProject != "" {
					sharedConfig.Jira.DefaultProject = cmdConfig.Jira.DefaultProject
				}
			}
		case "confluence":
			if cmdConfig.Confluence != nil {
				if sharedConfig.Confluence == nil {
					sharedConfig.Confluence = &ConfluenceConfig{}
				}
				if cmdConfig.Confluence.BaseURL != "" {
					sharedConfig.Confluence.BaseURL = cmdConfig.Confluence.BaseURL
				}
				if cmdConfig.Confluence.Username != "" {
					sharedConfig.Confluence.Username = cmdConfig.Confluence.Username
				}
				if cmdConfig.Confluence.AtlassianToken != "" {
					sharedConfig.Confluence.AtlassianToken = cmdConfig.Confluence.AtlassianToken
				}
			}
		case "newrelic":
			if cmdConfig.NewRelic != nil {
				if sharedConfig.NewRelic == nil {
					sharedConfig.NewRelic = &NewRelicConfig{}
				}
				if cmdConfig.NewRelic.APIKey != "" {
					sharedConfig.NewRelic.APIKey = cmdConfig.NewRelic.APIKey
				}
				if cmdConfig.NewRelic.AccountID != "" {
					sharedConfig.NewRelic.AccountID = cmdConfig.NewRelic.AccountID
				}
				if cmdConfig.NewRelic.DefaultQuery != "" {
					sharedConfig.NewRelic.DefaultQuery = cmdConfig.NewRelic.DefaultQuery
				}
				if cmdConfig.NewRelic.LogLevel != "" {
					sharedConfig.NewRelic.LogLevel = cmdConfig.NewRelic.LogLevel
				}
				if cmdConfig.NewRelic.Region != "" {
					sharedConfig.NewRelic.Region = cmdConfig.NewRelic.Region
				}
			}
		}
	}

	return sharedConfig, nil
}

// SaveConfig saves the global configuration by merging with existing config
// This ensures existing data is not lost when updating configuration
// DEPRECATED: Use SaveCommandConfig for command-specific configs instead
func SaveConfig(cfg *AppConfig) error {
	manager := GetConfigManager()
	
	// Load existing config first to preserve all data
	if err := manager.Load(); err != nil {
		// If config doesn't exist, create default first
		if os.IsNotExist(err) {
			if err := manager.CreateDefault(); err != nil {
				return fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
	}
	
	// Merge the updates with existing config
	manager.Update(cfg)
	
	// Save the merged config
	return manager.Save()
}

// SaveGitHubConfig saves GitHub configuration to github.toml
func SaveGitHubConfig(cfg *GitHubConfig) error {
	appCfg := &AppConfig{GitHub: cfg}
	return SaveCommandConfig("github", appCfg)
}

// SaveJiraConfig saves Jira configuration to jira.toml
func SaveJiraConfig(cfg *JiraConfig) error {
	appCfg := &AppConfig{Jira: cfg}
	return SaveCommandConfig("jira", appCfg)
}

// SaveConfluenceConfig saves Confluence configuration to confluence.toml
func SaveConfluenceConfig(cfg *ConfluenceConfig) error {
	appCfg := &AppConfig{Confluence: cfg}
	return SaveCommandConfig("confluence", appCfg)
}

// SaveNewRelicConfig saves NewRelic configuration to newrelic.toml
func SaveNewRelicConfig(cfg *NewRelicConfig) error {
	appCfg := &AppConfig{NewRelic: cfg}
	return SaveCommandConfig("newrelic", appCfg)
}

// GetJiraCredentials returns Jira credentials from Jira config
func (cfg *AppConfig) GetJiraCredentials() (baseURL, username, token string) {
	if cfg.Jira != nil {
		baseURL = cfg.Jira.BaseURL
		username = cfg.Jira.Username
		token = cfg.Jira.AtlassianToken
	}
	return baseURL, username, token
}

// GetConfluenceCredentials returns Confluence credentials from Confluence config
func (cfg *AppConfig) GetConfluenceCredentials() (baseURL, username, token string) {
	if cfg.Confluence != nil {
		baseURL = cfg.Confluence.BaseURL
		username = cfg.Confluence.Username
		token = cfg.Confluence.AtlassianToken
	}
	return baseURL, username, token
}
