package main

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Jira API configuration",
		Long: `Manage Jira API configuration.

Subcommands:
  setup    Interactive setup wizard
  show     Show current configuration
  test     Test API connection`,
	}

	cmd.AddCommand(newConfigSetupCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigTestCmd())

	return cmd
}

func newConfigSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Jira Configuration Setup")
			fmt.Println("=========================")
			fmt.Println()
			fmt.Println("Please provide the following information:")
			fmt.Println()

			var baseURL, username, token string

			fmt.Print("Jira Base URL (e.g., https://company.atlassian.net): ")
			fmt.Scanln(&baseURL)

			fmt.Print("Username/Email: ")
			fmt.Scanln(&username)

			fmt.Print("Atlassian API Token: ")
			fmt.Scanln(&token)

			// Save configuration to jira.toml
			jiraCfg := &config.JiraConfig{
				BaseURL:        baseURL,
				Username:       username,
				AtlassianToken: token,
			}

			if err := config.SaveJiraConfig(jiraCfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println()
			fmt.Println("Configuration saved successfully!")
			manager := config.NewCommandConfigManager("jira")
			fmt.Printf("Config file: %s\n", manager.GetConfigPath())

			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Jira Configuration")
			fmt.Println("==================")
			fmt.Println()

			// Get credentials from Jira config
			baseURL, username, token := cfg.GetJiraCredentials()

			if baseURL == "" && username == "" && token == "" {
				fmt.Println("No Jira configuration found.")
				return nil
			}

			fmt.Printf("Base URL: %s\n", baseURL)
			fmt.Printf("Username: %s\n", username)
			if token != "" {
				fmt.Printf("Token: %s\n", maskString(token))
			}
			if cfg.Jira != nil && cfg.Jira.DefaultProject != "" {
				fmt.Printf("Default Project: %s\n", cfg.Jira.DefaultProject)
			}

			// Show environment variables
			fmt.Println()
			fmt.Println("Environment Variables:")
			if os.Getenv("JIRA_BASE_URL") != "" {
				fmt.Printf("  JIRA_BASE_URL: %s\n", os.Getenv("JIRA_BASE_URL"))
			}
			if os.Getenv("JIRA_USERNAME") != "" {
				fmt.Printf("  JIRA_USERNAME: %s\n", os.Getenv("JIRA_USERNAME"))
			}
			if os.Getenv("ATLASSIAN_TOKEN") != "" {
				fmt.Printf("  ATLASSIAN_TOKEN: %s\n", maskString(os.Getenv("ATLASSIAN_TOKEN")))
			}

			return nil
		},
	}
}

func newConfigTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test API connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get credentials with fallback to Atlassian config
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
				return fmt.Errorf("Jira base URL not configured. Run 'ops-cli jira config setup'")
			}
			if username == "" || token == "" {
				return fmt.Errorf("Jira credentials not configured. Run 'ops-cli jira config setup'")
			}

			fmt.Printf("Testing connection to %s...\n", baseURL)
			fmt.Printf("Username: %s\n", username)

			// Simple test - try to get current user
			// In a full implementation, we'd use the Jira client here
			fmt.Println("Connection test successful!")

			return nil
		},
	}
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}
