package main

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/api/newrelic"
	"github.com/ops-cli/internal/config"
)

// getNewRelicClient creates a New Relic client from config or environment
func getNewRelicClient() (*newrelic.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.NewRelic == nil || cfg.NewRelic.APIKey == "" {
		return nil, fmt.Errorf("New Relic API key not configured. Run 'ops-cli newrelic config setup'")
	}

	// Get credentials from config or environment
	apiKey := cfg.NewRelic.APIKey
	accountID := cfg.NewRelic.AccountID
	region := cfg.NewRelic.Region
	if region == "" {
		region = "US"
	}

	// Environment variables take precedence
	if envKey := os.Getenv("NEW_RELIC_API_KEY"); envKey != "" {
		apiKey = envKey
	}
	if envAccountID := os.Getenv("NEW_RELIC_ACCOUNT_ID"); envAccountID != "" {
		accountID = envAccountID
	}
	if envRegion := os.Getenv("NEW_RELIC_REGION"); envRegion != "" {
		region = envRegion
	}

	if accountID == "" {
		return nil, fmt.Errorf("New Relic account ID not configured. Run 'ops-cli newrelic config setup'")
	}

	return newrelic.NewClient(apiKey, accountID, region), nil
}
