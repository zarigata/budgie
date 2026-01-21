package logs

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/runtime"
	"github.com/zarigata/budgie/pkg/types"
)

var (
	follow     bool
	tail       int
	timestamps bool
	since      string
)

var logsCmd = &cobra.Command{
	Use:   "logs <container-id>",
	Short: "Fetch the logs of a container",
	Long: `Fetch the logs of a container.

Use --follow to stream logs in real-time.
Use --tail to show only the last N lines.`,
	Args: cobra.ExactArgs(1),
	RunE: fetchLogs,
}

func fetchLogs(cmd *cobra.Command, args []string) error {
	containerID := args[0]

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

	// Find container by ID prefix or name
	ctr, err := findContainer(manager, containerID)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get logs reader
	reader, err := rt.Logs(ctx, ctr.ID, follow, tail)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	// Stream logs to stdout
	if follow {
		fmt.Fprintf(os.Stderr, "Streaming logs for %s (Ctrl+C to stop)...\n", ctr.ShortID())
	}

	_, err = io.Copy(os.Stdout, reader)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
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
		return nil, fmt.Errorf("no such container: %s", idOrName)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous container ID, multiple matches found")
	}

	return matches[0], nil
}

func GetLogsCmd() *cobra.Command {
	return logsCmd
}

func init() {
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVarP(&tail, "tail", "n", 0, "Number of lines to show from the end")
	logsCmd.Flags().BoolVarP(&timestamps, "timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().StringVar(&since, "since", "", "Show logs since timestamp (e.g., 2h, 30m)")
}
