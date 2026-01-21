package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config represents the budgie configuration
type Config struct {
	// DataDir is the directory for storing budgie data
	DataDir string `yaml:"data_dir"`

	// ContainerdAddress is the address of the containerd socket
	ContainerdAddress string `yaml:"containerd_address"`

	// SyncPort is the default port for volume synchronization
	SyncPort int `yaml:"sync_port"`

	// TLS configuration for sync protocol
	TLS TLSConfig `yaml:"tls"`

	// Discovery configuration
	Discovery DiscoveryConfig `yaml:"discovery"`

	// Defaults for new containers
	Defaults ContainerDefaults `yaml:"defaults"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`
}

// TLSConfig holds TLS settings
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

// DiscoveryConfig holds discovery settings
type DiscoveryConfig struct {
	Enabled bool   `yaml:"enabled"`
	Domain  string `yaml:"domain"`
	Timeout int    `yaml:"timeout"` // seconds
}

// ContainerDefaults holds default settings for new containers
type ContainerDefaults struct {
	RestartPolicy string `yaml:"restart_policy"` // "no", "always", "on-failure", "unless-stopped"
	MaxRetries    int    `yaml:"max_retries"`    // For on-failure policy
	StopTimeout   int    `yaml:"stop_timeout"`   // Seconds to wait before killing
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // text, json
	File       string `yaml:"file"`        // Log file path (empty = stdout)
	MaxSize    int    `yaml:"max_size"`    // Max size in MB before rotation
	MaxBackups int    `yaml:"max_backups"` // Number of old log files to keep
}

var (
	globalConfig *Config
	configOnce   sync.Once
	configPaths  = []string{
		"budgie.yaml",
		"budgie.yml",
		".budgie.yaml",
		".budgie.yml",
	}
)

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		DataDir:           "/var/lib/budgie",
		ContainerdAddress: "/run/containerd/containerd.sock",
		SyncPort:          18733,
		TLS: TLSConfig{
			Enabled: false,
		},
		Discovery: DiscoveryConfig{
			Enabled: true,
			Domain:  "local",
			Timeout: 10,
		},
		Defaults: ContainerDefaults{
			RestartPolicy: "no",
			MaxRetries:    3,
			StopTimeout:   10,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			MaxSize:    100,
			MaxBackups: 3,
		},
	}
}

// Load reads configuration from file
func Load() (*Config, error) {
	var loadErr error
	configOnce.Do(func() {
		globalConfig = DefaultConfig()

		// Try to find config file
		configPath := findConfigFile()
		if configPath == "" {
			return // Use defaults
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			loadErr = fmt.Errorf("failed to read config file %s: %w", configPath, err)
			return
		}

		if err := yaml.Unmarshal(data, globalConfig); err != nil {
			loadErr = fmt.Errorf("failed to parse config file %s: %w", configPath, err)
			return
		}

		// Apply environment variable overrides
		applyEnvOverrides(globalConfig)
	})

	return globalConfig, loadErr
}

// Get returns the current configuration (loads if not already loaded)
func Get() *Config {
	cfg, _ := Load()
	return cfg
}

// findConfigFile searches for a config file in standard locations
func findConfigFile() string {
	// 1. Check environment variable
	if envPath := os.Getenv("BUDGIE_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// 2. Check current directory
	for _, name := range configPaths {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	// 3. Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		budgieDir := filepath.Join(homeDir, ".budgie")
		for _, name := range configPaths {
			path := filepath.Join(budgieDir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}

		// Also check ~/.config/budgie/
		configDir := filepath.Join(homeDir, ".config", "budgie")
		for _, name := range configPaths {
			path := filepath.Join(configDir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// 4. Check /etc/budgie/
	for _, name := range configPaths {
		path := filepath.Join("/etc/budgie", name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("BUDGIE_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("CONTAINERD_ADDRESS"); v != "" {
		cfg.ContainerdAddress = v
	}
	if v := os.Getenv("BUDGIE_SYNC_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.SyncPort = port
		}
	}
	if v := os.Getenv("BUDGIE_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
}

// Save writes the configuration to a file
func Save(cfg *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the user's config file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "budgie.yaml"
	}
	return filepath.Join(homeDir, ".budgie", "budgie.yaml")
}

// Init creates a default config file if none exists
func Init() error {
	path := GetConfigPath()

	// Check if already exists
	if _, err := os.Stat(path); err == nil {
		return nil // Already exists
	}

	return Save(DefaultConfig(), path)
}
