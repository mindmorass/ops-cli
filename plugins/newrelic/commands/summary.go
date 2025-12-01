package main

import (
	"fmt"
	"strings"

	"github.com/ops-cli/internal/api/newrelic"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Display key metrics summary in table format",
		Long: `Display key metrics summary in table format.

Examples:
  ops-cli newrelic summary
  ops-cli newrelic summary --verbose`,
		RunE: runSummary,
	}

	cmd.Flags().Bool("verbose", false, "Show detailed output")

	return cmd
}

func runSummary(cmd *cobra.Command, args []string) error {
	client, err := getNewRelicClient()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	accountID := cfg.NewRelic.AccountID
	fmt.Printf("Fetching New Relic summary metrics for account %s...\n", accountID)

	// Define the queries for summary metrics
	queries := []struct {
		name      string
		query     string
		queryType string
	}{
		{
			name:      "SSL Certificates Expiring in 60 Days",
			query:     "SELECT count(*) AS 'value' FROM SyntheticCheck WHERE monitorExtendedType = 'CERT_CHECK' AND websiteDaysUntilExpiration <= 60 SINCE 1 hour ago",
			queryType: "count",
		},
		{
			name:      "Open Critical/High Incidents",
			query:     "SELECT count(*) AS 'value' FROM NrAiIncident WHERE priority IN ('critical', 'high') AND event = 'open' SINCE 1 hour ago",
			queryType: "count",
		},
		{
			name:      "CCU Usage by Account",
			query:     "FROM NrConsumption SELECT sum(consumption) AS 'value' WHERE metric = 'CoreCCU' AND dimension_dataCategory != 'LiveArchive' AND dimension_productCapability != 'IAST' AND dimension_productCapability NOT LIKE 'Entities & Relationships%' AND dimension_productFeature NOT LIKE 'Entity%' AND dimension_computeType NOT LIKE 'Entity%' FACET consumingAccountName SINCE 1 month ago",
			queryType: "faceted",
		},
	}

	summaryData := []struct {
		Metric string
		Value  string
	}{}

	for _, queryInfo := range queries {
		fmt.Printf("Executing: %s...\n", queryInfo.name)

		options := newrelic.LogsQueryOptions{
			Query: queryInfo.query,
		}

		stopSpinner := ui.StartSpinner("Querying...")
		result, err := client.QueryLogs(options)
		stopSpinner()
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			summaryData = append(summaryData, struct {
				Metric string
				Value  string
			}{
				Metric: queryInfo.name,
				Value:  fmt.Sprintf("Error: %v", err),
			})
			continue
		}

		if queryInfo.queryType == "count" {
			// For count queries, extract the single value
			if len(result.Results) > 0 {
				log := result.Results[0]
				value := 0
				for k, v := range log.Attributes {
					if k == "value" || k == "count" {
						if num, ok := v.(float64); ok {
							value = int(num)
						}
					}
				}
				summaryData = append(summaryData, struct {
					Metric string
					Value  string
				}{
					Metric: queryInfo.name,
					Value:  fmt.Sprintf("%d", value),
				})
			} else {
				summaryData = append(summaryData, struct {
					Metric string
					Value  string
				}{
					Metric: queryInfo.name,
					Value:  "No data available",
				})
			}
		} else if queryInfo.queryType == "faceted" {
			// For faceted queries, show top accounts
			if len(result.Results) > 0 {
				// Sort by value (simplified - would need proper sorting)
				topCount := 5
				if len(result.Results) < topCount {
					topCount = len(result.Results)
				}

				for i := 0; i < topCount; i++ {
					log := result.Results[i]
					accountName := "Unknown"
					value := 0

					for k, v := range log.Attributes {
						if k == "consumingAccountName" || k == "facet" {
							if str, ok := v.(string); ok {
								accountName = str
							}
						}
						if k == "value" {
							if num, ok := v.(float64); ok {
								value = int(num)
							}
						}
					}

					metricName := queryInfo.name
					if i > 0 {
						metricName = ""
					}

					summaryData = append(summaryData, struct {
						Metric string
						Value  string
					}{
						Metric: metricName,
						Value:  fmt.Sprintf("%s: %d CCUs", accountName, value),
					})
				}
			} else {
				summaryData = append(summaryData, struct {
					Metric string
					Value  string
				}{
					Metric: queryInfo.name,
					Value:  "No data available",
				})
			}
		}
	}

	// Display results in a table
	fmt.Println("\nNew Relic Summary Report:")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("%-40s %s\n", "METRIC", "VALUE")
	fmt.Println(strings.Repeat("-", 50))

	for _, data := range summaryData {
		fmt.Printf("%-40s %s\n", data.Metric, data.Value)
	}

	return nil
}
