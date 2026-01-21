package exec

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/runtime"
	"github.com/zarigata/budgie/pkg/types"
)

var (
	interactive bool
	tty         bool
	detach      bool
	user        string
	workdir     string
	envVars     []string
)

var execCmd = &cobra.Command{
	Use:   "exec <container-id> <command> [args...]",
	Short: "Execute a command in a running container",
	Long: `Execute a command inside a running container.

Use -i for interactive mode (keeps STDIN open).
Use -t to allocate a pseudo-TTY.`,
	Args: cobra.MinimumNArgs(2),
	RunE: execCommand,
}

func execCommand(cmd *cobra.Command, args []string) error {
	containerID := args[0]
	execArgs := args[1:]

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

	if !ctr.IsRunning() {
		return fmt.Errorf("container %s is not running", ctr.ShortID())
	}

	ctx := context.Background()

	// Execute command
	exitCode, err := rt.Exec(ctx, ctr.ID, execArgs, interactive || tty)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	if exitCode != 0 {
		os.Exit(exitCode)
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

func GetExecCmd() *cobra.Command {
	return execCmd
}

func init() {
	execCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Keep STDIN open")
	execCmd.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")
	execCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Run command in background")
	execCmd.Flags().StringVarP(&user, "user", "u", "", "Username or UID")
	execCmd.Flags().StringVarP(&workdir, "workdir", "w", "", "Working directory inside the container")
	execCmd.Flags().StringArrayVarP(&envVars, "env", "e", nil, "Set environment variables")
}
