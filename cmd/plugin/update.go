package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ghapi "github.com/google/go-github/v55/github"
	ghclient "github.com/ops-cli/internal/api/github"
	"github.com/ops-cli/internal/config"
	"github.com/ops-cli/internal/plugin"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

const (
	defaultRepoOwner = "ops-cli"
	defaultRepoName  = "ops-cli"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [flags]",
		Short: "Update ops-cli and plugins from GitHub releases",
		Long: `Check for and install updates to ops-cli and plugins from GitHub releases.

This command:
  - Checks the latest release on GitHub
  - Compares with the current version
  - Downloads and installs updates for the main binary and plugins
  - Verifies checksums if available

Examples:
  # Check for updates
  ops-cli plugin update --check

  # Update everything
  ops-cli plugin update

  # Update only plugins
  ops-cli plugin update --plugins-only

  # Update only the main binary
  ops-cli plugin update --binary-only`,
		RunE: runUpdate,
	}

	cmd.Flags().Bool("check", false, "Only check for updates, don't install")
	cmd.Flags().Bool("plugins-only", false, "Update only plugins")
	cmd.Flags().Bool("binary-only", false, "Update only the main binary")
	cmd.Flags().String("repo", fmt.Sprintf("%s/%s", defaultRepoOwner, defaultRepoName), "Repository in format owner/repo")
	cmd.Flags().String("version", "", "Specific version to install (default: latest)")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	checkOnly, _ := cmd.Flags().GetBool("check")
	pluginsOnly, _ := cmd.Flags().GetBool("plugins-only")
	binaryOnly, _ := cmd.Flags().GetBool("binary-only")
	repoFlag, _ := cmd.Flags().GetString("repo")
	versionFlag, _ := cmd.Flags().GetString("version")

	// Get GitHub client first to check config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Try to get repository from config if not provided
	if repoFlag == fmt.Sprintf("%s/%s", defaultRepoOwner, defaultRepoName) {
		if cfg.GitHub != nil && cfg.GitHub.DefaultOwner != "" {
			// Try to construct from default owner (would need repo name too)
			// For now, we'll try to detect from git remote
			if detectedRepo := detectRepoFromGit(); detectedRepo != "" {
				repoFlag = detectedRepo
			}
		} else {
			// Try to detect from git remote
			if detectedRepo := detectRepoFromGit(); detectedRepo != "" {
				repoFlag = detectedRepo
			}
		}
	}

	// Parse repository
	parts := strings.Split(repoFlag, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format. Use 'owner/repo' or set via --repo flag")
	}
	owner, repo := parts[0], parts[1]

	var token string
	if cfg.GitHub != nil {
		token = cfg.GitHub.Token
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	ghClient := ghclient.NewClient(token)

	// Get current version
	currentVersion := getCurrentVersion()
	fmt.Printf("Current version: %s\n", currentVersion)

	// Get release
	var release *ghapi.RepositoryRelease
	var releaseErr error

	stopSpinner := ui.StartSpinner("Fetching release information...")
	if versionFlag != "" {
		release, releaseErr = ghClient.GetReleaseByTag(owner, repo, versionFlag)
	} else {
		release, releaseErr = ghClient.GetLatestRelease(owner, repo)
	}
	stopSpinner()

	if releaseErr != nil {
		// Check if it's a 404 error (repository not found or no releases)
		errStr := releaseErr.Error()
		if strings.Contains(errStr, "404") {
			return fmt.Errorf(`repository "%s/%s" not found or has no releases

This could mean:
  - The repository doesn't exist on GitHub
  - The repository exists but has no releases yet
  - You don't have access to the repository

To fix this:
  1. Ensure the repository exists: https://github.com/%s/%s
  2. Create a release if none exist
  3. Use --repo flag to specify a different repository: ops-cli plugin update --repo owner/repo`,
				owner, repo, owner, repo)
		}
		return fmt.Errorf("failed to get release from %s/%s: %w", owner, repo, releaseErr)
	}


	releaseVersion := release.GetTagName()
	if strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = releaseVersion[1:]
	}

	fmt.Printf("Latest release: %s\n", release.GetTagName())

	// Compare versions
	if !checkOnly && !isNewerVersion(releaseVersion, currentVersion) {
		fmt.Println("✓ You are already on the latest version!")
		return nil
	}

	if checkOnly {
		if isNewerVersion(releaseVersion, currentVersion) {
			fmt.Printf("⚠ Update available: %s (current: %s)\n", release.GetTagName(), currentVersion)
			return nil
		}
		fmt.Println("✓ You are on the latest version!")
		return nil
	}

	// Download and install updates
	if !pluginsOnly {
		if err := updateBinary(ghClient, release, owner, repo); err != nil {
			return fmt.Errorf("failed to update binary: %w", err)
		}
	}

	if !binaryOnly {
		if err := updatePlugins(ghClient, release); err != nil {
			return fmt.Errorf("failed to update plugins: %w", err)
		}
	}

	fmt.Println("\n✓ Update completed successfully!")
	return nil
}

func getCurrentVersion() string {
	// Try to get version from the binary itself
	// The version is set at build time via ldflags
	// For now, we'll use a simple approach - check if we can get it from the binary
	// In a real scenario, you might want to parse it from `ops-cli --version`
	cmd := exec.Command("ops-cli", "--version")
	output, err := cmd.Output()
	if err == nil {
		versionStr := strings.TrimSpace(string(output))
		// Extract version from output like "ops-cli version 1.0.0 (commit: abc123, built: 2024-01-01)"
		parts := strings.Fields(versionStr)
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	// Fallback: try to get from git
	cmd = exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err = cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		if strings.HasPrefix(version, "v") {
			version = version[1:]
		}
		return version
	}

	return "dev"
}

func isNewerVersion(newVersion, currentVersion string) bool {
	// Simple version comparison
	// Remove 'v' prefix if present
	newVersion = strings.TrimPrefix(newVersion, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	// If current is "dev", always consider update available
	if currentVersion == "dev" || currentVersion == "unknown" {
		return true
	}

	// Simple string comparison for semantic versions
	// This is a basic implementation - for production, use a proper semver library
	return newVersion > currentVersion
}

func updateBinary(ghClient *ghclient.Client, release *ghapi.RepositoryRelease, owner, repo string) error {
	fmt.Println("\nUpdating main binary...")

	// Find the binary asset
	var binaryAsset *ghapi.ReleaseAsset
	for _, asset := range release.Assets {
		if asset.GetName() == "ops-cli" {
			binaryAsset = asset
			break
		}
	}

	if binaryAsset == nil {
		return fmt.Errorf("binary asset 'ops-cli' not found in release")
	}

	// Download binary
	stopSpinner := ui.StartSpinner("Downloading binary...")
	data, err := ghClient.DownloadReleaseAsset(owner, repo, binaryAsset)
	stopSpinner()

	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify checksum if available
	checksumAsset := findChecksumAsset(release.Assets, "ops-cli.sha256")
	if checksumAsset != nil {
		stopSpinner := ui.StartSpinner("Verifying checksum...")
		if err := verifyChecksum(data, checksumAsset, "ops-cli", owner, repo, ghClient); err != nil {
			stopSpinner()
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		stopSpinner()
	}

	// Get current binary path
	currentBinary, err := exec.LookPath("ops-cli")
	if err != nil {
		// Binary not in PATH, try to find it
		currentBinary = os.Args[0]
	}

	// Create backup
	backupPath := currentBinary + ".backup"
	if err := copyFile(currentBinary, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	defer os.Remove(backupPath) // Clean up backup on success

	// Install new binary
	stopSpinner = ui.StartSpinner("Installing binary...")
	if err := os.WriteFile(currentBinary, data, 0755); err != nil {
		stopSpinner()
		// Restore backup on failure
		copyFile(backupPath, currentBinary)
		return fmt.Errorf("failed to install binary: %w", err)
	}
	stopSpinner()

	fmt.Printf("✓ Binary updated successfully to %s\n", release.GetTagName())
	return nil
}

func updatePlugins(ghClient *ghclient.Client, release *ghapi.RepositoryRelease) error {
	// Extract owner/repo from release HTML URL or use defaults
	owner, repo := defaultRepoOwner, defaultRepoName
	if release.HTMLURL != nil {
		// Try to extract from HTML URL: https://github.com/owner/repo/releases/tag/v1.0.0
		url := release.GetHTMLURL()
		parts := strings.Split(url, "/")
		if len(parts) >= 4 {
			owner = parts[3]
			repo = parts[4]
		}
	}
	fmt.Println("\nUpdating plugins...")

	loader, err := plugin.NewLoader()
	if err != nil {
		return fmt.Errorf("failed to create plugin loader: %w", err)
	}

	pluginDir := loader.GetPluginDir()

	// Find plugin assets (files ending in .so in plugins/ directory)
	pluginAssets := make(map[string]*ghapi.ReleaseAsset)
	checksumAsset := findChecksumAsset(release.Assets, "plugins.sha256")

	for _, asset := range release.Assets {
		name := asset.GetName()
		if strings.HasSuffix(name, ".so") && strings.HasPrefix(name, "plugins/") {
			pluginName := strings.TrimSuffix(strings.TrimPrefix(name, "plugins/"), ".so")
			pluginAssets[pluginName] = asset
		}
	}

	if len(pluginAssets) == 0 {
		fmt.Println("  No plugin assets found in release")
		return nil
	}

	// Download checksums if available
	var checksums map[string]string
	if checksumAsset != nil {
		stopSpinner := ui.StartSpinner("Downloading plugin checksums...")
		data, err := ghClient.DownloadReleaseAsset(owner, repo, checksumAsset)
		stopSpinner()
		if err == nil {
			checksums = parseChecksums(string(data))
		}
	}

	// Download and install each plugin
	for pluginName, asset := range pluginAssets {
		fmt.Printf("  Updating plugin: %s\n", pluginName)

		stopSpinner := ui.StartSpinner(fmt.Sprintf("  Downloading %s...", pluginName))
		data, err := ghClient.DownloadReleaseAsset(owner, repo, asset)
		stopSpinner()

		if err != nil {
			fmt.Printf("  ⚠ Failed to download %s: %v\n", pluginName, err)
			continue
		}

		// Verify checksum if available
		expectedChecksum := checksums[asset.GetName()]
		if expectedChecksum != "" {
			actualChecksum := calculateSHA256(data)
			if actualChecksum != expectedChecksum {
				fmt.Printf("  ⚠ Checksum verification failed for %s\n", pluginName)
				continue
			}
		}

		// Install plugin
		pluginPath := filepath.Join(pluginDir, pluginName+".so")
		if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
			fmt.Printf("  ⚠ Failed to create plugin directory: %v\n", err)
			continue
		}

		if err := os.WriteFile(pluginPath, data, 0644); err != nil {
			fmt.Printf("  ⚠ Failed to install %s: %v\n", pluginName, err)
			continue
		}

		fmt.Printf("  ✓ %s updated successfully\n", pluginName)
	}

	return nil
}

func findChecksumAsset(assets []*ghapi.ReleaseAsset, filename string) *ghapi.ReleaseAsset {
	for _, asset := range assets {
		if asset.GetName() == filename {
			return asset
		}
	}
	return nil
}

func verifyChecksum(data []byte, checksumAsset *ghapi.ReleaseAsset, assetName, owner, repo string, ghClient *ghclient.Client) error {
	checksumData, err := ghClient.DownloadReleaseAsset(owner, repo, checksumAsset)
	if err != nil {
		return fmt.Errorf("failed to download checksum file: %w", err)
	}

	checksums := parseChecksums(string(checksumData))
	expectedChecksum := checksums[assetName]
	if expectedChecksum == "" {
		return fmt.Errorf("checksum not found for %s", assetName)
	}

	actualChecksum := calculateSHA256(data)
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func parseChecksums(checksumData string) map[string]string {
	checksums := make(map[string]string)
	lines := strings.Split(checksumData, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[1]
			// Handle paths like "plugins/plugin.so"
			if idx := strings.LastIndex(filename, "/"); idx >= 0 {
				filename = filename[idx+1:]
			}
			checksums[filename] = checksum
		}
	}
	return checksums
}

func calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// detectRepoFromGit tries to detect the GitHub repository from git remote
func detectRepoFromGit() string {
	// Try to get remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	remoteURL := strings.TrimSpace(string(output))
	
	// Parse GitHub URLs
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	if strings.Contains(remoteURL, "github.com") {
		var owner, repo string
		
		if strings.HasPrefix(remoteURL, "https://") {
			// https://github.com/owner/repo.git
			parts := strings.Split(remoteURL, "/")
			if len(parts) >= 3 {
				owner = parts[len(parts)-2]
				repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
				return fmt.Sprintf("%s/%s", owner, repo)
			}
		} else if strings.Contains(remoteURL, "@") {
			// git@github.com:owner/repo.git
			parts := strings.Split(remoteURL, ":")
			if len(parts) >= 2 {
				repoPart := strings.TrimSuffix(parts[1], ".git")
				repoParts := strings.Split(repoPart, "/")
				if len(repoParts) >= 2 {
					owner = repoParts[0]
					repo = repoParts[1]
					return fmt.Sprintf("%s/%s", owner, repo)
				}
			}
		}
	}

	return ""
}

