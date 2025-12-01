package main

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/api/confluence"
	"github.com/ops-cli/internal/config"
)

// getConfluenceClient creates a Confluence client from config or environment
func getConfluenceClient() (*confluence.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get credentials from Confluence config
	baseURL, username, token := cfg.GetConfluenceCredentials()

	// Environment variables take precedence
	if envURL := os.Getenv("CONFLUENCE_BASE_URL"); envURL != "" {
		baseURL = envURL
	}
	if envUser := os.Getenv("CONFLUENCE_USERNAME"); envUser != "" {
		username = envUser
	}
	if envToken := os.Getenv("ATLASSIAN_TOKEN"); envToken != "" {
		token = envToken
	}

	if baseURL == "" {
		return nil, fmt.Errorf("Confluence base URL not configured. Run 'ops-cli confluence config setup'")
	}
	if username == "" || token == "" {
		return nil, fmt.Errorf("Confluence credentials not configured. Run 'ops-cli confluence config setup'")
	}

	return confluence.NewClient(baseURL, username, token)
}
