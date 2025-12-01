package docker

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show container resource usage statistics",
		Long: `Show container resource usage statistics including CPU and memory usage.

Examples:
  ops-cli docker stats`,
		RunE: runStats,
	}

	return cmd
}

func runStats(cmd *cobra.Command, args []string) error {
	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if !client.IsAvailable() {
		return fmt.Errorf("Docker is not available. Make sure Docker is running")
	}

	runningContainers, err := client.GetRunningContainers()
	if err != nil {
		return fmt.Errorf("failed to get running containers: %w", err)
	}

	if len(runningContainers) == 0 {
		fmt.Println("No running containers to show stats for.")
		return nil
	}

	fmt.Println("\nContainer Resource Usage:")
	fmt.Println("============================")

	for _, container := range runningContainers {
		stopSpinner := ui.StartSpinner(fmt.Sprintf("Fetching stats for %s...", getContainerName(container)))
		stats, err := client.GetContainerStats(container.ID)
		stopSpinner()
		if err != nil {
			fmt.Printf("%s\n", getContainerName(container))
			fmt.Printf("   Stats unavailable: %v\n", err)
			fmt.Println()
			continue
		}

		// Calculate CPU percentage
		cpuPercent := calculateCPUPercent(stats)

		// Calculate memory usage
		memoryUsage := stats.MemoryStats.Usage
		memoryLimit := stats.MemoryStats.Limit
		memoryPercent := float64(0)
		if memoryLimit > 0 {
			memoryPercent = (float64(memoryUsage) / float64(memoryLimit)) * 100
		}
		memoryUsageMB := float64(memoryUsage) / 1024 / 1024
		memoryLimitMB := float64(memoryLimit) / 1024 / 1024

		fmt.Printf("%s\n", getContainerName(container))
		fmt.Printf("   CPU: %.2f%%\n", cpuPercent)
		fmt.Printf("   Memory: %.2fMB / %.2fMB (%.2f%%)\n", memoryUsageMB, memoryLimitMB, memoryPercent)
		fmt.Println()
	}

	return nil
}

func getContainerName(container types.Container) string {
	if len(container.Names) > 0 {
		return container.Names[0]
	}
	return container.ID[:12]
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	if stats.CPUStats.CPUUsage.TotalUsage == 0 || stats.PreCPUStats.CPUUsage.TotalUsage == 0 {
		return 0.0
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent := (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		return cpuPercent
	}
	return 0.0
}
