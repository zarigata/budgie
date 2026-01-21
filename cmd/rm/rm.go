package rm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/runtime"
	"github.com/zarigata/budgie/pkg/types"
)

var (
	force   bool
	volumes bool
)

var rmCmd = &cobra.Command{
	Use:     "rm <container-id> [container-id...]",
	Aliases: []string{"remove", "delete"},
	Short:   "Remove one or more containers",
	Long: `Remove one or more stopped containers.

Use --force to remove running containers (they will be stopped first).
Use --volumes to also remove associated volumes.`,
	Args: cobra.MinimumNArgs(1),
	RunE: removeContainers,
}

func removeContainers(cmd *cobra.Command, args []string) error {
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

	ctx := context.Background()
	var errors []string
	var removed []string

	for _, idOrName := range args {
		// Find container by ID prefix or name
		ctr, err := findContainer(manager, idOrName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", idOrName, err))
			continue
		}

		// Check if running
		if ctr.IsRunning() {
			if !force {
				errors = append(errors, fmt.Sprintf("%s: container is running (use --force to remove)", ctr.ShortID()))
				continue
			}

			// Stop the container first
			fmt.Printf("Stopping container %s...\n", ctr.ShortID())
			if err := manager.Stop(ctx, ctr.ID, 10*time.Second); err != nil {
				errors = append(errors, fmt.Sprintf("%s: failed to stop: %v", ctr.ShortID(), err))
				continue
			}
		}

		// Remove the container
		if err := manager.Remove(ctx, ctr.ID); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to remove: %v", ctr.ShortID(), err))
			continue
		}

		// Remove volumes if requested
		if volumes && len(ctr.Volumes) > 0 {
			for _, vol := range ctr.Volumes {
				if vol.Mode == "rw" {
					// Only remove volumes within our data directory for safety
					if strings.HasPrefix(vol.Source, dataDir) {
						if err := os.RemoveAll(vol.Source); err != nil {
							fmt.Printf("Warning: failed to remove volume %s: %v\n", vol.Source, err)
						} else {
							fmt.Printf("Removed volume: %s\n", vol.Source)
						}
					}
				}
			}
		}

		removed = append(removed, ctr.ShortID())
		fmt.Println(ctr.ShortID())
	}

	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "\nErrors:")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
		if len(removed) == 0 {
			return fmt.Errorf("no containers removed")
		}
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
		return nil, fmt.Errorf("no such container")
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous container ID, multiple matches found")
	}

	return matches[0], nil
}

func GetRmCmd() *cobra.Command {
	return rmCmd
}

func init() {
	rmCmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal of running containers")
	rmCmd.Flags().BoolVarP(&volumes, "volumes", "v", false, "Remove associated volumes")
}
