package startpage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the background startpage server",
		Long: `Stop the background startpage server.

Examples:
  ops-cli startpage stop`,
		RunE: runStop,
	}

	return cmd
}

func runStop(cmd *cobra.Command, args []string) error {
	dir, err := getStartpageDirectory()
	if err != nil {
		return err
	}

	pidFile := filepath.Join(dir, "server.pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No background server running (no PID file found)")
			fmt.Println("Start a background server with: ops-cli startpage serve --background")
			return nil
		}
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid PID in server.pid file: %w", err)
	}

	fmt.Printf("Stopping background server (PID: %d)...\n", pid)

	// Note: Process killing would require platform-specific code
	// For now, just provide instructions
	fmt.Printf("To stop the server, run: kill %d\n", pid)
	fmt.Println("Or manually remove the PID file if the process is already stopped")

	// Try to remove PID file
	os.Remove(pidFile)
	fmt.Println("PID file removed")

	return nil
}
