package devtools

import (
	"fmt"
	"runtime"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check status of development tools",
		Long: `Check status of development tools configured in config.toml.

Examples:
  ops-cli devtools check
  ops-cli devtools check --verbose`,
		RunE: runCheck,
	}

	cmd.Flags().Bool("verbose", false, "Show detailed output")

	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Check if we're on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("devtools module only supports macOS")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	// Initialize Homebrew client
	brew, err := NewHomebrewClient()
	if err != nil {
		return err
	}

	// Load tools from config
	tools, err := LoadDevToolsConfig()
	if err != nil {
		return fmt.Errorf("failed to load devtools config: %w", err)
	}

	if len(tools) == 0 {
		fmt.Println("No development tools configured in config.toml")
		fmt.Println("Add tools to the [development_tools] section in config.toml")
		return nil
	}

	fmt.Println("\nChecking development tools status (Homebrew)")

	type ToolStatus struct {
		Name    string
		Status  string
		Version string
	}

	results := []ToolStatus{}

	// Use spinner for checking tools
	stopSpinner := ui.StartSpinner("Checking tools...")
	defer stopSpinner()

	for _, tool := range tools {
		if verbose {
			stopSpinner()
			fmt.Printf("Checking %s (%s)...\n", tool.Name, tool.Homebrew)
			stopSpinner = ui.StartSpinner("Checking tools...")
		}

		installed, version, err := brew.IsInstalled(tool.Homebrew)
		if err != nil {
			results = append(results, ToolStatus{
				Name:   tool.Name,
				Status: "Unknown",
			})
			continue
		}

		if installed {
			status := "Installed"
			if version != "" {
				results = append(results, ToolStatus{
					Name:    tool.Name,
					Status:  status,
					Version: version,
				})
			} else {
				results = append(results, ToolStatus{
					Name:   tool.Name,
					Status: status,
				})
			}
		} else {
			results = append(results, ToolStatus{
				Name:   tool.Name,
				Status: "Not Installed",
			})
		}
	}

	stopSpinner()
	fmt.Println()

	// Display results
	fmt.Println("Development Tools Status:")
	fmt.Println()

	for _, result := range results {
		versionInfo := ""
		if result.Version != "" {
			versionInfo = fmt.Sprintf(" (%s)", result.Version)
		}
		fmt.Printf("  %-20s %s%s\n", result.Name, result.Status, versionInfo)
	}

	// Summary
	installed := 0
	for _, result := range results {
		if result.Status == "Installed" {
			installed++
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d/%d tools installed\n", installed, len(results))

	if installed < len(results) {
		fmt.Println()
		fmt.Println("Run 'ops-cli devtools install' to install missing tools")
	}

	return nil
}
