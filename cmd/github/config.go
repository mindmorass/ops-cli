package github

import (
	"fmt"
	"os"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ops-cli/internal/api/github"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage GitHub configuration",
		Long: `Manage GitHub API configuration.

Subcommands:
  setup    Set up GitHub API credentials interactively
  show     Display current configuration
  test     Test connection to GitHub API`,
	}

	cmd.AddCommand(newConfigSetupCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigTestCmd())

	return cmd
}

func newConfigSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up GitHub API credentials interactively",
		RunE:  runConfigSetup,
	}
}

func runConfigSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Setting up GitHub API configuration...")
	fmt.Println()

	// Prompt for token
	token := ""
	tokenPrompt := &survey.Password{
		Message: "Enter your GitHub personal access token:",
	}
	if err := survey.AskOne(tokenPrompt, &token, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("token is required")
	}

	// Validate token format (basic check)
	tokenPattern := regexp.MustCompile(`^(ghp_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9_]{82})$`)
	hexPattern := regexp.MustCompile(`^[a-f0-9]{40}$`)

	if !tokenPattern.MatchString(token) && !hexPattern.MatchString(token) {
		fmt.Println("Warning: Token format appears invalid.")
		fmt.Println("Expected formats:")
		fmt.Println("  - Fine-grained tokens: github_pat_...")
		fmt.Println("  - Classic tokens: ghp_... or 40-character hex")

		confirm := false
		confirmPrompt := &survey.Confirm{
			Message: "Continue anyway?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil || !confirm {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}

	// Prompt for default user (optional)
	defaultUser := ""
	userPrompt := &survey.Input{
		Message: "Enter default GitHub username (optional):",
	}
	survey.AskOne(userPrompt, &defaultUser)

	// Prompt for default org (optional)
	defaultOrg := ""
	orgPrompt := &survey.Input{
		Message: "Enter default GitHub organization (optional):",
	}
	survey.AskOne(orgPrompt, &defaultOrg)

	// Test the configuration
	fmt.Println("\nTesting connection...")
	client := github.NewClient(token)

	stopSpinner := ui.StartSpinner("Testing connection...")
	ok, username, err := client.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		return fmt.Errorf("connection test failed: %w", err)
	}

	fmt.Printf("Authentication test passed! Authenticated as: %s\n", username)

	// Save configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.GitHub == nil {
		cfg.GitHub = &config.GitHubConfig{}
	}

	cfg.GitHub.Token = token
	if defaultUser != "" {
		cfg.GitHub.DefaultOwner = defaultUser
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\nConfiguration saved successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  • View configuration: ops-cli github config show")
	fmt.Println("  • Test connection: ops-cli github config test")
	fmt.Println("  • List repositories: ops-cli github repos")

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

	if cfg.GitHub == nil || cfg.GitHub.Token == "" {
		fmt.Println("GitHub not configured. Run 'ops-cli github config setup'")
		return nil
	}

	fmt.Println()
	fmt.Println("Current GitHub Configuration:")
	fmt.Println("────────────────────────────────────────")
	tokenDisplay := "Not set"
	if cfg.GitHub.Token != "" {
		tokenDisplay = "***configured***"
	}
	fmt.Printf("Token:            %s\n", tokenDisplay)
	fmt.Printf("Default Owner:   %s\n", cfg.GitHub.DefaultOwner)

	// Check environment variables
	fmt.Println("\nEnvironment variables (take precedence):")
	if os.Getenv("GITHUB_TOKEN") != "" {
		fmt.Println("  GITHUB_TOKEN: ••••••••")
	}

	// Test connection if token is available
	if cfg.GitHub.Token != "" || os.Getenv("GITHUB_TOKEN") != "" {
		fmt.Println("\nTesting connection...")
		client, err := getGitHubClient()
		if err == nil {
			stopSpinner := ui.StartSpinner("Testing...")
			ok, username, err := client.TestConnection()
			stopSpinner()
			if err == nil && ok {
				fmt.Printf("Connected as: %s\n", username)
			} else {
				fmt.Printf("Connection failed: %v\n", err)
			}
		}
	}

	return nil
}

func newConfigTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to GitHub API",
		RunE:  runConfigTest,
	}
}

func runConfigTest(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing GitHub API connection...")
	fmt.Println()

	client, err := getGitHubClient()
	if err != nil {
		return err
	}

	stopSpinner := ui.StartSpinner("Testing connection...")
	ok, username, err := client.TestConnection()
	stopSpinner()
	if err != nil || !ok {
		return fmt.Errorf("connection test failed: %w", err)
	}

	fmt.Println("GitHub API connection successful!")
	fmt.Printf("Authenticated as: %s\n", username)

	return nil
}
