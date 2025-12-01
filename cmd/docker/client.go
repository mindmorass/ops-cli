package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// DockerClient wraps the Docker API client
type DockerClient struct {
	cli *client.Client
	ctx context.Context
}

// GetCLI returns the underlying Docker client (for advanced operations)
func (d *DockerClient) GetCLI() *client.Client {
	return d.cli
}

// NewDockerClient creates a new Docker client
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerClient{
		cli: cli,
		ctx: context.Background(),
	}, nil
}

// IsAvailable checks if Docker is available and running
func (d *DockerClient) IsAvailable() bool {
	_, err := d.cli.Ping(d.ctx)
	return err == nil
}

// GetVersion returns Docker version information
func (d *DockerClient) GetVersion() (*types.Version, error) {
	version, err := d.cli.ServerVersion(d.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker version: %w", err)
	}
	return &version, nil
}

// GetSystemInfo returns Docker system information
func (d *DockerClient) GetSystemInfo() (*types.Info, error) {
	info, err := d.cli.Info(d.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker system info: %w", err)
	}
	return &info, nil
}

// ListContainers lists containers
func (d *DockerClient) ListContainers(all bool) ([]types.Container, error) {
	containers, err := d.cli.ContainerList(d.ctx, container.ListOptions{
		All: all,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	return containers, nil
}

// ListImages lists images
func (d *DockerClient) ListImages(all bool) ([]types.ImageSummary, error) {
	images, err := d.cli.ImageList(d.ctx, types.ImageListOptions{
		All: all,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	return images, nil
}

// GetContainerStats returns container statistics
func (d *DockerClient) GetContainerStats(containerID string) (*types.StatsJSON, error) {
	stats, err := d.cli.ContainerStats(d.ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var statsJSON types.StatsJSON
	decoder := json.NewDecoder(stats.Body)
	if err := decoder.Decode(&statsJSON); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}
	return &statsJSON, nil
}

// GetContainerLogsReader returns a reader for container logs
func (d *DockerClient) GetContainerLogsReader(containerID string, tail string) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	reader, err := d.cli.ContainerLogs(d.ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	return reader, nil
}

// GetContainer returns container details
func (d *DockerClient) GetContainer(containerID string) (*types.ContainerJSON, error) {
	container, err := d.cli.ContainerInspect(d.ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	return &container, nil
}

// GetRunningContainers returns only running containers
func (d *DockerClient) GetRunningContainers() ([]types.Container, error) {
	containers, err := d.ListContainers(false)
	if err != nil {
		return nil, err
	}

	// Filter to only running containers
	running := []types.Container{}
	for _, c := range containers {
		if c.State == "running" {
			running = append(running, c)
		}
	}
	return running, nil
}

// GetStoppedContainers returns only stopped containers
func (d *DockerClient) GetStoppedContainers() ([]types.Container, error) {
	containers, err := d.ListContainers(true)
	if err != nil {
		return nil, err
	}

	// Filter to only stopped containers
	stopped := []types.Container{}
	for _, c := range containers {
		if c.State != "running" {
			stopped = append(stopped, c)
		}
	}
	return stopped, nil
}

// StartContainer starts a container by ID or name
func (d *DockerClient) StartContainer(containerID string) error {
	return d.cli.ContainerStart(d.ctx, containerID, container.StartOptions{})
}

// StopContainer stops a container by ID or name
func (d *DockerClient) StopContainer(containerID string, timeout *int) error {
	opts := container.StopOptions{}
	if timeout != nil {
		opts.Timeout = timeout
	}
	return d.cli.ContainerStop(d.ctx, containerID, opts)
}

// FindContainerByNameOrID finds a container by name or ID
func (d *DockerClient) FindContainerByNameOrID(nameOrID string) (*types.Container, error) {
	containers, err := d.ListContainers(true)
	if err != nil {
		return nil, err
	}

	// Try exact ID match first
	for _, c := range containers {
		if c.ID == nameOrID || c.ID[:12] == nameOrID {
			return &c, nil
		}
	}

	// Try name match
	for _, c := range containers {
		for _, n := range c.Names {
			// Remove leading slash from name
			cleanName := n
			if len(n) > 0 && n[0] == '/' {
				cleanName = n[1:]
			}
			if cleanName == nameOrID {
				return &c, nil
			}
		}
	}

	return nil, fmt.Errorf("container not found: %s", nameOrID)
}

// Close closes the Docker client
func (d *DockerClient) Close() error {
	return d.cli.Close()
}

// GetDockerSocketPath returns the Docker socket path for the current platform
func GetDockerSocketPath() string {
	if runtime.GOOS == "windows" {
		return "//./pipe/docker_engine"
	}
	return "/var/run/docker.sock"
}
