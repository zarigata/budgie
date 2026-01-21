package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	budgieconfig "github.com/zarigata/budgie/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage budgie configuration",
	Long: `View and manage budgie configuration.

Without subcommands, displays the current configuration.`,
	RunE: showConfig,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	Long:  `Create a default configuration file at ~/.budgie/budgie.yaml`,
	RunE:  initConfig,
}

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	RunE:  showPath,
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get the value of a specific configuration key.

Examples:
  budgie config get data_dir
  budgie config get tls.enabled
  budgie config get defaults.restart_policy`,
	Args: cobra.ExactArgs(1),
	RunE: getConfig,
}

func showConfig(cmd *cobra.Command, args []string) error {
	cfg := budgieconfig.Get()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println("# Current Budgie Configuration")
	fmt.Println("# Config file:", findConfigSource())
	fmt.Println()
	fmt.Print(string(data))

	return nil
}

func initConfig(cmd *cobra.Command, args []string) error {
	path := budgieconfig.GetConfigPath()

	// Check if already exists
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Configuration file already exists at: %s\n", path)
		return nil
	}

	if err := budgieconfig.Init(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	fmt.Printf("Created default configuration at: %s\n", path)
	return nil
}

func showPath(cmd *cobra.Command, args []string) error {
	source := findConfigSource()
	if source == "" {
		fmt.Println("No configuration file found (using defaults)")
		fmt.Printf("Default location: %s\n", budgieconfig.GetConfigPath())
	} else {
		fmt.Println(source)
	}
	return nil
}

func getConfig(cmd *cobra.Command, args []string) error {
	cfg := budgieconfig.Get()
	key := args[0]

	// Convert config to map for dynamic access
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var configMap map[string]interface{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Navigate to the key
	value, err := getNestedValue(configMap, key)
	if err != nil {
		return err
	}

	// Print the value
	switch v := value.(type) {
	case string:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	case int:
		fmt.Println(v)
	default:
		// For complex values, print as YAML
		data, _ := yaml.Marshal(v)
		fmt.Print(string(data))
	}

	return nil
}

func getNestedValue(m map[string]interface{}, key string) (interface{}, error) {
	parts := splitKey(key)
	current := interface{}(m)

	for _, part := range parts {
		switch c := current.(type) {
		case map[string]interface{}:
			val, ok := c[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", key)
			}
			current = val
		default:
			return nil, fmt.Errorf("key not found: %s", key)
		}
	}

	return current, nil
}

func splitKey(key string) []string {
	var parts []string
	current := ""
	for _, c := range key {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func findConfigSource() string {
	// Check environment variable first
	if path := os.Getenv("BUDGIE_CONFIG"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check standard locations
	locations := []string{
		"budgie.yaml",
		"budgie.yml",
		".budgie.yaml",
		".budgie.yml",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		for _, loc := range locations {
			path := fmt.Sprintf("%s/.budgie/%s", homeDir, loc)
			if _, err := os.Stat(path); err == nil {
				return path
			}
			path = fmt.Sprintf("%s/.config/budgie/%s", homeDir, loc)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	for _, loc := range locations {
		path := fmt.Sprintf("/etc/budgie/%s", loc)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func GetConfigCmd() *cobra.Command {
	return configCmd
}

func init() {
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(pathCmd)
	configCmd.AddCommand(getCmd)
}
