package main

import (
	"github.com/spf13/cobra"
)

// Plugin must be exported for plugin loading
var Plugin = &WindowsPlugin{}

type WindowsPlugin struct{}

func (p *WindowsPlugin) Name() string {
	return "windows"
}

func (p *WindowsPlugin) Version() string {
	return "1.0.0"
}

func (p *WindowsPlugin) Register(rootCmd *cobra.Command) error {
	windowsCmd := &cobra.Command{
		Use:   "windows",
		Short: "Windows remote management via WinRM",
		Long: `Windows Remote Management

Usage: ops-cli windows <subcommand> [options] [args]

The Windows module provides comprehensive remote management capabilities for Windows systems
via Windows Remote Management (WinRM). Features include:

• PowerShell Command Execution
  - Execute single commands or scripts
  - Interactive PowerShell sessions
  - Predefined system administration commands

• Service Management
  - List, start, stop, and restart Windows services
  - Monitor service status changes
  - Filter services by status or name

• Process Management
  - List and monitor running processes
  - Terminate processes safely
  - View process details and relationships

• File Operations
  - Browse directories and files
  - Read, create, and delete files
  - Copy files and create directories

• System Information
  - Hardware and software inventory
  - Disk usage and network configuration
  - Event log monitoring

Requirements:
  - WinRM must be enabled on target Windows systems
  - Valid Windows credentials with appropriate permissions
  - Network connectivity to WinRM ports (5985/5986)`,
	}

	// Register subcommands
	registerWindowsCommands(windowsCmd)

	rootCmd.AddCommand(windowsCmd)
	return nil
}

