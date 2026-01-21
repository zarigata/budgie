package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zarigata/budgie/pkg/types"
)

type Bundle struct {
	Version       string                  `yaml:"version"`
	Name          string                  `yaml:"name"`
	Image         types.ImageConfig       `yaml:"image"`
	Ports         []types.PortMapping     `yaml:"ports"`
	Volumes       []types.VolumeMapping   `yaml:"volumes"`
	Env           []string                `yaml:"environment"`
	EnvFile       string                  `yaml:"env_file"`
	Health        *types.HealthCheck      `yaml:"healthcheck"`
	Replicas      *types.ReplicasConfig   `yaml:"replicas"`
	Resources     *types.ResourceLimits   `yaml:"resources"`
	RestartPolicy *types.RestartPolicy    `yaml:"restart_policy"`
	DependsOn     []string                `yaml:"depends_on"`
	StopTimeout   int                     `yaml:"stop_timeout"`
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
	// Load environment from file if specified
	env := b.Env
	if b.EnvFile != "" {
		fileEnv, err := loadEnvFile(b.EnvFile, bundlePath)
		if err == nil {
			env = append(fileEnv, env...) // Bundle env overrides file env
		}
	}

	ctr := &types.Container{
		ID:            types.GenerateContainerID(),
		Name:          b.Name,
		State:         types.StateCreating,
		Image:         b.Image,
		Ports:         b.Ports,
		Volumes:       b.Volumes,
		Env:           env,
		Health:        b.Health,
		Replicas:      b.Replicas,
		Resources:     b.Resources,
		RestartPolicy: b.RestartPolicy,
		BundlePath:    bundlePath,
		NodeID:        getNodeID(),
		CreatedAt:     time.Now(),
	}

	// Set default restart policy if not specified
	if ctr.RestartPolicy == nil {
		ctr.RestartPolicy = &types.RestartPolicy{Name: "no"}
	}

	return ctr
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(envFile, bundlePath string) ([]string, error) {
	// Resolve relative to bundle path
	if !filepath.IsAbs(envFile) {
		envFile = filepath.Join(filepath.Dir(bundlePath), envFile)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	var env []string
	lines := splitLines(string(data))
	for _, line := range lines {
		line = trimSpace(line)
		// Skip empty lines and comments
		if line == "" || line[0] == '#' {
			continue
		}
		// Simple KEY=VALUE parsing
		if idx := indexOf(line, '='); idx > 0 {
			env = append(env, line)
		}
	}

	return env, nil
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else if c != '\r' {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func indexOf(s string, c rune) int {
	for i, r := range s {
		if r == c {
			return i
		}
	}
	return -1
}

func getNodeID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}
