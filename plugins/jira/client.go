package main

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/api/jira"
	"github.com/ops-cli/internal/config"
)

// getJiraClient creates a Jira client from config or environment
func getJiraClient() (*jira.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get credentials from Jira config
	baseURL, username, token := cfg.GetJiraCredentials()

	// Environment variables take precedence
	if envURL := os.Getenv("JIRA_BASE_URL"); envURL != "" {
		baseURL = envURL
	}
	if envUser := os.Getenv("JIRA_USERNAME"); envUser != "" {
		username = envUser
	}
	if envToken := os.Getenv("ATLASSIAN_TOKEN"); envToken != "" {
		token = envToken
	}

	if baseURL == "" {
		return nil, fmt.Errorf("Jira base URL not configured. Run 'ops-cli jira config setup'")
	}
	if username == "" || token == "" {
		return nil, fmt.Errorf("Jira credentials not configured. Run 'ops-cli jira config setup'")
	}

	return jira.NewClient(baseURL, username, token)
}
