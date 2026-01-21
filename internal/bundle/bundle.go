package bundle

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zarigata/budgie/pkg/types"
)

type Bundle struct {
	Version   string                  `yaml:"version"`
	Name      string                  `yaml:"name"`
	Image     types.ImageConfig       `yaml:"image"`
	Ports     []types.PortMapping     `yaml:"ports"`
	Volumes   []types.VolumeMapping   `yaml:"volumes"`
	Env       []string                `yaml:"environment"`
	Health    *types.HealthCheck      `yaml:"healthcheck"`
	Replicas  *types.ReplicasConfig   `yaml:"replicas"`
	Resources *types.ResourceLimits   `yaml:"resources"`
}

func Parse(path string) (*Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle file: %w", err)
	}

	var bundle Bundle
	if err := yaml.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse bundle file: %w", err)
	}

	if bundle.Version == "" {
		return nil, fmt.Errorf("bundle version is required")
	}

	if bundle.Name == "" {
		bundle.Name = filepath.Base(path)
	}

	if len(bundle.Ports) == 0 {
		return nil, fmt.Errorf("at least one port mapping is required")
	}

	return &bundle, nil
}

func (b *Bundle) ToContainer(bundlePath string) *types.Container {
	ctr := &types.Container{
		ID:         types.GenerateContainerID(),
		Name:       b.Name,
		State:      types.StateCreating,
		Image:      b.Image,
		Ports:      b.Ports,
		Volumes:    b.Volumes,
		Env:        b.Env,
		Health:     b.Health,
		Replicas:   b.Replicas,
		Resources:  b.Resources,
		BundlePath: bundlePath,
		NodeID:     getNodeID(),
		CreatedAt:  time.Now(),
	}

	return ctr
}

func getNodeID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}
