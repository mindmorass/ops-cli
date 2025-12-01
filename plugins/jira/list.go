package main

import (
	"encoding/json"
	"fmt"
	"os"

	jiraapi "github.com/andygrunwald/go-jira"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Jira issues",
		Long: `List Jira issues based on various filters.

Examples:
  ops-cli jira list --project ABC
  ops-cli jira list --assignee me --status "In Progress"
  ops-cli jira list --format json --color`,
		RunE: runList,
	}

	cmd.Flags().String("project", "", "Filter by project key")
	cmd.Flags().String("assignee", "", "Filter by assignee (use 'me' for current user)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().Int("max-results", 50, "Maximum number of results")
	cmd.Flags().String("format", "table", "Output format: table, json")
	cmd.Flags().Bool("color", false, "Enable colored output")
	cmd.Flags().String("output", "", "Write output to file")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := getJiraClient()
	if err != nil {
		return err
	}

	// Build JQL query
	project, _ := cmd.Flags().GetString("project")
	assignee, _ := cmd.Flags().GetString("assignee")
	status, _ := cmd.Flags().GetString("status")
	maxResults, _ := cmd.Flags().GetInt("max-results")
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")

	var conditions []string

	if project != "" {
		conditions = append(conditions, fmt.Sprintf("project = %s", project))
	}

	if assignee != "" {
		if assignee == "me" || assignee == "currentUser()" {
			conditions = append(conditions, "assignee = currentUser()")
		} else {
			conditions = append(conditions, fmt.Sprintf(`assignee = "%s"`, assignee))
		}
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf(`status = "%s"`, status))
	}

	jql := ""
	if len(conditions) > 0 {
		jql = conditions[0]
		for i := 1; i < len(conditions); i++ {
			jql = fmt.Sprintf("%s AND %s", jql, conditions[i])
		}
		jql += " ORDER BY updated DESC"
	} else {
		jql = "assignee = currentUser() AND resolution = Unresolved ORDER BY updated DESC"
	}

	// Use spinner while searching
	stopSpinner := ui.StartSpinner("Searching issues...")
	issues, err := client.SearchIssues(jql, maxResults, []string{"summary", "status", "assignee", "priority", "created", "updated"})
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to search issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No issues found matching the criteria.")
		return nil
	}

	// Format output
	var output string
	if format == "json" {
		output = formatIssuesJSON(issues)
	} else {
		output = formatIssuesTable(issues)
	}

	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Output written to: %s\n", outputFile)
	} else {
		fmt.Println(output)
	}

	return nil
}

func formatIssuesTable(issues []jiraapi.Issue) string {
	var result string
	result += fmt.Sprintf("%-12s %-50s %-15s %-20s %-10s\n", "KEY", "SUMMARY", "STATUS", "ASSIGNEE", "PRIORITY")
	result += fmt.Sprintf("%-12s %-50s %-15s %-20s %-10s\n", "---", "-------", "------", "--------", "--------")

	for _, issue := range issues {
		assignee := "Unassigned"
		if issue.Fields.Assignee != nil {
			assignee = issue.Fields.Assignee.DisplayName
		}

		priority := "None"
		if issue.Fields.Priority != nil {
			priority = issue.Fields.Priority.Name
		}

		status := "Unknown"
		if issue.Fields.Status != nil {
			status = issue.Fields.Status.Name
		}

		summary := issue.Fields.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}

		result += fmt.Sprintf("%-12s %-50s %-15s %-20s %-10s\n",
			issue.Key,
			summary,
			status,
			assignee,
			priority,
		)
	}

	return result
}

func formatIssuesJSON(issues []jiraapi.Issue) string {
	type IssueOutput struct {
		Key      string `json:"key"`
		Summary  string `json:"summary"`
		Status   string `json:"status,omitempty"`
		Assignee string `json:"assignee,omitempty"`
		Priority string `json:"priority,omitempty"`
	}

	output := make([]IssueOutput, len(issues))
	for i, issue := range issues {
		output[i] = IssueOutput{
			Key:     issue.Key,
			Summary: issue.Fields.Summary,
		}
		if issue.Fields.Status != nil {
			output[i].Status = issue.Fields.Status.Name
		}
		if issue.Fields.Assignee != nil {
			output[i].Assignee = issue.Fields.Assignee.DisplayName
		}
		if issue.Fields.Priority != nil {
			output[i].Priority = issue.Fields.Priority.Name
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(data)
}
