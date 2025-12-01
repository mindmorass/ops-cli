package main

import (
	"fmt"
	"strings"

	"github.com/ops-cli/internal/api/newrelic"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newEntitiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entities",
		Short: "List New Relic entities with optional filtering",
		Long: `List New Relic entities with optional filtering by type.

Examples:
  ops-cli newrelic entities
  ops-cli newrelic entities --filter APPLICATION
  ops-cli newrelic entities --filter HOST --max-results 500`,
		RunE: runEntities,
	}

	cmd.Flags().String("filter", "", "Filter entities by type (APPLICATION, HOST, CONTAINER, etc.)")
	cmd.Flags().Int("max-results", 100, "Maximum number of entities to fetch")

	return cmd
}

func runEntities(cmd *cobra.Command, args []string) error {
	client, err := getNewRelicClient()
	if err != nil {
		return err
	}

	filter, _ := cmd.Flags().GetString("filter")
	maxResults, _ := cmd.Flags().GetInt("max-results")

	if maxResults <= 0 {
		return fmt.Errorf("--max-results must be a positive number")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	accountID := cfg.NewRelic.AccountID
	if accountID == "" {
		return fmt.Errorf("account ID not configured")
	}

	fmt.Printf("Fetching New Relic entities for account %s", accountID)
	if filter != "" {
		fmt.Printf(" (filtering by: %s)", filter)
	}
	if maxResults != 100 {
		fmt.Printf(" (max results: %d)", maxResults)
	}
	fmt.Println("...")

	options := newrelic.EntitiesOptions{
		Filter:     filter,
		MaxResults: maxResults,
	}

	stopSpinner := ui.StartSpinner("Fetching entities...")
	entities, err := client.GetEntities(options)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get entities: %w", err)
	}

	if len(entities) == 0 {
		fmt.Println("\nNo entities found.")
		return nil
	}

	foundText := fmt.Sprintf("Found %d entities", len(entities))
	if len(entities) == maxResults {
		foundText += " (limited)"
	}
	fmt.Printf("\n%s:\n\n", foundText)

	// Display entities in a table format
	fmt.Printf("%-40s %-20s %-15s %-20s\n", "NAME", "TYPE", "DOMAIN", "STATUS")
	fmt.Println(strings.Repeat("-", 95))

	for _, entity := range entities {
		name := entity.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}
		entityType := entity.EntityType
		if entityType == "" {
			entityType = "Unknown"
		}
		domain := entity.Domain
		if domain == "" {
			domain = "Unknown"
		}
		status := entity.AlertSeverity
		if status == "" {
			status = "Not Configured"
		}
		fmt.Printf("%-40s %-20s %-15s %-20s\n", name, entityType, domain, status)
	}

	if len(entities) == maxResults && maxResults < 1000 {
		fmt.Printf("\nTip: Use --max-results <number> to fetch more entities (current limit: %d)\n", maxResults)
	}

	return nil
}
