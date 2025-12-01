package main

import (
	"encoding/json"
	"fmt"
	"os"

	jiraapi "github.com/andygrunwald/go-jira"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <issue-key>",
		Short: "Get a specific Jira issue",
		Args:  cobra.ExactArgs(1),
		Long: `Get detailed information about a specific Jira issue.

Examples:
  ops-cli jira get ABC-123
  ops-cli jira get ABC-123 --format json --color`,
		RunE: runGet,
	}

	cmd.Flags().String("format", "table", "Output format: table, json")
	cmd.Flags().Bool("color", false, "Enable colored output")
	cmd.Flags().Bool("unformatted", false, "Show all fields as raw JSON")
	cmd.Flags().Bool("raw", false, "Show all fields as raw JSON (alias for --unformatted)")
	cmd.Flags().String("output", "", "Write output to file")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	client, err := getJiraClient()
	if err != nil {
		return err
	}

	// Use spinner while fetching
	stopSpinner := ui.StartSpinner("Fetching issue...")
	jiraClient := client.GetClient()
	issue, _, err := jiraClient.Issue.Get(issueKey, nil)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Format output
	format, _ := cmd.Flags().GetString("format")
	unformatted, _ := cmd.Flags().GetBool("unformatted")
	raw, _ := cmd.Flags().GetBool("raw")
	outputFile, _ := cmd.Flags().GetString("output")

	var output string
	if unformatted || raw || format == "json" {
		output = formatIssueJSON(issue)
	} else {
		output = formatIssueTable(issue)
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

func formatIssueTable(issue *jiraapi.Issue) string {
	var result string
	result += fmt.Sprintf("Issue: %s\n", issue.Key)
	result += fmt.Sprintf("Summary: %s\n", issue.Fields.Summary)
	if issue.Fields.Status != nil {
		result += fmt.Sprintf("Status: %s\n", issue.Fields.Status.Name)
	}
	if issue.Fields.Assignee != nil {
		result += fmt.Sprintf("Assignee: %s\n", issue.Fields.Assignee.DisplayName)
	}
	if issue.Fields.Priority != nil {
		result += fmt.Sprintf("Priority: %s\n", issue.Fields.Priority.Name)
	}
	if issue.Fields.Description != "" {
		result += fmt.Sprintf("Description: %s\n", issue.Fields.Description)
	}
	return result
}

func formatIssueJSON(issue *jiraapi.Issue) string {
	type IssueOutput struct {
		Key         string `json:"key"`
		Summary     string `json:"summary"`
		Status      string `json:"status,omitempty"`
		Assignee    string `json:"assignee,omitempty"`
		Priority    string `json:"priority,omitempty"`
		Description string `json:"description,omitempty"`
		Created     string `json:"created,omitempty"`
		Updated     string `json:"updated,omitempty"`
	}

	output := IssueOutput{
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
	}

	if issue.Fields.Status != nil {
		output.Status = issue.Fields.Status.Name
	}
	if issue.Fields.Assignee != nil {
		output.Assignee = issue.Fields.Assignee.DisplayName
	}
	if issue.Fields.Priority != nil {
		output.Priority = issue.Fields.Priority.Name
	}
	// Convert jira.Time to string using fmt.Sprintf
	output.Created = fmt.Sprintf("%v", issue.Fields.Created)
	output.Updated = fmt.Sprintf("%v", issue.Fields.Updated)

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(data)
}
