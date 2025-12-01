package devtools

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ops-cli/internal/config"
)

// DevelopmentTool represents a development tool configuration
type DevelopmentTool struct {
	Name        string
	Homebrew    string
	HomebrewTap string // Optional Homebrew tap (e.g., "kovetskiy/mark")
	Description string
}

// LoadDevToolsConfig loads development tools from config.toml
func LoadDevToolsConfig() (map[string]DevelopmentTool, error) {
	tools := make(map[string]DevelopmentTool)

	// Try loading from project config.toml first (more reliable)
	projectTools, err := loadFromProjectConfig()
	if err == nil && len(projectTools) > 0 {
		return projectTools, nil
	}

	// Fallback: try to get from config manager
	manager := config.GetConfigManager()
	if devTools := manager.GetValue("development_tools"); devTools != nil {
		if devToolsMap, ok := devTools.(map[string]interface{}); ok {
			for key, value := range devToolsMap {
				if toolMap, ok := value.(map[string]interface{}); ok {
					tool := DevelopmentTool{
						Name:        getString(toolMap, "name"),
						Homebrew:    getString(toolMap, "homebrew"),
						HomebrewTap: getString(toolMap, "homebrew_tap"),
						Description: getString(toolMap, "description"),
					}
					if tool.Homebrew != "" {
						tools[key] = tool
					}
				}
			}
		}
	}

	return tools, nil
}

// loadFromProjectConfig loads tools from project's config.toml
func loadFromProjectConfig() (map[string]DevelopmentTool, error) {
	// Try to find config.toml in current directory or parent directories
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(cwd, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try parent directory
		configPath = filepath.Join(filepath.Dir(cwd), "config.toml")
	}

	// Read and parse config.toml manually for development_tools section
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	return parseDevToolsFromTOML(string(data))
}

// parseDevToolsFromTOML parses development_tools section from TOML
func parseDevToolsFromTOML(content string) (map[string]DevelopmentTool, error) {
	tools := make(map[string]DevelopmentTool)
	lines := strings.Split(content, "\n")

	inDevToolsSection := false
	currentKey := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for section header
		if trimmed == "[development_tools]" {
			inDevToolsSection = true
			continue
		}

		// If we hit another section, stop
		if strings.HasPrefix(trimmed, "[") && trimmed != "[development_tools]" {
			inDevToolsSection = false
			continue
		}

		if !inDevToolsSection {
			continue
		}

		// Parse key = value
		if idx := strings.Index(trimmed, " = "); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			value := strings.TrimSpace(trimmed[idx+3:])

			// Check if this is a tool definition (inline table)
			if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
				currentKey = key
				tool := parseToolInlineTable(value)
				if tool.Homebrew != "" {
					tools[currentKey] = tool
				}
			}
		}
	}

	return tools, nil
}

// parseToolInlineTable parses an inline TOML table for a tool
func parseToolInlineTable(value string) DevelopmentTool {
	tool := DevelopmentTool{}

	// Remove braces
	content := strings.Trim(value, "{}")

	// Split by comma
	parts := strings.Split(content, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, " = "); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			val := strings.TrimSpace(part[idx+3:])

			// Remove quotes
			val = strings.Trim(val, "\"'")

			switch key {
			case "name":
				tool.Name = val
			case "homebrew":
				tool.Homebrew = val
			case "homebrew_tap":
				tool.HomebrewTap = val
			case "description":
				tool.Description = val
			}
		}
	}

	return tool
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
