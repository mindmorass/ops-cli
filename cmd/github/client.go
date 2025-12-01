package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ops-cli/internal/api/github"
	"github.com/ops-cli/internal/config"
)

// getGitHubClient creates a GitHub client from config or environment
// Token is optional - public repos work without it, private repos require it
func getGitHubClient() (*github.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get token from config or environment
	token := ""
	if cfg.GitHub != nil && cfg.GitHub.Token != "" {
		token = cfg.GitHub.Token
	}

	// Environment variable takes precedence
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		token = envToken
	}

	// Token is optional - allow client without token for public repos
	return github.NewClient(token), nil
}

// hasToken checks if a GitHub token is configured
func hasToken() bool {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}

	if cfg.GitHub != nil && cfg.GitHub.Token != "" {
		return true
	}

	return os.Getenv("GITHUB_TOKEN") != ""
}

// handleGitHubError provides user-friendly error messages for GitHub API errors
func handleGitHubError(err error, owner, repo, operation string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	
	// Check for 404 errors
	if strings.Contains(errStr, "404") {
		if !hasToken() {
			return fmt.Errorf(`%s failed: repository "%s/%s" not found or is private

This could mean:
  - The repository doesn't exist
  - The repository is private and requires authentication

To access private repositories, configure a GitHub token:
  ops-cli github config setup

Or set the GITHUB_TOKEN environment variable`,
				operation, owner, repo)
		}
		return fmt.Errorf(`%s failed: repository "%s/%s" not found or you don't have access

This could mean:
  - The repository doesn't exist
  - The repository is private and your token doesn't have access
  - The repository name is incorrect

Verify the repository exists: https://github.com/%s/%s`,
			operation, owner, repo, owner, repo)
	}

	// Check for 401 errors (unauthorized)
	if strings.Contains(errStr, "401") {
		return fmt.Errorf(`%s failed: authentication required

Your GitHub token may be invalid or expired. Please reconfigure:
  ops-cli github config setup`,
			operation)
	}

	// Check for 403 errors (forbidden)
	if strings.Contains(errStr, "403") {
		return fmt.Errorf(`%s failed: access forbidden

This could mean:
  - Your token doesn't have the required permissions
  - You've exceeded the rate limit
  - The repository requires different permissions

Check your token permissions and rate limit status`,
			operation)
	}

	// Return original error for other cases
	return err
}

// detectRepoFromGit tries to detect the GitHub repository from the current git directory
func detectRepoFromGit() (string, error) {
	// Check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not a git repository")
	}

	// Try to get remote URL
	cmd = exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
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
				return fmt.Sprintf("%s/%s", owner, repo), nil
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
					return fmt.Sprintf("%s/%s", owner, repo), nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not parse GitHub repository from remote URL: %s", remoteURL)
}

// getRepoArg gets the repository argument, using git detection if not provided
func getRepoArg(args []string, argIndex int) (string, error) {
	if len(args) > argIndex {
		return args[argIndex], nil
	}

	// Try to detect from git
	repo, err := detectRepoFromGit()
	if err != nil {
		return "", fmt.Errorf("repository not provided and could not detect from git: %w", err)
	}

	return repo, nil
}
