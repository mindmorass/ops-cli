package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ops-cli/internal/api/newrelic"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage New Relic configuration",
		Long: `Manage New Relic API configuration.

Subcommands:
  setup    Set up New Relic API credentials interactively
  show     Display current configuration
  test     Test connection to New Relic API`,
	}

	cmd.AddCommand(newConfigSetupCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigTestCmd())

	return cmd
}

func newConfigSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up New Relic API credentials interactively",
		RunE:  runConfigSetup,
	}
}

func runConfigSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Setting up New Relic API configuration...")
	fmt.Println()

	// Prompt for API key
	apiKey := ""
	apiKeyPrompt := &survey.Password{
		Message: "Enter your New Relic API key:",
	}
	if err := survey.AskOne(apiKeyPrompt, &apiKey, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("API key is required")
	}

	// Prompt for account ID
	accountID := ""
	accountIDPrompt := &survey.Input{
		Message: "Enter your New Relic account ID:",
	}
	accountIDValidator := func(ans interface{}) error {
		str := ans.(string)
		if str == "" {
			return fmt.Errorf("account ID is required")
		}
		matched, _ := regexp.MatchString(`^\d+$`, str)
		if !matched {
			return fmt.Errorf("account ID must be numeric")
		}
		return nil
	}
	if err := survey.AskOne(accountIDPrompt, &accountID, survey.WithValidator(accountIDValidator)); err != nil {
		return fmt.Errorf("valid numeric account ID is required")
	}

	// Prompt for default query (optional)
	defaultQuery := ""
	defaultQueryPrompt := &survey.Input{
		Message: "Default NRQL query (optional):",
		Default: "FROM Log SELECT *",
	}
	survey.AskOne(defaultQueryPrompt, &defaultQuery)

	if defaultQuery == "" {
		defaultQuery = "FROM Log SELECT *"
	}

	// Prompt for region (optional)
	region := ""
	regionPrompt := &survey.Select{
		Message: "Region:",
		Options: []string{"US", "EU"},
		Default: "US",
	}
	survey.AskOne(regionPrompt, &region)

	// Test the configuration
	fmt.Println("\nTesting connection...")
	client := newrelic.NewClient(apiKey, accountID, region)

	stopSpinner := ui.StartSpinner("Testing connection...")
	ok, err := client.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Test with a simple query
	testQueryOptions := newrelic.LogsQueryOptions{
		Query: "FROM Log SELECT count(*) SINCE 1 hour ago",
		Limit: 1,
	}

	stopSpinner = ui.StartSpinner("Testing query...")
	result, err := client.QueryLogs(testQueryOptions)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("query test failed: %w", err)
	}

	fmt.Printf("Connection successful! Account ID: %s\n", accountID)
	if len(result.Results) > 0 {
		fmt.Printf("Test query returned: %d result(s)\n", len(result.Results))
	}

	// Save configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.NewRelic == nil {
		cfg.NewRelic = &config.NewRelicConfig{}
	}

	cfg.NewRelic.APIKey = apiKey
	cfg.NewRelic.AccountID = accountID
	cfg.NewRelic.DefaultQuery = defaultQuery
	cfg.NewRelic.LogLevel = "info"
	cfg.NewRelic.Region = region

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\nConfiguration saved successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  • View configuration: ops-cli newrelic config show")
	fmt.Println("  • Test connection: ops-cli newrelic config test")
	fmt.Println("  • Query logs: ops-cli newrelic logs search")

	return nil
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		RunE:  runConfigShow,
	}
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.NewRelic == nil || cfg.NewRelic.APIKey == "" {
		fmt.Println("New Relic not configured. Run 'ops-cli newrelic config setup'")
		return nil
	}

	fmt.Println()
	fmt.Println("Current New Relic Configuration:")
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("Account ID:    %s\n", cfg.NewRelic.AccountID)

	apiKeyDisplay := "Not set"
	if cfg.NewRelic.APIKey != "" {
		keyLen := len(cfg.NewRelic.APIKey)
		if keyLen > 4 {
			apiKeyDisplay = fmt.Sprintf("%s%s", strings.Repeat("*", 20), cfg.NewRelic.APIKey[keyLen-4:])
		} else {
			apiKeyDisplay = strings.Repeat("*", 20)
		}
	}
	fmt.Printf("API Key:       %s\n", apiKeyDisplay)
	fmt.Printf("Default Query: %s\n", cfg.NewRelic.DefaultQuery)
	fmt.Printf("Log Level:    %s\n", cfg.NewRelic.LogLevel)
	if cfg.NewRelic.Region != "" {
		fmt.Printf("Region:       %s\n", cfg.NewRelic.Region)
	}

	// Check environment variables
	fmt.Println("\nEnvironment variables (take precedence):")
	if os.Getenv("NEW_RELIC_API_KEY") != "" {
		fmt.Println("  NEW_RELIC_API_KEY: ••••••••")
	}
	if os.Getenv("NEW_RELIC_ACCOUNT_ID") != "" {
		fmt.Printf("  NEW_RELIC_ACCOUNT_ID: %s\n", os.Getenv("NEW_RELIC_ACCOUNT_ID"))
	}
	if os.Getenv("NEW_RELIC_REGION") != "" {
		fmt.Printf("  NEW_RELIC_REGION: %s\n", os.Getenv("NEW_RELIC_REGION"))
	}

	return nil
}

func newConfigTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to New Relic API",
		RunE:  runConfigTest,
	}
}

func runConfigTest(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing New Relic connection...\n")

	client, err := getNewRelicClient()
	if err != nil {
		return err
	}

	stopSpinner := ui.StartSpinner("Testing connection...")
	ok, err := client.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Test with a simple query
	testQueryOptions := newrelic.LogsQueryOptions{
		Query: "FROM Log SELECT count(*) SINCE 1 hour ago",
		Limit: 1,
	}

	stopSpinner = ui.StartSpinner("Testing query...")
	result, err := client.QueryLogs(testQueryOptions)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("query test failed: %w", err)
	}

	fmt.Println("Connection successful!")
	if len(result.Results) > 0 {
		fmt.Printf("Test query returned: %d result(s)\n", len(result.Results))
	} else {
		fmt.Println("Connection established but no data returned")
	}

	return nil
}
