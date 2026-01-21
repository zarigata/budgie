package inspect

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/runtime"
	"github.com/zarigata/budgie/pkg/types"
)

var (
	format string
	size   bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <container-id> [container-id...]",
	Short: "Display detailed information on containers",
	Long: `Return low-level information on containers as JSON.

You can specify multiple container IDs to inspect.`,
	Args: cobra.MinimumNArgs(1),
	RunE: inspectContainers,
}

func inspectContainers(cmd *cobra.Command, args []string) error {
	rt, err := runtime.GetDefaultRuntime()
	if err != nil {
		return fmt.Errorf("failed to get runtime: %w", err)
	}

	dataDir := os.Getenv("BUDGIE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/var/lib/budgie"
	}

	manager, err := api.NewContainerManager(rt, dataDir)
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	var results []interface{}
	var errors []string

	for _, idOrName := range args {
		ctr, err := findContainer(manager, idOrName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", idOrName, err))
			continue
		}

		// Create detailed inspection result
		result := map[string]interface{}{
			"Id":      ctr.ID,
			"Name":    ctr.Name,
			"Created": ctr.CreatedAt,
			"State": map[string]interface{}{
				"Status":     string(ctr.State),
				"Running":    ctr.IsRunning(),
				"Paused":     ctr.State == types.StatePaused,
				"StartedAt":  ctr.StartedAt,
				"FinishedAt": ctr.ExitedAt,
				"Pid":        ctr.Pid,
			},
			"Image": ctr.Image.DockerImage,
			"Config": map[string]interface{}{
				"Image":      ctr.Image.DockerImage,
				"Cmd":        ctr.Image.Command,
				"WorkingDir": ctr.Image.WorkDir,
				"Env":        ctr.Env,
			},
			"NetworkSettings": map[string]interface{}{
				"Ports": formatPortsForInspect(ctr.Ports),
			},
			"Mounts":    formatMountsForInspect(ctr.Volumes),
			"HostConfig": map[string]interface{}{
				"Binds":         formatBindsForInspect(ctr.Volumes),
				"PortBindings":  formatPortBindingsForInspect(ctr.Ports),
				"Resources":     ctr.Resources,
				"RestartPolicy": ctr.RestartPolicy,
			},
			"Budgie": map[string]interface{}{
				"NodeID":     ctr.NodeID,
				"Peers":      ctr.Peers,
				"BundlePath": ctr.BundlePath,
				"Health":     ctr.Health,
				"Replicas":   ctr.Replicas,
			},
		}

		results = append(results, result)
	}

	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "\nErrors:")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
	}

	return nil
}

func formatPortsForInspect(ports []types.PortMapping) map[string]interface{} {
	result := make(map[string]interface{})
	for _, p := range ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		key := fmt.Sprintf("%d/%s", p.ContainerPort, proto)
		result[key] = []map[string]string{
			{
				"HostIp":   "0.0.0.0",
				"HostPort": fmt.Sprintf("%d", p.HostPort),
			},
		}
	}
	return result
}

func formatMountsForInspect(volumes []types.VolumeMapping) []map[string]interface{} {
	var mounts []map[string]interface{}
	for _, v := range volumes {
		mount := map[string]interface{}{
			"Type":        "bind",
			"Source":      v.Source,
			"Destination": v.Target,
			"Mode":        v.Mode,
			"RW":          v.Mode != "ro",
		}
		mounts = append(mounts, mount)
	}
	return mounts
}

func formatBindsForInspect(volumes []types.VolumeMapping) []string {
	var binds []string
	for _, v := range volumes {
		bind := fmt.Sprintf("%s:%s", v.Source, v.Target)
		if v.Mode != "" {
			bind += ":" + v.Mode
		}
		binds = append(binds, bind)
	}
	return binds
}

func formatPortBindingsForInspect(ports []types.PortMapping) map[string]interface{} {
	result := make(map[string]interface{})
	for _, p := range ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		key := fmt.Sprintf("%d/%s", p.ContainerPort, proto)
		result[key] = []map[string]string{
			{
				"HostIp":   "",
				"HostPort": fmt.Sprintf("%d", p.HostPort),
			},
		}
	}
	return result
}

func findContainer(manager *api.ContainerManager, idOrName string) (*types.Container, error) {
	// Try exact match first
	if ctr, err := manager.Get(idOrName); err == nil {
		return ctr, nil
	}

	// Try prefix match
	containers := manager.List()
	var matches []*types.Container

	for _, ctr := range containers {
		if strings.HasPrefix(ctr.ID, idOrName) || ctr.Name == idOrName {
			matches = append(matches, ctr)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no such container")
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous container ID, multiple matches found")
	}

	return matches[0], nil
}

func GetInspectCmd() *cobra.Command {
	return inspectCmd
}

func init() {
	inspectCmd.Flags().StringVarP(&format, "format", "f", "", "Format output using a Go template")
	inspectCmd.Flags().BoolVarP(&size, "size", "s", false, "Display total file sizes")
}
