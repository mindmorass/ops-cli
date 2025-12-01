package docker

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Docker Desktop application",
		Long: `Start the Docker Desktop application (daemon).

On macOS, this launches Docker Desktop. The command will wait for Docker to become available.

Examples:
  ops-cli docker start`,
		RunE: runStart,
	}

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("docker start is only supported on macOS")
	}

	// Check if Docker is already running
	client, err := NewDockerClient()
	if err == nil {
		if client.IsAvailable() {
			fmt.Println("Docker Desktop is already running")
			client.Close()
			return nil
		}
		client.Close()
	}

	// Start Docker Desktop using open command
	stopSpinner := ui.StartSpinner("Starting Docker Desktop...")
	startCmd := exec.Command("open", "-a", "Docker")
	err = startCmd.Run()
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to start Docker Desktop: %w", err)
	}

	// Wait for Docker to become available
	stopSpinner = ui.StartSpinner("Waiting for Docker Desktop to be ready...")

	maxWait := 60      // seconds
	checkInterval := 2 // seconds
	waited := 0

	for waited < maxWait {
		time.Sleep(time.Duration(checkInterval) * time.Second)
		waited += checkInterval

		client, err := NewDockerClient()
		if err == nil {
			if client.IsAvailable() {
				stopSpinner()
				client.Close()
				fmt.Println("Docker Desktop started successfully!")
				return nil
			}
			client.Close()
		}
	}

	stopSpinner()
	return fmt.Errorf("Docker Desktop did not become available within %d seconds. Please check Docker Desktop manually", maxWait)
}
