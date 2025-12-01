package github

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/api/github"
	"github.com/ops-cli/internal/config"
)

// getGitHubClient creates a GitHub client from config or environment
func getGitHubClient() (*github.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get token from config or environment
	token := ""
	if cfg.GitHub != nil && cfg.GitHub.Token != "" {
		token = cfg.GitHub.Token
	}

	// Environment variable takes precedence
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		token = envToken
	}

	if token == "" {
		return nil, fmt.Errorf("GitHub token not configured. Run 'ops-cli github config setup'")
	}

	return github.NewClient(token), nil
}
