package types

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// ContainerState represents the state of a container
type ContainerState string

const (
	StateCreating ContainerState = "creating"
	StateCreated ContainerState = "created"
	StateRunning  ContainerState = "running"
	StateStopped  ContainerState = "stopped"
	StatePaused   ContainerState = "paused"
	StateFailed   ContainerState = "failed"
)

// String returns the string representation of the state
func (s ContainerState) String() string {
	return string(s)
}

// PortMapping defines a port mapping between container and host
type PortMapping struct {
	ContainerPort int    `yaml:"container_port" json:"container_port"`
	HostPort      int    `yaml:"host_port" json:"host_port"`
	Protocol      string `yaml:"protocol" json:"protocol"` // "tcp" or "udp"
}

// VolumeMapping defines a volume mount
type VolumeMapping struct {
	Source string `yaml:"source" json:"source"`
	Target string `yaml:"target" json:"target"`
	Mode   string `yaml:"mode" json:"mode"` // "rw" or "ro"
}

// HealthCheck defines health check configuration
type HealthCheck struct {
	Path     string        `yaml:"path" json:"path"`
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Retries  int           `yaml:"retries" json:"retries"`
}

// ReplicasConfig defines replica configuration
type ReplicasConfig struct {
	Min int `yaml:"min" json:"min"`
	Max int `yaml:"max" json:"max"`
}

// ResourceLimits defines resource constraints for containers
type ResourceLimits struct {
	CPUShares   int64  `yaml:"cpu_shares" json:"cpu_shares"`     // CPU shares (relative weight)
	CPUQuota    int64  `yaml:"cpu_quota" json:"cpu_quota"`       // CPU CFS quota in microseconds
	MemoryLimit int64  `yaml:"memory_limit" json:"memory_limit"` // Memory limit in bytes
	MemorySwap  int64  `yaml:"memory_swap" json:"memory_swap"`   // Memory + Swap limit
	BlkioWeight uint16 `yaml:"blkio_weight" json:"blkio_weight"` // Block I/O weight (10-1000)
	PidsLimit   int64  `yaml:"pids_limit" json:"pids_limit"`     // Max number of PIDs
}

// RestartPolicy defines container restart behavior
type RestartPolicy struct {
	Name              string `yaml:"name" json:"name"`                               // "no", "always", "on-failure", "unless-stopped"
	MaximumRetryCount int    `yaml:"maximum_retry_count" json:"maximum_retry_count"` // Max retries for "on-failure"
}

// ImageConfig defines image configuration
type ImageConfig struct {
	DockerImage string   `yaml:"docker_image" json:"docker_image"`
	Command     []string `yaml:"command" json:"command"`
	WorkDir     string   `yaml:"workdir" json:"workdir"`
}

// Container represents a budgie container
type Container struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	State         ContainerState  `json:"state"`
	Image         ImageConfig     `json:"image"`
	Ports         []PortMapping   `json:"ports"`
	Volumes       []VolumeMapping `json:"volumes"`
	Env           []string        `json:"env"`
	Health        *HealthCheck    `json:"health_check,omitempty"`
	Replicas      *ReplicasConfig `json:"replicas,omitempty"`
	Resources     *ResourceLimits `json:"resources,omitempty"`
	RestartPolicy *RestartPolicy  `json:"restart_policy,omitempty"`
	DependsOn     []string        `json:"depends_on,omitempty"`
	Network       string          `json:"network,omitempty"`
	NetworkConfig *NetworkConfig  `json:"network_config,omitempty"`

	// Runtime fields
	BundlePath   string    `json:"-"`                       // Path to .bun file
	NodeID       string    `json:"node_id"`                 // Primary node ID
	Peers        []string  `json:"peers"`                   // Replica node IDs
	CreatedAt    time.Time `json:"created_at"`
	StartedAt    time.Time `json:"started_at"`
	ExitedAt     time.Time `json:"exited_at,omitempty"`
	Pid          int       `json:"pid"`                     // Container process ID
	RestartCount int       `json:"restart_count,omitempty"` // Number of times container has been restarted
}

// NetworkConfig defines network settings for a container
type NetworkConfig struct {
	IPAddress   string   `json:"ip_address,omitempty"`
	Gateway     string   `json:"gateway,omitempty"`
	DNS         []string `json:"dns,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ExtraHosts  []string `json:"extra_hosts,omitempty"`
}

// GenerateContainerID generates a unique 64-character hex container ID
func GenerateContainerID() string {
	b := make([]byte, 32)
	for {
		if _, err := rand.Read(b); err != nil {
			panic(err) // This shouldn't happen
		}
		id := hex.EncodeToString(b)
		// make sure that the truncated ID does not consist of only numeric
		// characters, as it's used as default hostname for containers.
		//
		// See: https://github.com/moby/moby/blob/master/daemon/internal/stringid/stringid.go#L39-L50
		isNumeric := true
		for _, c := range id[:12] {
			if c < '0' || c > '9' {
				isNumeric = false
				break
			}
		}
		if !isNumeric {
			return id
		}
	}
}

// ShortID returns the first 12 characters of the container ID
func (c *Container) ShortID() string {
	if len(c.ID) >= 12 {
		return c.ID[:12]
	}
	return c.ID
}

// IsRunning returns true if the container is in running state
func (c *Container) IsRunning() bool {
	return c.State == StateRunning
}

// IsStopped returns true if the container is in stopped state
func (c *Container) IsStopped() bool {
	return c.State == StateStopped
}
