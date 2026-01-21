package ps

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/runtime"
	"github.com/zarigata/budgie/pkg/types"
)

var (
	all    bool
	quiet  bool
	format string
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers",
	Long: `List all containers running on this machine.

Use --all to show stopped containers as well.
Use --quiet to only display container IDs.`,
	RunE: listContainers,
}

func listContainers(cmd *cobra.Command, args []string) error {
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

	containers := manager.List()

	// Filter containers based on flags
	var filtered []*types.Container
	for _, ctr := range containers {
		if all || ctr.State == types.StateRunning {
			filtered = append(filtered, ctr)
		}
	}

	if len(filtered) == 0 {
		if all {
			fmt.Println("No containers found")
		} else {
			fmt.Println("No running containers found (use --all to show stopped)")
		}
		return nil
	}

	// Quiet mode: only show IDs
	if quiet {
		for _, ctr := range filtered {
			fmt.Println(ctr.ShortID())
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "CONTAINER ID\tNAME\tIMAGE\tSTATUS\tPORTS\tCREATED\n")

	for _, ctr := range filtered {
		ports := formatPorts(ctr.Ports)
		created := formatTimeAgo(ctr.CreatedAt)
		status := formatStatus(ctr)
		image := ctr.Image.DockerImage
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			ctr.ShortID(),
			ctr.Name,
			image,
			status,
			ports,
			created)
	}

	return nil
}

func formatPorts(ports []types.PortMapping) string {
	if len(ports) == 0 {
		return "-"
	}

	var parts []string
	for _, p := range ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		parts = append(parts, fmt.Sprintf("%d->%d/%s", p.HostPort, p.ContainerPort, proto))
	}

	result := strings.Join(parts, ", ")
	if len(result) > 30 {
		return result[:27] + "..."
	}
	return result
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "Just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func formatStatus(ctr *types.Container) string {
	switch ctr.State {
	case types.StateRunning:
		uptime := formatTimeAgo(ctr.StartedAt)
		if uptime == "-" {
			return "Running"
		}
		return fmt.Sprintf("Up %s", strings.TrimSuffix(uptime, " ago"))
	case types.StateStopped:
		exitTime := formatTimeAgo(ctr.ExitedAt)
		if exitTime == "-" {
			return "Stopped"
		}
		return fmt.Sprintf("Exited (%s)", exitTime)
	case types.StateCreated:
		return "Created"
	case types.StatePaused:
		return "Paused"
	case types.StateFailed:
		return "Failed"
	default:
		return string(ctr.State)
	}
}

func GetPsCmd() *cobra.Command {
	return psCmd
}

func init() {
	psCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all containers (including stopped)")
	psCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only display container IDs")
	psCmd.Flags().StringVar(&format, "format", "", "Format output using Go template")
}
