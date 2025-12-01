package atlassian

import (
	"fmt"
	"net/url"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ops-cli/internal/api/confluence"
	"github.com/ops-cli/internal/api/jira"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Atlassian API configuration",
		Long: `Manage shared Atlassian API configuration.

Subcommands:
  setup    Set up shared Atlassian API credentials interactively
  show     Display current configuration
  test     Test connection to Atlassian services`,
	}

	cmd.AddCommand(newConfigSetupCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigTestCmd())

	return cmd
}

func newConfigSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up shared Atlassian API credentials interactively",
		RunE:  runConfigSetup,
	}
}

func runConfigSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Setting up shared Atlassian API configuration...")
	fmt.Println("This configuration will be used by both Jira and Confluence.\n")

	// Prompt for base URL
	baseURL := ""
	baseURLPrompt := &survey.Input{
		Message: "Enter your Atlassian base URL:",
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
	if baseURL != "" && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	// Prompt for username
	username := ""
	usernamePrompt := &survey.Input{
		Message: "Enter your Atlassian username/email:",
	}
	if err := survey.AskOne(usernamePrompt, &username, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("username is required")
	}

	// Prompt for token
	token := ""
	tokenPrompt := &survey.Password{
		Message: "Enter your Atlassian API token:",
	}
	if err := survey.AskOne(tokenPrompt, &token, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("token is required")
	}

	// Test the configuration with both services
	fmt.Println("\nTesting configuration...")

	// Test Jira connection
	fmt.Println("Testing Jira connection...")
	jiraClient, err := jira.NewClient(baseURL, username, token)
	if err != nil {
		return fmt.Errorf("failed to create Jira client: %w", err)
	}

	stopSpinner := ui.StartSpinner("Testing Jira connection...")
	_, err = jiraClient.GetCurrentUser()
	stopSpinner()
	if err != nil {
		fmt.Printf("Warning: Jira connection test failed: %v\n", err)
		fmt.Println("You can still save the configuration and test later.")
	} else {
		fmt.Println("✓ Jira connection successful")
	}

	// Test Confluence connection
	fmt.Println("Testing Confluence connection...")
	confluenceClient, err := confluence.NewClient(baseURL, username, token)
	if err != nil {
		return fmt.Errorf("failed to create Confluence client: %w", err)
	}

	stopSpinner = ui.StartSpinner("Testing Confluence connection...")
	ok, err := confluenceClient.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		fmt.Printf("Warning: Confluence connection test failed: %v\n", err)
		fmt.Println("You can still save the configuration and test later.")
	} else {
		// Get user info
		user, err := confluenceClient.GetCurrentUser()
		if err == nil {
			fmt.Printf("✓ Confluence connection successful (User: %s)\n", user.DisplayName)
		} else {
			fmt.Println("✓ Confluence connection successful")
		}
	}

	// Save configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Atlassian == nil {
		cfg.Atlassian = &config.AtlassianConfig{}
	}

	cfg.Atlassian.BaseURL = baseURL
	cfg.Atlassian.Username = username
	cfg.Atlassian.AtlassianToken = token

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\n✓ Shared Atlassian configuration saved successfully!")
	fmt.Println("\nThis configuration will be used by:")
	fmt.Println("  • Jira (ops-cli jira)")
	fmt.Println("  • Confluence (ops-cli confluence)")
	fmt.Println("\nNext steps:")
	fmt.Println("  • View configuration: ops-cli atlassian config show")
	fmt.Println("  • Test connections: ops-cli atlassian config test")
	fmt.Println("  • Override for Jira: ops-cli jira config setup")
	fmt.Println("  • Override for Confluence: ops-cli confluence config setup")

	return nil
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current shared Atlassian configuration",
		RunE:  runConfigShow,
	}
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Shared Atlassian Configuration:\n")

	if cfg.Atlassian == nil || (cfg.Atlassian.BaseURL == "" && cfg.Atlassian.Username == "" && cfg.Atlassian.AtlassianToken == "") {
		fmt.Println("No shared Atlassian configuration found.")
		fmt.Println("\nYou can set it up with: ops-cli atlassian config setup")
		fmt.Println("\nOr configure individually:")
		fmt.Println("  • Jira: ops-cli jira config setup")
		fmt.Println("  • Confluence: ops-cli confluence config setup")
		return nil
	}

	fmt.Printf("Base URL:      %s\n", cfg.Atlassian.BaseURL)
	fmt.Printf("Username:      %s\n", cfg.Atlassian.Username)

	tokenDisplay := "Not set"
	if cfg.Atlassian.AtlassianToken != "" {
		tokenDisplay = "••••••••"
	}
	fmt.Printf("Token:         %s\n", tokenDisplay)

	// Show which services are using this config
	fmt.Println("\nUsed by:")
	jiraBaseURL, jiraUsername, jiraToken := cfg.GetJiraCredentials()
	confluenceBaseURL, confluenceUsername, confluenceToken := cfg.GetConfluenceCredentials()

	if jiraBaseURL == cfg.Atlassian.BaseURL && jiraUsername == cfg.Atlassian.Username && jiraToken == cfg.Atlassian.AtlassianToken {
		fmt.Println("  ✓ Jira (using shared config)")
	} else {
		fmt.Println("  • Jira (using individual config)")
	}

	if confluenceBaseURL == cfg.Atlassian.BaseURL && confluenceUsername == cfg.Atlassian.Username && confluenceToken == cfg.Atlassian.AtlassianToken {
		fmt.Println("  ✓ Confluence (using shared config)")
	} else {
		fmt.Println("  • Confluence (using individual config)")
	}

	// Check environment variables
	fmt.Println("\nEnvironment variables (take precedence):")
	if os.Getenv("JIRA_BASE_URL") != "" || os.Getenv("CONFLUENCE_BASE_URL") != "" {
		if os.Getenv("JIRA_BASE_URL") != "" {
			fmt.Printf("  JIRA_BASE_URL: %s\n", os.Getenv("JIRA_BASE_URL"))
		}
		if os.Getenv("CONFLUENCE_BASE_URL") != "" {
			fmt.Printf("  CONFLUENCE_BASE_URL: %s\n", os.Getenv("CONFLUENCE_BASE_URL"))
		}
	}
	if os.Getenv("JIRA_USERNAME") != "" || os.Getenv("CONFLUENCE_USERNAME") != "" {
		if os.Getenv("JIRA_USERNAME") != "" {
			fmt.Printf("  JIRA_USERNAME: %s\n", os.Getenv("JIRA_USERNAME"))
		}
		if os.Getenv("CONFLUENCE_USERNAME") != "" {
			fmt.Printf("  CONFLUENCE_USERNAME: %s\n", os.Getenv("CONFLUENCE_USERNAME"))
		}
	}
	if os.Getenv("ATLASSIAN_TOKEN") != "" {
		fmt.Println("  ATLASSIAN_TOKEN: ••••••••")
	}

	return nil
}

func newConfigTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to Atlassian services",
		RunE:  runConfigTest,
	}
}

func runConfigTest(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing Atlassian connections...\n")

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get credentials with fallback
	baseURL, username, token := cfg.GetAtlassianCredentials()

	if baseURL == "" {
		return fmt.Errorf("Atlassian base URL not configured. Run 'ops-cli atlassian config setup'")
	}
	if username == "" || token == "" {
		return fmt.Errorf("Atlassian credentials not configured. Run 'ops-cli atlassian config setup'")
	}

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

	// Test Jira
	fmt.Println("Testing Jira connection...")
	jiraClient, err := jira.NewClient(baseURL, username, token)
	if err != nil {
		return fmt.Errorf("failed to create Jira client: %w", err)
	}

	stopSpinner := ui.StartSpinner("Testing Jira connection...")
	user, err := jiraClient.GetCurrentUser()
	stopSpinner()
	if err != nil {
		fmt.Printf("✗ Jira connection failed: %v\n", err)
	} else {
		fmt.Printf("✓ Jira connection successful!\n")
		fmt.Printf("  User: %s (%s)\n", user.DisplayName, user.EmailAddress)
	}

	// Test Confluence
	fmt.Println("\nTesting Confluence connection...")
	confluenceClient, err := confluence.NewClient(baseURL, username, token)
	if err != nil {
		return fmt.Errorf("failed to create Confluence client: %w", err)
	}

	stopSpinner = ui.StartSpinner("Testing Confluence connection...")
	ok, err := confluenceClient.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		fmt.Printf("✗ Confluence connection failed: %v\n", err)
	} else {
		// Get user info
		stopSpinner = ui.StartSpinner("Fetching user info...")
		user, err := confluenceClient.GetCurrentUser()
		stopSpinner()
		if err != nil {
			fmt.Println("✓ Confluence connection successful!")
		} else {
			fmt.Printf("✓ Confluence connection successful!\n")
			fmt.Printf("  User: %s (%s)\n", user.DisplayName, user.Email)
		}
	}

	fmt.Println("\nAtlassian services are ready to use!")
	return nil
}

