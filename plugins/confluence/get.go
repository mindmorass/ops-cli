package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <page-id>",
		Short: "Get a Confluence page or content",
		Long: `Get a Confluence page by ID and display its information.

Examples:
  ops-cli confluence get 12345
  ops-cli confluence get 12345 --format json
  ops-cli confluence get 12345 --raw`,
		Args: cobra.ExactArgs(1),
		RunE: runGet,
	}

	cmd.Flags().String("format", "summary", "Output format: summary, table, json")
	cmd.Flags().Bool("raw", false, "Show raw storage format content")
	cmd.Flags().Int("length", 1000, "Content preview length")
	cmd.Flags().Bool("verbose", false, "Show additional information")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	pageID := args[0]

	client, err := getConfluenceClient()
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	raw, _ := cmd.Flags().GetBool("raw")
	length, _ := cmd.Flags().GetInt("length")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Fetch page
	stopSpinner := ui.StartSpinner("Fetching page...")
	page, err := client.GetPage(pageID, []string{
		"body.storage",
		"body.view",
		"version",
		"space",
		"ancestors",
		"children.page",
		"descendants.comment",
	})
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}

	if format == "json" {
		output, err := json.MarshalIndent(page, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Display page information
	fmt.Printf("\n📄 %s\n\n", page.Title)

	if format == "table" || format == "summary" {
		fmt.Println("Property          | Value")
		fmt.Println("------------------|----------------------------------------")
		fmt.Printf("ID                | %s\n", page.ID)
		fmt.Printf("Title             | %s\n", page.Title)
		fmt.Printf("Type              | %s\n", page.Type)
		fmt.Printf("Status            | %s\n", page.Status)

		if page.Space != nil {
			fmt.Printf("Space             | %s\n", page.Space.Name)
			fmt.Printf("Space Key         | %s\n", page.Space.Key)
		}

		if page.Version != nil {
			fmt.Printf("Version           | %d\n", page.Version.Number)
			if page.Version.When != "" {
				fmt.Printf("Created           | %s\n", formatDate(page.Version.When))
			}
			if page.Version.By != nil {
				fmt.Printf("Author            | %s\n", page.Version.By.DisplayName)
			}
		}
		fmt.Println()
	}

	// Show content preview unless format is 'table'
	if format != "table" {
		fmt.Println("📝 Content Preview:")
		fmt.Println(strings.Repeat("-", 50))

		var content string
		if raw {
			content, err = client.GetPageContent(pageID, "storage")
		} else {
			content, err = client.GetPageContent(pageID, "view")
		}
		if err != nil {
			fmt.Printf("Error getting content: %v\n", err)
		} else {
			preview := formatContent(content, length)
			fmt.Println(preview)
		}
		fmt.Println()
	}

	// Show additional details if verbose
	if verbose {
		fmt.Println("🔗 Additional Information:")

		// Show ancestors (breadcrumb)
		if len(page.Ancestors) > 0 {
			fmt.Println("\nAncestors (breadcrumb):")
			for i, ancestor := range page.Ancestors {
				indent := strings.Repeat("  ", i)
				fmt.Printf("%s• %s\n", indent, ancestor.Title)
			}
		}

		// Show child pages
		children, err := client.GetChildPages(pageID, []string{"space"}, 5)
		if err == nil && len(children.Results) > 0 {
			fmt.Println("\nChild Pages:")
			for _, child := range children.Results {
				fmt.Printf("  • %s (%s)\n", child.Title, child.ID)
			}
			if len(children.Results) == 5 {
				fmt.Println("  ... (showing first 5)")
			}
		}

		// Show comments
		comments, err := client.GetComments(pageID, []string{"body.view"}, 3)
		if err == nil && len(comments.Results) > 0 {
			fmt.Println("\nRecent Comments:")
			for _, comment := range comments.Results {
				title := comment.Title
				if title == "" {
					title = "Untitled comment"
				}
				fmt.Printf("  • %s\n", title)
			}
			if len(comments.Results) == 3 {
				fmt.Println("  ... (showing latest 3)")
			}
		}

		// Show attachments
		attachments, err := client.GetAttachments(pageID, []string{"metadata"}, 5)
		if err == nil && len(attachments.Results) > 0 {
			fmt.Println("\nAttachments:")
			for _, attachment := range attachments.Results {
				size := ""
				if attachment.Metadata != nil && attachment.Metadata.FileSize > 0 {
					size = fmt.Sprintf(" (%s)", formatFileSize(attachment.Metadata.FileSize))
				}
				fmt.Printf("  • %s%s\n", attachment.Title, size)
			}
			if len(attachments.Results) == 5 {
				fmt.Println("  ... (showing first 5)")
			}
		}
	}

	// Show page URL
	pageURL := client.BuildPageURL(pageID)
	fmt.Printf("\n🌐 View online: %s\n", pageURL)

	return nil
}

func formatContent(content string, maxLength int) string {
	if content == "" {
		return "No content"
	}

	// Strip HTML tags for basic display
	textContent := html.UnescapeString(content)
	textContent = strings.ReplaceAll(textContent, "<", "&lt;")
	textContent = strings.ReplaceAll(textContent, ">", "&gt;")

	// Remove HTML tags (simple approach)
	textContent = removeHTMLTags(textContent)

	// Clean up whitespace
	textContent = strings.ReplaceAll(textContent, "\n", " ")
	textContent = strings.ReplaceAll(textContent, "\t", " ")
	for strings.Contains(textContent, "  ") {
		textContent = strings.ReplaceAll(textContent, "  ", " ")
	}
	textContent = strings.TrimSpace(textContent)

	if len(textContent) <= maxLength {
		return textContent
	}

	return textContent[:maxLength] + "..."
}

func removeHTMLTags(s string) string {
	// Simple HTML tag removal
	var result strings.Builder
	inTag := false
	for _, char := range s {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}
	return result.String()
}

func formatDate(dateString string) string {
	// Simple date formatting - could be enhanced
	return dateString
}

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB"}

	i := 0
	size := float64(bytes)
	for size >= unit && i < len(sizes)-1 {
		size /= unit
		i++
	}

	return fmt.Sprintf("%.1f %s", size, sizes[i])
}
