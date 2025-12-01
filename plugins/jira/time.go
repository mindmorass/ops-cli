package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	jiraapi "github.com/andygrunwald/go-jira"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "time",
		Short: "Log time worked on assigned issues or generate time reports",
		Long: `Log time worked on assigned issues or generate time reports.

Examples:
  ops-cli jira time
  ops-cli jira time 10
  ops-cli jira time --daily-total
  ops-cli jira time report --week`,
		RunE: runTime,
	}

	cmd.Flags().Bool("no-active-sprint", false, "Exclude active sprint issues")
	cmd.Flags().Bool("daily-total", false, "Show daily total")
	cmd.Flags().Bool("day-total", false, "Show day total")
	cmd.Flags().String("date", "", "Date for daily total")
	cmd.Flags().Bool("distribute", false, "Distribute time across issues")
	cmd.Flags().String("distribute-time", "8h", "Time to distribute (default: 8h)")
	cmd.Flags().Bool("report", false, "Generate time report")
	cmd.Flags().Bool("week", false, "Generate weekly report")
	cmd.Flags().Bool("month", false, "Generate monthly report")
	cmd.Flags().String("start", "", "Start date for report (YYYY-MM-DD)")
	cmd.Flags().String("end", "", "End date for report (YYYY-MM-DD)")

	return cmd
}

func runTime(cmd *cobra.Command, args []string) error {
	jiraClient, err := getJiraClient()
	if err != nil {
		return err
	}

	// Get the underlying go-jira client
	client := jiraClient.GetClient()

	dailyTotal, _ := cmd.Flags().GetBool("daily-total")
	dayTotal, _ := cmd.Flags().GetBool("day-total")
	distribute, _ := cmd.Flags().GetBool("distribute")
	report, _ := cmd.Flags().GetBool("report")

	// Check subcommand
	subcommand := ""
	if len(args) > 0 {
		subcommand = args[0]
	}

	if dailyTotal || dayTotal {
		return runDailyTotal(cmd, client, args)
	}

	if distribute {
		return runDistribute(cmd, client, args)
	}

	if subcommand == "report" || report {
		return runReport(cmd, client, args)
	}

	// Default: interactive time logging
	return runInteractiveTimeLogging(cmd, client, args)
}

func runInteractiveTimeLogging(cmd *cobra.Command, client *jiraapi.Client, args []string) error {
	maxResults := 20
	if len(args) > 0 {
		if parsed, err := strconv.Atoi(args[0]); err == nil {
			maxResults = parsed
		}
	}

	noActiveSprint, _ := cmd.Flags().GetBool("no-active-sprint")

	// Build JQL
	jql := `assignee = currentUser() AND statusCategory = "In Progress"`
	if !noActiveSprint {
		jql += ` AND sprint in openSprints()`
	}
	jql += " ORDER BY updated DESC"

	// Search for issues
	stopSpinner := ui.StartSpinner("Finding assigned issues...")
	issues, _, err := client.Issue.Search(jql, &jiraapi.SearchOptions{
		MaxResults: maxResults,
		Fields:     []string{"summary", "status", "assignee", "updated", "priority", "sprint", "issuetype", "project"},
	})
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to search issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No assigned issues found!")
		return nil
	}

	// Sort issues by key
	sortedIssues := sortIssues(issues)

	fmt.Printf("\nFound %d issue(s):\n\n", len(sortedIssues))
	for i, issue := range sortedIssues {
		status := "Unknown"
		if issue.Fields.Status != nil {
			status = issue.Fields.Status.Name
		}
		issueType := "Unknown"
		if issue.Fields.Type.Name != "" {
			issueType = issue.Fields.Type.Name
		}
		fmt.Printf("%d. %s - %s\n", i+1, issue.Key, truncateText(issue.Fields.Summary, 60))
		fmt.Printf("   Type: %s | Status: %s\n\n", issueType, status)
	}

	// Interactive selection using survey
	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message:  "Select issues to log time on:",
		Options:  buildIssueOptions(sortedIssues),
		PageSize: 10,
	}

	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No issues selected. Exiting.")
		return nil
	}

	fmt.Printf("\nLogging time for %d selected issue(s)...\n", len(selectedIndices))

	// Log time for each selected issue
	for _, idx := range selectedIndices {
		if idx < 0 || idx >= len(sortedIssues) {
			continue
		}
		issue := sortedIssues[idx]

		fmt.Printf("\n--- %s - %s ---\n", issue.Key, issue.Fields.Summary)

		// Prompt for time spent
		timeSpent := ""
		timePrompt := &survey.Input{
			Message: "Enter time worked (Jira format):",
			Help:    "Examples: 2h, 1d, 30m, 1h 30m, 2d 4h (w=weeks, d=days, h=hours, m=minutes)",
		}
		if err := survey.AskOne(timePrompt, &timeSpent, survey.WithValidator(survey.Required)); err != nil {
			fmt.Printf("Skipped time entry for %s\n", issue.Key)
			continue
		}

		// Prompt for work description (optional)
		comment := ""
		commentPrompt := &survey.Input{
			Message: "Enter work description (optional):",
		}
		survey.AskOne(commentPrompt, &comment)

		// Prompt for work date (optional)
		workDate := ""
		datePrompt := &survey.Input{
			Message: "Enter work date (optional, default: today):",
			Help:    "Format: YYYY-MM-DD or leave empty for today",
		}
		survey.AskOne(datePrompt, &workDate)

		// Parse work date
		var started time.Time
		if workDate != "" {
			parsed, err := time.Parse("2006-01-02", workDate)
			if err != nil {
				fmt.Printf("Invalid date format, using today: %v\n", err)
				started = time.Now()
			} else {
				started = parsed
			}
		} else {
			started = time.Now()
		}

		// Create worklog
		startedTime := jiraapi.Time(started)
		worklog := &jiraapi.WorklogRecord{
			TimeSpent: timeSpent,
			Started:   &startedTime,
		}
		if comment != "" {
			worklog.Comment = comment
		}

		// Add worklog
		stopSpinner := ui.StartSpinner(fmt.Sprintf("Logging time for %s...", issue.Key))
		_, _, err := client.Issue.AddWorklogRecord(issue.Key, worklog)
		stopSpinner()
		if err != nil {
			fmt.Printf("Failed to log time for %s: %v\n", issue.Key, err)
			continue
		}

		fmt.Printf("Logged %s for %s\n", timeSpent, issue.Key)
	}

	fmt.Println("\nTime logging completed!")
	return nil
}

func runDistribute(cmd *cobra.Command, client *jiraapi.Client, args []string) error {
	distributeTime, _ := cmd.Flags().GetString("distribute-time")
	noActiveSprint, _ := cmd.Flags().GetBool("no-active-sprint")

	// Build JQL
	jql := `assignee = currentUser() AND statusCategory = "In Progress"`
	if !noActiveSprint {
		jql += ` AND sprint in openSprints()`
	}
	jql += " ORDER BY updated DESC"

	stopSpinner := ui.StartSpinner("Finding assigned issues...")
	issues, _, err := client.Issue.Search(jql, &jiraapi.SearchOptions{
		MaxResults: 50,
		Fields:     []string{"summary", "status", "assignee", "updated", "priority", "sprint", "issuetype", "project"},
	})
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to search issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No assigned issues found!")
		return nil
	}

	sortedIssues := sortIssues(issues)

	fmt.Printf("\nFound %d issue(s):\n\n", len(sortedIssues))
	for i, issue := range sortedIssues {
		status := "Unknown"
		if issue.Fields.Status != nil {
			status = issue.Fields.Status.Name
		}
		fmt.Printf("%d. %s - %s\n", i+1, issue.Key, truncateText(issue.Fields.Summary, 60))
		fmt.Printf("   Status: %s\n\n", status)
	}

	// Select issues
	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message:  "Select issues to distribute time across:",
		Options:  buildIssueOptions(sortedIssues),
		PageSize: 10,
	}

	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No issues selected.")
		return nil
	}

	// Calculate time per issue
	// For now, just divide evenly (could be enhanced)
	timePerIssue := distributeTime // Simplified - in production, parse and divide

	fmt.Printf("\nDistributing %s across %d issues...\n", distributeTime, len(selectedIndices))

	for _, idx := range selectedIndices {
		if idx < 0 || idx >= len(sortedIssues) {
			continue
		}
		issue := sortedIssues[idx]

		nowTime := jiraapi.Time(time.Now())
		worklog := &jiraapi.WorklogRecord{
			TimeSpent: timePerIssue,
			Started:   &nowTime,
		}

		stopSpinner := ui.StartSpinner(fmt.Sprintf("Logging time for %s...", issue.Key))
		_, _, err := client.Issue.AddWorklogRecord(issue.Key, worklog)
		stopSpinner()
		if err != nil {
			fmt.Printf("Failed to log time for %s: %v\n", issue.Key, err)
			continue
		}

		fmt.Printf("Logged %s for %s\n", timePerIssue, issue.Key)
	}

	fmt.Println("\nTime distribution completed!")
	return nil
}

func runDailyTotal(cmd *cobra.Command, client *jiraapi.Client, args []string) error {
	// TODO: Implement daily total
	fmt.Println("Daily total feature not yet implemented")
	return nil
}

func runReport(cmd *cobra.Command, client *jiraapi.Client, args []string) error {
	// TODO: Implement time report
	fmt.Println("Time report feature not yet implemented")
	return nil
}

func buildIssueOptions(issues []jiraapi.Issue) []string {
	options := make([]string, len(issues))
	for i, issue := range issues {
		options[i] = fmt.Sprintf("%s - %s", issue.Key, truncateText(issue.Fields.Summary, 60))
	}
	return options
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

func sortIssues(issues []jiraapi.Issue) []jiraapi.Issue {
	// Simple sort by key (extract number and sort)
	// For now, just return as-is
	return issues
}
