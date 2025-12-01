package startpage

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the startpage for production",
		Long: `Build the startpage for production.

Examples:
  ops-cli startpage build`,
		RunE: runBuild,
	}

	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
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

	fmt.Println("Building startpage...")

	// Generate the OpenLinks.json file
	if err := generateStartpage(config); err != nil {
		return fmt.Errorf("failed to generate startpage: %w", err)
	}

	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = config.Path
	buildCmd.Stdout = cmd.OutOrStdout()
	buildCmd.Stderr = cmd.ErrOrStderr()

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("Build completed successfully!")
	fmt.Printf("Output: %s\n", filepath.Join(config.Path, "dist"))
	fmt.Println("\nTo use your startpage:")
	fmt.Println("   1. Run: ops-cli startpage serve")
	fmt.Println("   2. Set browser homepage to: http://localhost:3000")
	fmt.Println("\nNote: Startpage requires a web server to work properly")

	return nil
}
