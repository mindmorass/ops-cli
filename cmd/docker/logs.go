package docker

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <container>",
		Short: "Show container logs",
		Args:  cobra.MinimumNArgs(1),
		Long: `Show container logs.

Examples:
  ops-cli docker logs my-container
  ops-cli docker logs my-container --tail 50`,
		RunE: runLogs,
	}

	cmd.Flags().Int("tail", 100, "Number of lines to show from the end of logs")
	cmd.Flags().Bool("follow", false, "Follow log output")

	return cmd
}

func runLogs(cmd *cobra.Command, args []string) error {
	containerName := args[0]

	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if !client.IsAvailable() {
		return fmt.Errorf("Docker is not available. Make sure Docker is running")
	}

	tail, _ := cmd.Flags().GetInt("tail")
	follow, _ := cmd.Flags().GetBool("follow")

	// Get container ID from name
	containers, err := client.ListContainers(true)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var containerID string
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.Contains(name, containerName) || c.ID[:12] == containerName {
				containerID = c.ID
				break
			}
		}
		if containerID != "" {
			break
		}
	}

	if containerID == "" {
		return fmt.Errorf("container not found: %s", containerName)
	}

	// Get logs using Docker API
	dockerClient := client.GetCLI()
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(tail),
		Follow:     follow,
	}

	reader, err := dockerClient.ContainerLogs(client.ctx, containerID, options)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	fmt.Printf("\nLogs for container: %s\n", containerName)
	fmt.Println(strings.Repeat("=", 50))

	// Read and print logs
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		// Docker logs are prefixed with 8 bytes (stream type + timestamp)
		// Skip the prefix if present
		if len(line) > 8 {
			fmt.Println(string(line[8:]))
		} else {
			fmt.Println(string(line))
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}
