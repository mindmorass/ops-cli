package main

import (
	"fmt"
	"os"

	jiraapi "github.com/andygrunwald/go-jira"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <summary>",
		Short: "Create a new Jira issue",
		Args:  cobra.MinimumNArgs(1),
		Long: `Create a new Jira issue.

Examples:
  ops-cli jira create "Fix login bug" --project ABC --type Bug
  ops-cli jira create "New feature" --project ABC --assignee me`,
		RunE: runCreate,
	}

	cmd.Flags().String("project", "", "Project key (required)")
	cmd.Flags().String("type", "Task", "Issue type (e.g., Bug, Task, Story)")
	cmd.Flags().String("assignee", "", "Assignee (use 'me' for current user)")
	cmd.Flags().String("priority", "", "Priority (e.g., Highest, High, Medium, Low, Lowest)")
	cmd.Flags().String("description", "", "Issue description")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	summary := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := getJiraClient()
	if err != nil {
		return err
	}

	// Get credentials with fallback to Atlassian config
	baseURL, username, _ := cfg.GetJiraCredentials()

	// Environment variables take precedence
	if envURL := os.Getenv("JIRA_BASE_URL"); envURL != "" {
		baseURL = envURL
	}
	if envUser := os.Getenv("JIRA_USERNAME"); envUser != "" {
		username = envUser
	}

	// Get flags
	project, _ := cmd.Flags().GetString("project")
	issueType, _ := cmd.Flags().GetString("type")
	assignee, _ := cmd.Flags().GetString("assignee")
	priority, _ := cmd.Flags().GetString("priority")
	description, _ := cmd.Flags().GetString("description")

	if project == "" {
		if cfg.Jira != nil && cfg.Jira.DefaultProject != "" {
			project = cfg.Jira.DefaultProject
		} else {
			return fmt.Errorf("project is required. Use --project or set default_project in config")
		}
	}

	// Create issue
	issue := &jiraapi.Issue{
		Fields: &jiraapi.IssueFields{
			Type: jiraapi.IssueType{
				Name: issueType,
			},
			Project: jiraapi.Project{
				Key: project,
			},
			Summary: summary,
		},
	}

	if description != "" {
		issue.Fields.Description = description
	}

	if assignee != "" {
		if assignee == "me" {
			issue.Fields.Assignee = &jiraapi.User{
				Name: username,
			}
		} else {
			issue.Fields.Assignee = &jiraapi.User{
				Name: assignee,
			}
		}
	}

	if priority != "" {
		issue.Fields.Priority = &jiraapi.Priority{
			Name: priority,
		}
	}

	// Use spinner while creating
	stopSpinner := ui.StartSpinner("Creating issue...")
	jiraClient := client.GetClient()
	created, _, err := jiraClient.Issue.Create(issue)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	fmt.Printf("Created issue: %s\n", created.Key)
	fmt.Printf("URL: %s/browse/%s\n", baseURL, created.Key)

	return nil
}
