package startpage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func newThemeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "theme [theme-name]",
		Short: "Change the visual theme of your startpage",
		Long: `Change the visual theme of your startpage.

Examples:
  ops-cli startpage theme
  ops-cli startpage theme ocean
  ops-cli startpage theme --list`,
		RunE: runTheme,
	}

	cmd.Flags().Bool("list", false, "List available themes")

	return cmd
}

func runTheme(cmd *cobra.Command, args []string) error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no startpage found. Run 'ops-cli startpage init' to create one")
	}

	list, _ := cmd.Flags().GetBool("list")
	if list {
		fmt.Println("Available OpenLinks themes:")
		fmt.Println()
		themes := []string{
			"default - Clean default theme",
			"ocean - Blue ocean inspired theme",
			"forest - Green forest theme",
			"sunrise - Warm sunrise colors",
			"ness - Retro game inspired",
			"arctic - Cool arctic theme",
			"cherry - Pink cherry blossom theme",
			"brutalism - Bold brutalist design",
		}
		for _, theme := range themes {
			fmt.Printf("  • %s\n", theme)
		}
		fmt.Println()
		fmt.Println("Usage: ops-cli startpage theme [theme-name]")
		return nil
	}

	themeName := ""
	if len(args) > 0 {
		themeName = args[0]
	} else {
		if err := survey.AskOne(&survey.Input{
			Message: "Which theme would you like to use?",
		}, &themeName, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("theme name is required")
		}
	}

	validThemes := []string{
		"default", "ocean", "forest", "sunrise", "ness", "arctic", "cherry", "brutalism",
	}

	themeName = strings.ToLower(themeName)
	valid := false
	for _, t := range validThemes {
		if t == themeName {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid theme \"%s\". Available themes: %s", themeName, strings.Join(validThemes, ", "))
	}

	// Update OpenLinks.json with new theme
	openLinksPath := filepath.Join(config.Path, "OpenLinks.json")
	data, err := os.ReadFile(openLinksPath)
	if err != nil {
		return fmt.Errorf("failed to read OpenLinks.json: %w", err)
	}

	var openLinks OpenLinksConfig
	if err := json.Unmarshal(data, &openLinks); err != nil {
		return fmt.Errorf("failed to parse OpenLinks.json: %w", err)
	}

	openLinks.Theme = themeName

	updatedData, err := json.MarshalIndent(openLinks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenLinks.json: %w", err)
	}

	if err := os.WriteFile(openLinksPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write OpenLinks.json: %w", err)
	}

	fmt.Printf("Applied \"%s\" theme\n", themeName)

	// Rebuild the startpage
	fmt.Println("Rebuilding startpage...")
	if err := rebuildStartpage(config); err != nil {
		fmt.Println("Theme applied but rebuild failed. Run 'ops-cli startpage build' manually.")
		return nil
	}

	fmt.Println("Startpage rebuilt with new theme!")
	fmt.Println("Visit http://localhost:3000 to see the changes")

	return nil
}
