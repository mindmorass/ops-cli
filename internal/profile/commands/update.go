package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ops-cli/internal/profile/ui"
)

type UpdateOptions struct {
	ProfileName string
	Force       bool
	DryRun      bool
	NoBackup    bool
}

// UpdateProfile updates an existing profile with new features
func UpdateProfile(profilesDir string, opts UpdateOptions) error {
	// If no profile name provided, show interactive selection
	if opts.ProfileName == "" {
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			return fmt.Errorf("failed to read profiles directory: %w", err)
		}

		var profiles []string
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != ".git" {
				profilePath := filepath.Join(profilesDir, entry.Name())
				envrcPath := filepath.Join(profilePath, ".envrc")
				if _, err := os.Stat(envrcPath); err == nil {
					profiles = append(profiles, entry.Name())
				}
			}
		}

		if len(profiles) == 0 {
			return fmt.Errorf("no profiles found")
		}

		selected, err := ui.SelectProfile(profiles, "Select profile to update:")
		if err != nil {
			return err
		}
		opts.ProfileName = selected
	}

	profileDir := filepath.Join(profilesDir, opts.ProfileName)

	// Check if profile exists
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist at: %s", opts.ProfileName, profileDir)
	}

	envrcPath := filepath.Join(profileDir, ".envrc")
	if _, err := os.Stat(envrcPath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not appear to be a valid profile (missing .envrc)", opts.ProfileName)
	}

	ui.PrintInfo(fmt.Sprintf("Updating profile: %s", opts.ProfileName))
	fmt.Printf("  Location: %s\n", profileDir)
	fmt.Println()

	// Create backup unless --no-backup is specified
	if !opts.NoBackup && !opts.DryRun {
		if err := createBackup(profileDir, opts.ProfileName); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to create backup: %v", err))
			if !opts.Force {
				confirmed, err := ui.Confirm("Continue without backup?", false)
				if err != nil || !confirmed {
					return fmt.Errorf("update cancelled")
				}
			}
		}
	}

	// Track what was updated
	updates := []string{}

	// Update directories
	if updated, err := updateDirectories(profileDir, opts.DryRun); err != nil {
		return fmt.Errorf("failed to update directories: %w", err)
	} else if len(updated) > 0 {
		updates = append(updates, fmt.Sprintf("Created directories: %s", strings.Join(updated, ", ")))
	}

	// Update .envrc
	if updated, err := updateEnvrc(profileDir, opts.ProfileName, opts.DryRun, opts.Force); err != nil {
		return fmt.Errorf("failed to update .envrc: %w", err)
	} else if updated {
		updates = append(updates, "Updated .envrc with new environment variables")
	}

	// Update .gitignore
	if updated, err := updateGitignore(profileDir, opts.DryRun, opts.Force); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	} else if updated {
		updates = append(updates, "Updated .gitignore with new patterns")
	}

	// Summary
	if opts.DryRun {
		ui.PrintInfo("DRY RUN - No changes were made")
		if len(updates) > 0 {
			fmt.Println()
			fmt.Println("Would update:")
			for _, update := range updates {
				fmt.Printf("  - %s\n", update)
			}
		} else {
			fmt.Println("  Profile is already up to date")
		}
	} else {
		if len(updates) > 0 {
			ui.PrintSuccess("Profile updated successfully")
			fmt.Println()
			fmt.Println("Updates applied:")
			for _, update := range updates {
				fmt.Printf("  ✓ %s\n", update)
			}
		} else {
			ui.PrintInfo("Profile is already up to date")
		}
	}

	return nil
}

func createBackup(profileDir, _profileName string) error {
	backupDir := filepath.Join(profileDir, ".backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("update_%s", timestamp))

	// Copy important files
	filesToBackup := []string{
		".envrc",
		".gitconfig",
		".gitignore",
	}

	for _, file := range filesToBackup {
		src := filepath.Join(profileDir, file)
		if _, err := os.Stat(src); err == nil {
			content, err := os.ReadFile(src)
			if err != nil {
				continue
			}

			backupFile := filepath.Join(backupPath, file)
			if err := os.MkdirAll(filepath.Dir(backupFile), 0755); err != nil {
				continue
			}

			if err := os.WriteFile(backupFile, content, 0644); err != nil {
				continue
			}
		}
	}

	ui.PrintInfo(fmt.Sprintf("Backup created: %s", backupPath))
	return nil
}

func updateDirectories(profileDir string, dryRun bool) ([]string, error) {
	requiredDirs := []string{
		".config/1Password",
		".config/claude",
		".config/gemini",
		".ssh",
		".aws",
		".azure",
		".gcloud",
		".kube",
		"bin",
		"code",
	}

	var created []string
	for _, dir := range requiredDirs {
		fullPath := filepath.Join(profileDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			if !dryRun {
				if err := os.MkdirAll(fullPath, 0755); err != nil {
					return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
			}
			created = append(created, dir)
		}
	}

	// Set SSH directory permissions
	sshDir := filepath.Join(profileDir, ".ssh")
	if _, err := os.Stat(sshDir); err == nil && !dryRun {
		if err := os.Chmod(sshDir, 0700); err != nil {
			// Non-fatal, just warn
			ui.PrintWarning(fmt.Sprintf("Failed to set SSH directory permissions: %v", err))
		}
	}

	return created, nil
}

func updateEnvrc(profileDir, _profileName string, dryRun, _force bool) (bool, error) {
	envrcPath := filepath.Join(profileDir, ".envrc")
	content, err := os.ReadFile(envrcPath)
	if err != nil {
		return false, fmt.Errorf("failed to read .envrc: %w", err)
	}

	envrcContent := string(content)
	updated := false

	// Define sections with their variables
	sections := []struct {
		comment string
		vars    []struct {
			name string
			line string
		}
	}{
		{
			comment: "# XDG Base Directory specification\n# Point all XDG-compliant tools to workspace-specific config\n",
			vars: []struct {
				name string
				line string
			}{
				{"XDG_CONFIG_HOME", `export XDG_CONFIG_HOME="$WORKSPACE_HOME/.config"`},
			},
		},
		{
			comment: "# Git configuration\n",
			vars: []struct {
				name string
				line string
			}{
				{"GIT_CONFIG_GLOBAL", `export GIT_CONFIG_GLOBAL="$WORKSPACE_HOME/.gitconfig"`},
			},
		},
		{
			comment: "# AWS configuration\n# Point AWS CLI and SDKs to workspace-specific config and credentials\n",
			vars: []struct {
				name string
				line string
			}{
				{"AWS_CONFIG_FILE", `export AWS_CONFIG_FILE="$WORKSPACE_HOME/.aws/config"`},
				{"AWS_SHARED_CREDENTIALS_FILE", `export AWS_SHARED_CREDENTIALS_FILE="$WORKSPACE_HOME/.aws/credentials"`},
			},
		},
		{
			comment: "# Kubernetes configuration\n# Point kubectl to workspace-specific kubeconfig\n",
			vars: []struct {
				name string
				line string
			}{
				{"KUBECONFIG", `export KUBECONFIG="$WORKSPACE_HOME/.kube/config"`},
			},
		},
		{
			comment: "# Terraform configuration\n# Use workspace-specific Terraform CLI config\n",
			vars: []struct {
				name string
				line string
			}{
				{"TF_CLI_CONFIG_FILE", `export TF_CLI_CONFIG_FILE="$WORKSPACE_HOME/.terraformrc"`},
			},
		},
		{
			comment: "# Azure CLI configuration\n# Point Azure CLI to workspace-specific config directory\n",
			vars: []struct {
				name string
				line string
			}{
				{"AZURE_CONFIG_DIR", `export AZURE_CONFIG_DIR="$WORKSPACE_HOME/.azure"`},
			},
		},
		{
			comment: "# Google Cloud SDK configuration\n# Point gcloud CLI to workspace-specific config directory\n",
			vars: []struct {
				name string
				line string
			}{
				{"CLOUDSDK_CONFIG", `export CLOUDSDK_CONFIG="$WORKSPACE_HOME/.gcloud"`},
			},
		},
		{
			comment: "# Claude Code configuration\n# Point Claude Code to workspace-specific config directory\n",
			vars: []struct {
				name string
				line string
			}{
				{"CLAUDE_CONFIG_DIR", `export CLAUDE_CONFIG_DIR="$WORKSPACE_HOME/.config/claude"`},
			},
		},
		{
			comment: "# Gemini CLI configuration\n# Point Gemini CLI to workspace-specific config directory\n",
			vars: []struct {
				name string
				line string
			}{
				{"GEMINI_CONFIG_DIR", `export GEMINI_CONFIG_DIR="$WORKSPACE_HOME/.config/gemini"`},
			},
		},
	}

	// Find insertion point (before "# Load .env file")
	insertPoint := strings.Index(envrcContent, "# Load .env file if it exists")
	if insertPoint == -1 {
		insertPoint = strings.Index(envrcContent, "dotenv_if_exists .env")
		if insertPoint == -1 {
			// Append at end before welcome message
			insertPoint = strings.LastIndex(envrcContent, "# Welcome message")
			if insertPoint == -1 {
				insertPoint = len(envrcContent)
			}
		}
	}

	before := envrcContent[:insertPoint]
	after := envrcContent[insertPoint:]

	// Process each section
	for _, section := range sections {
		// Check which variables in this section are missing
		var missingVars []string
		for _, v := range section.vars {
			if !strings.Contains(envrcContent, v.name) {
				missingVars = append(missingVars, v.line)
			}
		}

		if len(missingVars) > 0 {
			// Check if section comment already exists
			sectionExists := strings.Contains(before, section.comment)

			var newContent string
			if !sectionExists {
				// Add section comment and all missing variables
				newContent = section.comment
				for _, varLine := range missingVars {
					newContent += varLine + "\n"
				}
				newContent += "\n"
			} else {
				// Section exists, find where to insert variables
				// Insert after the section comment
				sectionStart := strings.Index(before, section.comment)
				if sectionStart != -1 {
					sectionEnd := sectionStart + len(section.comment)
					// Find next section or end
					nextSection := strings.Index(before[sectionEnd:], "\n# ")
					if nextSection == -1 {
						nextSection = len(before) - sectionEnd
					}
					// Insert variables before next section
					insertPos := sectionEnd + nextSection
					before = before[:insertPos] + strings.Join(missingVars, "\n") + "\n" + before[insertPos:]
					updated = true
					continue
				}
				// Fallback: just add variables
				newContent = strings.Join(missingVars, "\n") + "\n\n"
			}

			before += newContent
			updated = true
		}
	}

	if updated && !dryRun {
		envrcContent = before + after
		if err := os.WriteFile(envrcPath, []byte(envrcContent), 0644); err != nil {
			return false, fmt.Errorf("failed to write .envrc: %w", err)
		}
	}

	return updated, nil
}

func updateGitignore(profileDir string, dryRun, _force bool) (bool, error) {
	gitignorePath := filepath.Join(profileDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		// .gitignore doesn't exist, create it using the same function from create.go
		// We'll create a basic one inline
		if !dryRun {
			gitignoreContent := `# Workspace profile gitignore

# Environment files with secrets
.env
.envrc.local

# SSH keys and sensitive files
.ssh/id_*
.ssh/*.pem
.ssh/*.key
.ssh/known_hosts

# AWS credentials and sensitive config
.aws/credentials
.aws/cli/cache
.aws/sso/cache

# Azure CLI credentials and sensitive config
.azure/config
.azure/clouds.config
.azure/accessTokens.json
.azure/msal_token_cache.json
.azure/azureProfile.json

# Google Cloud SDK credentials and sensitive config
.gcloud/configurations/
.gcloud/credentials
.gcloud/access_tokens.db
.gcloud/legacy_credentials/
.gcloud/logs/

# Claude Code configuration (may contain API keys and sensitive data)
.config/claude/

# Gemini CLI configuration (may contain API keys and sensitive data)
.config/gemini/

# Terraform
.terraform/
.terraform.lock.hcl
*.tfstate
*.tfstate.*
*.tfvars
.terraform.d/plugin-cache/
.terraform.d/checkpoint_cache
.terraform.d/checkpoint_signature

# Terragrunt
.terragrunt-cache/
*.tfplan

# Kubernetes
.kube/cache
.kube/http-cache

# OS files
.DS_Store
Thumbs.db

# Editor files
.vscode/
.idea/
*.swp
*.swo
*~

# Build artifacts
bin/
dist/
build/
*.log
`
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
				return false, fmt.Errorf("failed to create .gitignore: %w", err)
			}
		}
		return true, nil
	}

	gitignoreContent := string(content)
	updated := false

	// Check and add missing patterns
	requiredPatterns := map[string]string{
		".azure/config":              "# Azure CLI credentials and sensitive config",
		".gcloud/configurations":     "# Google Cloud SDK credentials and sensitive config",
		".gcloud/credentials":        "",
		".gcloud/access_tokens.db":   "",
		".gcloud/legacy_credentials": "",
		".gcloud/logs":               "",
		".config/claude/":            "# Claude Code configuration (may contain API keys and sensitive data)",
		".config/gemini/":            "# Gemini CLI configuration (may contain API keys and sensitive data)",
	}

	// Group patterns by comment
	patternsByComment := make(map[string][]string)
	currentComment := ""
	for pattern, comment := range requiredPatterns {
		if comment != "" {
			currentComment = comment
		}
		if patternsByComment[currentComment] == nil {
			patternsByComment[currentComment] = []string{}
		}
		patternsByComment[currentComment] = append(patternsByComment[currentComment], pattern)
	}

	for comment, patterns := range patternsByComment {
		// Check if any pattern from this group is missing
		hasAny := false
		for _, pattern := range patterns {
			if strings.Contains(gitignoreContent, pattern) {
				hasAny = true
				break
			}
		}

		if !hasAny {
			// Find insertion point (after Azure section or at end)
			insertPoint := strings.Index(gitignoreContent, "# Azure CLI credentials")
			if insertPoint == -1 {
				insertPoint = strings.Index(gitignoreContent, "# Terraform")
				if insertPoint == -1 {
					insertPoint = len(gitignoreContent)
				}
			} else {
				// Find end of Azure section
				insertPoint = strings.Index(gitignoreContent[insertPoint:], "\n\n#")
				if insertPoint != -1 {
					insertPoint += insertPoint
				} else {
					insertPoint = strings.Index(gitignoreContent, "# Terraform")
					if insertPoint == -1 {
						insertPoint = len(gitignoreContent)
					}
				}
			}

			before := gitignoreContent[:insertPoint]
			after := gitignoreContent[insertPoint:]

			newSection := ""
			if comment != "" {
				newSection = comment + "\n"
			}
			for _, pattern := range patterns {
				newSection += pattern + "\n"
			}
			newSection += "\n"

			gitignoreContent = before + newSection + after
			updated = true
		}
	}

	if updated && !dryRun {
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return false, fmt.Errorf("failed to write .gitignore: %w", err)
		}
	}

	return updated, nil
}
