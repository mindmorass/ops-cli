package devtools

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install development tools",
		Long: `Install development tools using Homebrew.

Examples:
  ops-cli devtools install
  ops-cli devtools install --all
  ops-cli devtools install --yes`,
		RunE: runInstall,
	}

	cmd.Flags().Bool("all", false, "Install all missing tools")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Check if we're on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("devtools module only supports macOS")
	}

	installAll, _ := cmd.Flags().GetBool("all")
	skipConfirm, _ := cmd.Flags().GetBool("yes")

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

	fmt.Println("\nInstalling development tools (Homebrew)")
	fmt.Println()

	// Check which tools are missing
	type ToolChoice struct {
		Key         string
		Tool        DevelopmentTool
		Installed   bool
		PackageName string
	}

	choices := []ToolChoice{}

	for key, tool := range tools {
		installed, _, _ := brew.IsInstalled(tool.Homebrew)
		choices = append(choices, ToolChoice{
			Key:         key,
			Tool:        tool,
			Installed:   installed,
			PackageName: tool.Homebrew,
		})
	}

	// Filter to only missing tools
	missingTools := []ToolChoice{}
	for _, choice := range choices {
		if !choice.Installed {
			missingTools = append(missingTools, choice)
		}
	}

	if len(missingTools) == 0 {
		fmt.Println("All configured tools are already installed!")
		return nil
	}

	// Select tools to install
	var selectedTools []ToolChoice

	if installAll {
		selectedTools = missingTools
		fmt.Printf("Installing all missing tools: %d tools\n", len(selectedTools))
	} else {
		// Interactive selection
		fmt.Println("Select tools to install:")
		fmt.Println()

		selectedIndices := []int{}
		reader := bufio.NewReader(os.Stdin)

		for i, tool := range missingTools {
			fmt.Printf("%d. %s - %s\n", i+1, tool.Tool.Name, tool.Tool.Description)
		}
		fmt.Println()
		fmt.Print("Enter tool numbers (comma-separated) or 'all' for all: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "all" {
			selectedTools = missingTools
		} else {
			// Parse comma-separated numbers
			parts := strings.Split(input, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				var idx int
				if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
					if idx >= 1 && idx <= len(missingTools) {
						selectedIndices = append(selectedIndices, idx-1)
					}
				}
			}

			if len(selectedIndices) == 0 {
				fmt.Println("No tools selected. Installation cancelled.")
				return nil
			}

			for _, idx := range selectedIndices {
				selectedTools = append(selectedTools, missingTools[idx])
			}
		}
	}

	if len(selectedTools) == 0 {
		fmt.Println("No tools selected for installation.")
		return nil
	}

	// Confirm installation
	if !skipConfirm {
		fmt.Println()
		fmt.Printf("The following tools will be installed:\n")
		for _, tool := range selectedTools {
			fmt.Printf("  - %s (%s)\n", tool.Tool.Name, tool.PackageName)
		}
		fmt.Println()
		fmt.Print("Proceed with installation? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Installation cancelled.")
			return nil
		}
	}

	// Install tools
	fmt.Printf("\nInstalling %d development tools...\n", len(selectedTools))

	// First, add any required taps
	tapsToAdd := make(map[string]bool)
	for _, tool := range selectedTools {
		if tool.Tool.HomebrewTap != "" {
			tapsToAdd[tool.Tool.HomebrewTap] = true
		}
	}

	if len(tapsToAdd) > 0 {
		fmt.Println("Adding required Homebrew taps...")
		for tap := range tapsToAdd {
			installed, err := brew.IsTapInstalled(tap)
			if err != nil {
				fmt.Printf("Warning: failed to check tap %s: %v\n", tap, err)
				continue
			}
			if !installed {
				fmt.Printf("  Adding tap: %s\n", tap)
				if err := brew.AddTap(tap); err != nil {
					fmt.Printf("Warning: failed to add tap %s: %v\n", tap, err)
				} else {
					fmt.Printf("  ✓ Added tap: %s\n", tap)
				}
			} else {
				fmt.Printf("  ✓ Tap already installed: %s\n", tap)
			}
		}
		fmt.Println()
	}

	packageNames := make([]string, len(selectedTools))
	for i, tool := range selectedTools {
		packageNames[i] = tool.PackageName
	}

	// Use spinner during installation
	stopSpinner := ui.StartSpinner("Installing packages...")
	results := brew.InstallPackages(packageNames)
	stopSpinner()

	// Display results
	fmt.Println("\nInstallation Summary:")
	fmt.Println()

	successful := []InstallResult{}
	failed := []InstallResult{}

	for _, result := range results {
		if result.Success {
			successful = append(successful, result)
		} else {
			failed = append(failed, result)
		}
	}

	if len(successful) > 0 {
		fmt.Println("Successfully installed:")
		for _, result := range successful {
			// Find tool name
			toolName := result.Package
			for _, tool := range selectedTools {
				if tool.PackageName == result.Package {
					toolName = tool.Tool.Name
					break
				}
			}
			fmt.Printf("   %s\n", toolName)
		}
	}

	if len(failed) > 0 {
		fmt.Println("\nFailed to install:")
		for _, result := range failed {
			// Find tool name
			toolName := result.Package
			for _, tool := range selectedTools {
				if tool.PackageName == result.Package {
					toolName = tool.Tool.Name
					break
				}
			}
			errorMsg := "unknown error"
			if result.Error != nil {
				errorMsg = result.Error.Error()
			}
			fmt.Printf("   %s: %s\n", toolName, errorMsg)
		}
	}

	fmt.Printf("\nSummary: %d/%d tools installed successfully\n", len(successful), len(results))

	return nil
}
