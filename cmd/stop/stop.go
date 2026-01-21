package stop

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/budgie/budgie/internal/api"
	"github.com/budgie/budgie/internal/runtime"
	"github.com/budgie/budgie/pkg/types"
	"github.com/spf13/cobra"
)

var (
	timeout time.Duration
)

var stopCmd = &cobra.Command{
	Use:   "stop <container-id>",
	Short: "Stop a running container",
	Long: `Stop a running container gracefully.

You can specify a timeout with --timeout flag. Default is 10 seconds.`,
	Args: cobra.ExactArgs(1),
	RunE: stopContainer,
}

func stopContainer(cmd *cobra.Command, args []string) error {
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

	ctr, err := manager.Get(containerID)
	if err != nil {
		return fmt.Errorf("container not found: %s", containerID)
	}

	if !ctr.IsRunning() {
		return fmt.Errorf("container is not running: %s (current state: %s)", ctr.ShortID(), ctr.State)
	}

	fmt.Printf("🛑 Stopping container %s...\n", ctr.ShortID())

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	if err := manager.Stop(context.Background(), ctr.ID, timeout); err != nil {
		return err
	}

	fmt.Printf("✅ Container %s stopped\n", ctr.ShortID())

	printContainerInfo(ctr)

	return nil
}

func GetStopCmd() *cobra.Command {
	return stopCmd
}

func init() {
	stopCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "Timeout before forcefully killing container")
}
