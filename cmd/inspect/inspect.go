package inspect

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
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

You can specify multiple container IDs to inspect.
Use --format to extract specific fields using Go templates.

Examples:
  budgie inspect mycontainer
  budgie inspect --format '{{.Id}}' mycontainer
  budgie inspect --format '{{.State.Status}}' mycontainer
  budgie inspect --format '{{json .Config}}' mycontainer`,
	Args: cobra.MinimumNArgs(1),
	RunE: inspectContainers,
}

func inspectContainers(cmd *cobra.Command, args []string) error {
	// Initialize command context
	cmdCtx, err := cmdutil.NewCommandContext()
	if err != nil {
		return err
	}

	var results []interface{}
	var errors []string

	for _, idOrName := range args {
		ctr, err := cmdutil.FindContainer(cmdCtx.Manager, idOrName)
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

	// Output based on format flag
	if format != "" {
		// Use Go template for formatting
		tmpl, err := template.New("inspect").Funcs(template.FuncMap{
			"json": func(v interface{}) string {
				b, _ := json.Marshal(v)
				return string(b)
			},
		}).Parse(format)
		if err != nil {
			return fmt.Errorf("invalid format template: %w", err)
		}

		for _, result := range results {
			if err := tmpl.Execute(os.Stdout, result); err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}
			fmt.Println()
		}
	} else {
		// Default JSON output
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
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

func GetInspectCmd() *cobra.Command {
	return inspectCmd
}

func init() {
	inspectCmd.Flags().StringVarP(&format, "format", "f", "", "Format output using a Go template")
	inspectCmd.Flags().BoolVarP(&size, "size", "s", false, "Display total file sizes")
}
