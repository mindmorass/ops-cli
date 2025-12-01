package startpage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new Astro startpage project",
		Long: `Initialize a new Astro startpage project with OpenLinks theme.

Examples:
  ops-cli startpage init
  ops-cli startpage init ~/my-startpage
  ops-cli startpage init --name "My Browser Home"`,
		RunE: runInit,
	}

	cmd.Flags().String("name", "", "Project name")
	cmd.Flags().String("author-name", "", "Author name")
	cmd.Flags().Bool("yes", false, "Skip confirmation prompts")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	if err := ensureStartpageDirectory(); err != nil {
		return fmt.Errorf("failed to create startpage directory: %w", err)
	}

	// Get project path
	projectPath := ""
	if len(args) > 0 {
		var err error
		projectPath, err = expandTilde(args[0])
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
	} else {
		dir, err := getStartpageDirectory()
		if err != nil {
			return err
		}
		projectPath = filepath.Join(dir, "project")
	}

	// Get flags
	name, _ := cmd.Flags().GetString("name")
	authorName, _ := cmd.Flags().GetString("author-name")
	yes, _ := cmd.Flags().GetBool("yes")

	// Interactive prompts if not provided
	if name == "" && !yes {
		if err := survey.AskOne(&survey.Input{
			Message: "What should we call your startpage?",
			Default: "My Startpage",
		}, &name); err != nil {
			return fmt.Errorf("name is required")
		}
	}
	if name == "" {
		name = "My Startpage"
	}

	if authorName == "" && !yes {
		if err := survey.AskOne(&survey.Input{
			Message: "Your name:",
			Default: "Anonymous",
		}, &authorName); err != nil {
			return fmt.Errorf("author name is required")
		}
	}
	if authorName == "" {
		authorName = "Anonymous"
	}

	if projectPath == "" && !yes {
		if err := survey.AskOne(&survey.Input{
			Message: "Where should we create the project?",
			Default: projectPath,
		}, &projectPath); err != nil {
			return fmt.Errorf("project path is required")
		}
		var err error
		projectPath, err = expandTilde(projectPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
	}

	fmt.Printf("Creating Astro startpage: %s\n", name)
	fmt.Printf("Location: %s\n\n", projectPath)

	// Setup OpenLinks theme
	if err := setupOpenLinksTheme(projectPath, name, authorName); err != nil {
		return fmt.Errorf("failed to setup theme: %w", err)
	}

	// Create initial config
	config := &StartpageConfig{
		Name:  name,
		Path:  projectPath,
		Theme: "miduwind",
		Bookmarks: []BookmarkGroup{
			{
				ID:    "dev",
				Name:  "Development",
				Icon:  "🛠️",
				Color: "blue",
				Bookmarks: []Bookmark{
					{
						ID:          "github",
						Name:        "GitHub",
						URL:         "https://github.com",
						Icon:        "🐙",
						Description: "Code repositories",
					},
					{
						ID:          "vscode",
						Name:        "VS Code",
						URL:         "vscode://",
						Icon:        "💻",
						Description: "Code editor",
					},
				},
			},
			{
				ID:    "productivity",
				Name:  "Productivity",
				Icon:  "⚡",
				Color: "green",
				Bookmarks: []Bookmark{
					{
						ID:          "gmail",
						Name:        "Gmail",
						URL:         "https://gmail.com",
						Icon:        "📧",
						Description: "Email",
					},
				},
			},
		},
		Settings: Settings{
			Title:       name,
			Description: "My personalized browser startpage",
			Author: Author{
				Name: authorName,
				Bio:  fmt.Sprintf("Welcome to %s's startpage!", authorName),
			},
		},
	}

	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Startpage created successfully!")
	fmt.Println("Building startpage for immediate use...\n")

	// Generate the OpenLinks.json file
	if err := generateStartpage(config); err != nil {
		return fmt.Errorf("failed to generate startpage: %w", err)
	}

	// Build the project
	if err := buildProject(projectPath); err != nil {
		fmt.Printf("\nBuild failed. You can build manually later with:\n")
		fmt.Printf("   ops-cli startpage build\n")
		fmt.Printf("\nProject created at: %s\n", projectPath)
		return nil
	}

	fmt.Println("\nBuild completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("   1. ops-cli startpage serve   # Start web server")
	fmt.Println("   2. Set browser homepage to: http://localhost:3000")
	fmt.Println("   3. ops-cli startpage add     # Add more bookmarks")
	fmt.Println("   4. ops-cli startpage build  # Rebuild after changes")
	fmt.Println("\nNote: Modern themes require a web server, not file:// URLs")

	return nil
}

// setupOpenLinksTheme clones and sets up the OpenLinks theme
func setupOpenLinksTheme(projectPath, name, authorName string) error {
	fmt.Println("Setting up OpenLinks theme...")

	// Ensure parent directory exists
	parentDir := filepath.Dir(projectPath)
	if parentDir != "" && parentDir != "." {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
	}

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory already exists: %s", projectPath)
	}

	// Clone the OpenLinks repository
	fmt.Println("Cloning OpenLinks repository...")
	cloneCmd := exec.Command("git", "clone", "https://github.com/E10YDEV/OpenLinks.git", projectPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone OpenLinks: %w", err)
	}

	// Remove .git directory
	gitDir := filepath.Join(projectPath, ".git")
	os.RemoveAll(gitDir)

	// Install dependencies
	fmt.Println("Installing dependencies...")
	installCmd := exec.Command("npm", "install")
	installCmd.Dir = projectPath
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	return nil
}

// generateStartpage generates the OpenLinks.json file
func generateStartpage(config *StartpageConfig) error {
	// Get all bookmarks from all groups
	var allBookmarks []Bookmark
	for _, group := range config.Bookmarks {
		allBookmarks = append(allBookmarks, group.Bookmarks...)
	}

	// Load existing OpenLinks.json if it exists
	openLinksPath := filepath.Join(config.Path, "OpenLinks.json")
	existingTheme := "default"
	existingProfile := Profile{
		Name:        config.Settings.Author.Name,
		Avatar:      "avatar/Avatar.webp",
		Description: fmt.Sprintf("Welcome to %s's startpage", config.Settings.Author.Name),
		Adult:       false,
	}
	existingFooter := fmt.Sprintf("Made with ❤️ by %s", config.Settings.Author.Name)

	if data, err := os.ReadFile(openLinksPath); err == nil {
		var existing OpenLinksConfig
		if err := json.Unmarshal(data, &existing); err == nil {
			if existing.Theme != "" {
				existingTheme = existing.Theme
			}
			if existing.Profile.Name != "" {
				existingProfile = existing.Profile
			}
			if existing.Footer != "" {
				existingFooter = existing.Footer
			}
		}
	}

	// Create OpenLinks.json
	openLinksData := OpenLinksConfig{
		Title:       config.Name,
		Description: config.Settings.Description,
		URLBase:     "http://localhost:3000",
		Theme:       existingTheme,
		Footer:      existingFooter,
		Profile:     existingProfile,
		Links:       make([]OpenLink, len(allBookmarks)),
	}

	for i, bookmark := range allBookmarks {
		openLinksData.Links[i] = OpenLink{
			Name: bookmark.Name,
			URL:  bookmark.URL,
			Icon: bookmark.Icon,
		}
		if openLinksData.Links[i].Icon == "" {
			openLinksData.Links[i].Icon = "🔗"
		}
	}

	// Write OpenLinks.json
	data, err := json.MarshalIndent(openLinksData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenLinks config: %w", err)
	}

	return os.WriteFile(openLinksPath, data, 0644)
}

// buildProject builds the Astro project
func buildProject(projectPath string) error {
	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = projectPath
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	return buildCmd.Run()
}
