package devtools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// HomebrewClient handles Homebrew operations
type HomebrewClient struct {
	available bool
}

// NewHomebrewClient creates a new Homebrew client
func NewHomebrewClient() (*HomebrewClient, error) {
	// Check if we're on macOS
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("devtools module only supports macOS")
	}

	// Check if Homebrew is available
	cmd := exec.Command("brew", "--version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Homebrew is not installed. Please install it from https://brew.sh")
	}

	return &HomebrewClient{
		available: true,
	}, nil
}

// IsInstalled checks if a package is installed
func (h *HomebrewClient) IsInstalled(packageName string) (bool, string, error) {
	cmd := exec.Command("brew", "list", "--versions", packageName)
	output, err := cmd.Output()
	if err != nil {
		// If command fails, package is likely not installed
		return false, "", nil
	}

	// Parse version from output (format: "package version")
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return true, parts[1], nil
	}
	return true, "", nil
}

// AddTap adds a Homebrew tap
func (h *HomebrewClient) AddTap(tapName string) error {
	cmd := exec.Command("brew", "tap", tapName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if tap is already added (this is not an error)
		if strings.Contains(string(output), "already tapped") {
			return nil
		}
		return fmt.Errorf("failed to add tap %s: %s", tapName, string(output))
	}
	return nil
}

// IsTapInstalled checks if a tap is already installed
func (h *HomebrewClient) IsTapInstalled(tapName string) (bool, error) {
	cmd := exec.Command("brew", "tap")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list taps: %w", err)
	}

	// Check if tap is in the list
	taps := strings.Split(string(output), "\n")
	for _, tap := range taps {
		if strings.TrimSpace(tap) == tapName {
			return true, nil
		}
	}
	return false, nil
}

// InstallPackage installs a package using Homebrew
func (h *HomebrewClient) InstallPackage(packageName string) error {
	cmd := exec.Command("brew", "install", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %s", packageName, string(output))
	}
	return nil
}

// InstallPackages installs multiple packages
func (h *HomebrewClient) InstallPackages(packageNames []string) []InstallResult {
	results := make([]InstallResult, len(packageNames))
	for i, pkg := range packageNames {
		err := h.InstallPackage(pkg)
		results[i] = InstallResult{
			Package: pkg,
			Success: err == nil,
			Error:   err,
		}
	}
	return results
}

// InstallResult represents the result of an installation
type InstallResult struct {
	Package string
	Success bool
	Error   error
}

// GetInstalledPackages returns a list of all installed packages
func (h *HomebrewClient) GetInstalledPackages() ([]string, error) {
	cmd := exec.Command("brew", "list", "--formula")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	packages := strings.Fields(string(output))
	return packages, nil
}

// GetPackageInfo returns detailed information about a package
func (h *HomebrewClient) GetPackageInfo(packageName string) (*PackageInfo, error) {
	cmd := exec.Command("brew", "info", "--json", packageName)
	output, err := cmd.Output()
	if err != nil {
		return &PackageInfo{
			Name:      packageName,
			Installed: false,
		}, nil
	}

	var info []struct {
		Name      string `json:"name"`
		Installed bool   `json:"installed"`
		Versions  struct {
			Stable string `json:"stable"`
		} `json:"versions"`
		Desc string `json:"desc"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return &PackageInfo{
			Name:      packageName,
			Installed: false,
		}, nil
	}

	if len(info) > 0 {
		return &PackageInfo{
			Name:        info[0].Name,
			Installed:   info[0].Installed,
			Version:     info[0].Versions.Stable,
			Description: info[0].Desc,
		}, nil
	}

	return &PackageInfo{
		Name:      packageName,
		Installed: false,
	}, nil
}

// PackageInfo contains information about a package
type PackageInfo struct {
	Name        string
	Installed   bool
	Version     string
	Description string
}
