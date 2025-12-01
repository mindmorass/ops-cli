package docker

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop Docker Desktop application",
		Long: `Stop the Docker Desktop application (daemon).

On macOS, this stops Docker Desktop and all running containers.

Examples:
  ops-cli docker stop`,
		RunE: runStop,
	}

	return cmd
}

func runStop(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("docker stop is only supported on macOS")
	}

	// Check if Docker is running
	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("Docker Desktop is not running or not available")
	}

	if !client.IsAvailable() {
		client.Close()
		return fmt.Errorf("Docker Desktop is not running")
	}

	client.Close()

	// Try using docker desktop stop command first (if available)
	stopSpinner := ui.StartSpinner("Stopping Docker Desktop...")
	stopCmd := exec.Command("docker", "desktop", "stop")
	err = stopCmd.Run()
	stopSpinner()
	if err == nil {
		// Verify it's actually stopped
		verifySpinner := ui.StartSpinner("Verifying Docker Desktop has stopped...")
		time.Sleep(2 * time.Second)
		client, err := NewDockerClient()
		verifySpinner()
		if err != nil || !client.IsAvailable() {
			if client != nil {
				client.Close()
			}
			fmt.Println("Docker Desktop stopped successfully!")
			return nil
		}
		client.Close()
		// If still available, fall through to alternative method
	}

	// Fallback: Use pkill to stop Docker Desktop processes
	// This is a more forceful method
	stopSpinner = ui.StartSpinner("Stopping Docker Desktop (alternative method)...")

	// Stop Docker Desktop processes
	pkillCmd := exec.Command("pkill", "-f", "Docker")
	err = pkillCmd.Run()
	if err != nil {
		stopSpinner()
		// pkill returns error if no processes found, which might be okay
		// Check if Docker is actually stopped
		verifySpinner := ui.StartSpinner("Verifying Docker Desktop has stopped...")
		time.Sleep(2 * time.Second)
		client, err := NewDockerClient()
		verifySpinner()
		if err != nil || !client.IsAvailable() {
			if client != nil {
				client.Close()
			}
			fmt.Println("Docker Desktop stopped successfully!")
			return nil
		}
		client.Close()
		return fmt.Errorf("failed to stop Docker Desktop. Please stop it manually from the menu bar")
	}

	stopSpinner()

	// Wait a moment and verify it's stopped
	verifySpinner := ui.StartSpinner("Verifying Docker Desktop has stopped...")
	time.Sleep(2 * time.Second)
	client, err = NewDockerClient()
	verifySpinner()
	if err != nil || !client.IsAvailable() {
		if client != nil {
			client.Close()
		}
		fmt.Println("Docker Desktop stopped successfully!")
		return nil
	}
	client.Close()

	return fmt.Errorf("Docker Desktop may still be running. Please check manually")
}
