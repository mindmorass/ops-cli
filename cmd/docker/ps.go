package docker

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List containers",
		Long: `List Docker containers.

Examples:
  ops-cli docker ps
  ops-cli docker ps --all
  ops-cli docker ps --format json`,
		RunE: runPs,
	}

	cmd.Flags().Bool("all", false, "Show all containers (including stopped)")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runPs(cmd *cobra.Command, args []string) error {
	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if !client.IsAvailable() {
		return fmt.Errorf("Docker is not available. Make sure Docker is running")
	}

	showAll, _ := cmd.Flags().GetBool("all")
	format, _ := cmd.Flags().GetString("format")

	// Use spinner while fetching containers
	stopSpinner := ui.StartSpinner("Fetching containers...")
	containers, err := client.ListContainers(showAll)
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("No containers found.")
		return nil
	}

	if format == "json" {
		data, err := json.MarshalIndent(containers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Table format
	fmt.Println("\nContainers:")
	fmt.Println("==============")

	for _, container := range containers {
		status := "[running]"
		if container.State != "running" {
			status = "[stopped]"
		}

		name := "unnamed"
		if len(container.Names) > 0 {
			name = strings.TrimPrefix(container.Names[0], "/")
		}

		fmt.Printf("%s %s (%s)\n", status, name, container.ID[:12])
		fmt.Printf("   Image: %s\n", container.Image)
		fmt.Printf("   State: %s\n", container.State)
		fmt.Printf("   Status: %s\n", container.Status)

		if container.State == "running" {
			uptime := time.Since(time.Unix(container.Created, 0))
			fmt.Printf("   Uptime: %s\n", formatUptime(uptime))
		}

		if len(container.Ports) > 0 {
			ports := []string{}
			for _, port := range container.Ports {
				if port.PublicPort > 0 {
					ports = append(ports, fmt.Sprintf("%d:%d/%s", port.PublicPort, port.PrivatePort, port.Type))
				} else {
					ports = append(ports, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
				}
			}
			fmt.Printf("   Ports: %s\n", strings.Join(ports, ", "))
		}
		fmt.Println()
	}

	return nil
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
