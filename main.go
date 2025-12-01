package main

import (
	"fmt"
	"os"

	"github.com/ops-cli/internal/cli"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ops-cli",
		Short: "Modular API Command Line Tool",
		Long: `A Go CLI tool that provides modular access to various APIs.
The architecture follows the Cobra pattern with <module> <subcommand> structure for easy extensibility.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Register all modules
	cli.RegisterModules(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
