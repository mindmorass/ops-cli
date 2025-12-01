package docker

import (
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/ops-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List images",
		Long: `List Docker images.

Examples:
  ops-cli docker images
  ops-cli docker images --dangling
  ops-cli docker images --format json`,
		RunE: runImages,
	}

	cmd.Flags().Bool("dangling", false, "Show only dangling images")
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}

func runImages(cmd *cobra.Command, args []string) error {
	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if !client.IsAvailable() {
		return fmt.Errorf("Docker is not available. Make sure Docker is running")
	}

	showDangling, _ := cmd.Flags().GetBool("dangling")
	format, _ := cmd.Flags().GetString("format")

	// Use spinner while fetching images
	stopSpinner := ui.StartSpinner("Fetching images...")
	images, err := client.ListImages(true) // List all images
	stopSpinner()
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	// Filter dangling images if requested
	if showDangling {
		filtered := []types.ImageSummary{}
		for _, img := range images {
			if len(img.RepoTags) == 0 || (len(img.RepoTags) == 1 && img.RepoTags[0] == "<none>:<none>") {
				filtered = append(filtered, img)
			}
		}
		images = filtered
	}

	if len(images) == 0 {
		if showDangling {
			fmt.Println("No dangling images found.")
		} else {
			fmt.Println("No images found.")
		}
		return nil
	}

	if format == "json" {
		data, err := json.MarshalIndent(images, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Table format
	fmt.Println("\nImages:")
	fmt.Println("==========")

	for _, image := range images {
		isDangling := len(image.RepoTags) == 0 || (len(image.RepoTags) == 1 && image.RepoTags[0] == "<none>:<none>")
		prefix := ""
		if isDangling {
			prefix = "[dangling] "
		}

		repoTag := "<none>:<none>"
		if len(image.RepoTags) > 0 && image.RepoTags[0] != "<none>:<none>" {
			repoTag = image.RepoTags[0]
		}

		fmt.Printf("%s%s\n", prefix, repoTag)
		fmt.Printf("   ID: %s\n", image.ID)
		fmt.Printf("   Size: %s\n", formatSize(image.Size))
		fmt.Printf("   Containers: %d\n", image.Containers)
		fmt.Println()
	}

	return nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
