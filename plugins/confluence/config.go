package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ops-cli/internal/api/confluence"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Confluence API configuration",
		Long: `Manage Confluence API configuration.

Subcommands:
  setup    Set up Confluence API credentials interactively
  show     Display current configuration
  test     Test connection to Confluence API`,
	}

	cmd.AddCommand(newConfigSetupCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigTestCmd())

	return cmd
}

func newConfigSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up Confluence API credentials interactively",
		RunE:  runConfigSetup,
	}
}

func runConfigSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Setting up Confluence API configuration...\n")

	// Prompt for base URL
	baseURL := ""
	baseURLPrompt := &survey.Input{
		Message: "Enter your Confluence base URL:",
		Help:    "e.g., https://company.atlassian.net",
	}
	if err := survey.AskOne(baseURLPrompt, &baseURL, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("base URL is required")
	}

	// Validate URL
	if _, err := url.Parse(baseURL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Remove trailing slash
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	// Prompt for username
	username := ""
	usernamePrompt := &survey.Input{
		Message: "Enter your Confluence username/email:",
	}
	if err := survey.AskOne(usernamePrompt, &username, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("username is required")
	}

	// Prompt for token
	token := ""
	tokenPrompt := &survey.Password{
		Message: "Enter your Atlassian token:",
	}
	if err := survey.AskOne(tokenPrompt, &token, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("token is required")
	}

	// Test the configuration
	fmt.Println("\nTesting configuration...")
	client, err := confluence.NewClient(baseURL, username, token)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	stopSpinner := ui.StartSpinner("Testing connection...")
	ok, err := client.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Get current user
	stopSpinner = ui.StartSpinner("Fetching user info...")
	user, err := client.GetCurrentUser()
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	fmt.Printf("Successfully connected as: %s (%s)\n", user.DisplayName, user.Email)

	// Save configuration to confluence.toml
	confluenceCfg := &config.ConfluenceConfig{
		BaseURL:        baseURL,
		Username:       username,
		AtlassianToken: token,
	}

	if err := config.SaveConfluenceConfig(confluenceCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\nConfiguration saved successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  • View configuration: ops-cli confluence config show")
	fmt.Println("  • Test connection: ops-cli confluence config test")
	fmt.Println("  • Get a page: ops-cli confluence get <page-id>")

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

	// Get credentials from Confluence config
	baseURL, username, token := cfg.GetConfluenceCredentials()

	if baseURL == "" && username == "" && token == "" {
		fmt.Println("Confluence not configured. Run 'ops-cli confluence config setup'")
		return nil
	}

	fmt.Println("Current Confluence Configuration:\n")

	fmt.Printf("Base URL:      %s\n", baseURL)
	fmt.Printf("Username:      %s\n", username)

	tokenDisplay := "Not set"
	if token != "" {
		tokenDisplay = "••••••••"
	}
	fmt.Printf("Token:         %s\n", tokenDisplay)

	// Check environment variables
	fmt.Println("\nEnvironment variables (take precedence):")
	if os.Getenv("CONFLUENCE_BASE_URL") != "" {
		fmt.Printf("  CONFLUENCE_BASE_URL: %s\n", os.Getenv("CONFLUENCE_BASE_URL"))
	}
	if os.Getenv("CONFLUENCE_USERNAME") != "" {
		fmt.Printf("  CONFLUENCE_USERNAME: %s\n", os.Getenv("CONFLUENCE_USERNAME"))
	}
	if os.Getenv("ATLASSIAN_TOKEN") != "" {
		fmt.Println("  ATLASSIAN_TOKEN: ••••••••")
	}

	return nil
}

func newConfigTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to Confluence API",
		RunE:  runConfigTest,
	}
}

func runConfigTest(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing Confluence connection...\n")

	client, err := getConfluenceClient()
	if err != nil {
		return err
	}

	stopSpinner := ui.StartSpinner("Testing connection...")
	ok, err := client.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Get user info
	stopSpinner = ui.StartSpinner("Fetching user info...")
	user, err := client.GetCurrentUser()
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	fmt.Println("Connection successful!")
	fmt.Printf("User: %s (%s)\n", user.DisplayName, user.Email)
	if user.AccountID != "" {
		fmt.Printf("Account ID: %s\n", user.AccountID)
	}

	// Test basic operations
	fmt.Println("\nTesting basic operations...")
	stopSpinner = ui.StartSpinner("Fetching spaces...")
	spaces, err := client.GetSpaces(nil, 3)
	stopSpinner()
	if err != nil {
		fmt.Printf("Warning: Could not fetch spaces: %v\n", err)
	} else {
		fmt.Printf("Can access %d spaces\n", spaces.Size)
		if len(spaces.Results) > 0 {
			fmt.Println("Sample spaces:")
			for _, space := range spaces.Results {
				fmt.Printf("  • %s: %s\n", space.Key, space.Name)
			}
		}
	}

	fmt.Println("\nConfluence API is ready to use!")
	return nil
}
