package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ops-cli/internal/api/newrelic"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Stream and search New Relic logs",
		Long: `Stream and search New Relic logs using NRQL queries.

Subcommands:
  stream    Stream logs in real-time
  search    Search logs by text
  stats     Show log statistics`,
	}

	cmd.AddCommand(newLogsStreamCmd())
	cmd.AddCommand(newLogsSearchCmd())
	cmd.AddCommand(newLogsStatsCmd())

	return cmd
}

func newLogsStreamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream logs in real-time",
		Long: `Stream New Relic logs in real-time using polling.

Examples:
  ops-cli newrelic logs stream
  ops-cli newrelic logs stream --query "FROM Log SELECT * WHERE level = 'ERROR'"
  ops-cli newrelic logs stream --since 1h --follow`,
		RunE: runLogsStream,
	}

	cmd.Flags().String("query", "", "NRQL query (default: FROM Log SELECT *)")
	cmd.Flags().String("since", "1h", "Time range (e.g., 1h, 30m, 2d)")
	cmd.Flags().Int("limit", 100, "Maximum number of logs per poll")
	cmd.Flags().Bool("follow", false, "Continue streaming (default: false)")
	cmd.Flags().String("format", "json", "Output format: json, table, raw")

	return cmd
}

func runLogsStream(cmd *cobra.Command, args []string) error {
	client, err := getNewRelicClient()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	query, _ := cmd.Flags().GetString("query")
	if query == "" {
		if cfg.NewRelic != nil && cfg.NewRelic.DefaultQuery != "" {
			query = cfg.NewRelic.DefaultQuery
		} else {
			query = "FROM Log SELECT *"
		}
	}

	since, _ := cmd.Flags().GetString("since")
	limit, _ := cmd.Flags().GetInt("limit")
	follow, _ := cmd.Flags().GetBool("follow")
	format, _ := cmd.Flags().GetString("format")

	fmt.Printf("Streaming New Relic logs...\n")
	if query != "FROM Log SELECT *" {
		fmt.Printf("Query: %s\n", query)
	}

	// For now, just do a single query (streaming can be added later)
	options := newrelic.LogsQueryOptions{
		Query: query,
		Since: since,
		Limit: limit,
	}

	stopSpinner := ui.StartSpinner("Fetching logs...")
	result, err := client.QueryLogs(options)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to query logs: %w", err)
	}

	fmt.Printf("Found %d log entries\n\n", len(result.Results))

	// Display logs based on format
	for i, log := range result.Results {
		if i >= 20 && !follow {
			fmt.Printf("\n... and %d more logs (use --follow to see all)\n", len(result.Results)-20)
			break
		}

		switch format {
		case "json":
			jsonData, _ := json.Marshal(log)
			fmt.Println(string(jsonData))
		case "table":
			timestamp := time.Unix(log.Timestamp/1000, 0).Format(time.RFC3339)
			level := log.Level
			if level == "" {
				level = "INFO"
			}
			fmt.Printf("[%s] %s: %s\n", timestamp, level, log.Message)
		case "raw":
			fmt.Println(log.Message)
		default:
			jsonData, _ := json.Marshal(log)
			fmt.Println(string(jsonData))
		}
	}

	return nil
}

func newLogsSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search logs by text",
		Long: `Search New Relic logs by text content.

Examples:
  ops-cli newrelic logs search --text "error"
  ops-cli newrelic logs search --text "timeout" --level ERROR
  ops-cli newrelic logs search --text "login" --since 24h`,
		RunE: runLogsSearch,
	}

	cmd.Flags().String("text", "", "Text to search for in log messages (required)")
	cmd.Flags().String("level", "", "Filter by log level (e.g., ERROR, WARN, INFO)")
	cmd.Flags().String("since", "24h", "Time range (e.g., 1h, 30m, 2d)")
	cmd.Flags().Int("limit", 1000, "Maximum number of results")

	return cmd
}

func runLogsSearch(cmd *cobra.Command, args []string) error {
	client, err := getNewRelicClient()
	if err != nil {
		return err
	}

	text, _ := cmd.Flags().GetString("text")
	if text == "" {
		return fmt.Errorf("--text is required")
	}

	level, _ := cmd.Flags().GetString("level")
	since, _ := cmd.Flags().GetString("since")
	limit, _ := cmd.Flags().GetInt("limit")

	// Build query
	query := "FROM Log SELECT *"
	if text != "" {
		query += fmt.Sprintf(" WHERE message LIKE '%%%s%%'", text)
	}
	if level != "" && level != "all" {
		if text != "" {
			query += fmt.Sprintf(" AND level = '%s'", level)
		} else {
			query += fmt.Sprintf(" WHERE level = '%s'", level)
		}
	}

	options := newrelic.LogsQueryOptions{
		Query: query,
		Since: since,
		Limit: limit,
	}

	stopSpinner := ui.StartSpinner("Searching logs...")
	result, err := client.QueryLogs(options)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to search logs: %w", err)
	}

	fmt.Printf("Found %d matching logs\n\n", len(result.Results))

	// Display first 20 logs
	for i, log := range result.Results {
		if i >= 20 {
			fmt.Printf("\n... and %d more logs\n", len(result.Results)-20)
			break
		}
		timestamp := time.Unix(log.Timestamp/1000, 0).Format(time.RFC3339)
		level := log.Level
		if level == "" {
			level = "INFO"
		}
		fmt.Printf("[%s] %s: %s\n", timestamp, level, log.Message)
	}

	return nil
}

func newLogsStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show log statistics",
		Long: `Show log statistics grouped by level or other fields.

Examples:
  ops-cli newrelic logs stats
  ops-cli newrelic logs stats --since 24h
  ops-cli newrelic logs stats --group-by service`,
		RunE: runLogsStats,
	}

	cmd.Flags().String("since", "24h", "Time range (e.g., 1h, 30m, 2d)")
	cmd.Flags().String("group-by", "level", "Field to group by (level, service, etc.)")

	return cmd
}

func runLogsStats(cmd *cobra.Command, args []string) error {
	client, err := getNewRelicClient()
	if err != nil {
		return err
	}

	since, _ := cmd.Flags().GetString("since")
	groupBy, _ := cmd.Flags().GetString("group-by")

	query := fmt.Sprintf("SELECT count(*) FROM Log FACET %s", groupBy)

	options := newrelic.LogsQueryOptions{
		Query: query,
		Since: since,
	}

	stopSpinner := ui.StartSpinner("Fetching statistics...")
	result, err := client.QueryLogs(options)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("\nLog Statistics (grouped by %s):\n", groupBy)
	fmt.Println("────────────────────────────────────────")

	if len(result.Results) == 0 {
		fmt.Println("No data available")
		return nil
	}

	// Display stats
	for _, log := range result.Results {
		// Extract count and facet value from attributes
		count := 0
		facetValue := "unknown"

		for k, v := range log.Attributes {
			if k == "count" || k == "value" {
				if num, ok := v.(float64); ok {
					count = int(num)
				}
			} else if k == groupBy {
				if str, ok := v.(string); ok {
					facetValue = str
				}
			}
		}

		fmt.Printf("%-20s: %d\n", facetValue, count)
	}

	return nil
}
