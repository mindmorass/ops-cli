package docker

import (
	"encoding/json"
	"fmt"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show Docker system information",
		Long: `Show Docker system information including version, system stats, and configuration.

Examples:
  ops-cli docker info
  ops-cli docker info --verbose`,
		RunE: runInfo,
	}

	cmd.Flags().Bool("verbose", false, "Show detailed information")

	return cmd
}

func runInfo(cmd *cobra.Command, args []string) error {
	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if !client.IsAvailable() {
		return fmt.Errorf("Docker is not available. Make sure Docker is running")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	// Use spinner while fetching info
	stopSpinner := ui.StartSpinner("Fetching Docker information...")
	defer stopSpinner()

	// Get version and system info
	version, err := client.GetVersion()
	if err != nil {
		stopSpinner()
		return fmt.Errorf("failed to get Docker version: %w", err)
	}

	systemInfo, err := client.GetSystemInfo()
	if err != nil {
		stopSpinner()
		return fmt.Errorf("failed to get system info: %w", err)
	}

	stopSpinner()

	if verbose {
		// JSON output for verbose mode
		output := map[string]interface{}{
			"version": version,
			"system":  systemInfo,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Formatted output
	fmt.Println("\nDocker System Information")
	fmt.Println("============================")

	fmt.Println("\nVersion Info:")
	fmt.Printf("   Docker Version: %s\n", version.Version)
	fmt.Printf("   API Version: %s\n", version.APIVersion)
	fmt.Printf("   Go Version: %s\n", version.GoVersion)
	fmt.Printf("   Git Commit: %s\n", version.GitCommit)
	fmt.Printf("   Built: %s\n", version.BuildTime)
	fmt.Printf("   OS/Arch: %s/%s\n", version.Os, version.Arch)

	fmt.Println("\nSystem Stats:")
	fmt.Printf("   Total Containers: %d\n", systemInfo.Containers)
	fmt.Printf("   Running: %d\n", systemInfo.ContainersRunning)
	fmt.Printf("   Stopped: %d\n", systemInfo.ContainersStopped)
	fmt.Printf("   Total Images: %d\n", systemInfo.Images)
	fmt.Printf("   Memory Limit: %v\n", systemInfo.MemoryLimit)
	fmt.Printf("   CPUs: %d\n", systemInfo.NCPU)
	fmt.Printf("   Operating System: %s\n", systemInfo.OperatingSystem)
	fmt.Printf("   Architecture: %s\n", systemInfo.Architecture)

	return nil
}
