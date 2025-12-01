package startpage

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func newDevCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start the Astro development server",
		Long: `Start the Astro development server.

Examples:
  ops-cli startpage dev`,
		RunE: runDev,
	}

	return cmd
}

func runDev(cmd *cobra.Command, args []string) error {
	if err := ensureStartpageDirectory(); err != nil {
		return fmt.Errorf("failed to create startpage directory: %w", err)
	}

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no startpage found. Run 'ops-cli startpage init' first")
	}

	fmt.Println("Starting Astro development server...")

	devCmd := exec.Command("npm", "run", "dev")
	devCmd.Dir = config.Path
	devCmd.Stdout = cmd.OutOrStdout()
	devCmd.Stderr = cmd.ErrOrStderr()
	return devCmd.Run()
}
