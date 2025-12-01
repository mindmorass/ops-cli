package docker

import (
	"github.com/docker/docker/api/types"
)

// getContainerDisplayName returns a user-friendly container name
func getContainerDisplayName(container *types.Container) string {
	if len(container.Names) > 0 {
		name := container.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		return name
	}
	return container.ID[:12]
}
