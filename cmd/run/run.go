package run

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/bundle"
	"github.com/zarigata/budgie/internal/runtime"
)

var (
	detach bool
	name   string
)

var runCmd = &cobra.Command{

	Use:   "run <filename.bun>",
	Short: "Run a budgie container from .bun file",
	Long: `Run creates and starts a new budgie container from the specified .bun file.

The container will be announced on the local network via mDNS, allowing
other machines to discover and replicate it.`,
	Args: cobra.ExactArgs(1),
	RunE: runContainer,
}

func runContainer(cmd *cobra.Command, args []string) error {
	filename := args[0]

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("bundle file not found: %s", filename)
	}

	bun, err := bundle.Parse(filename)
	if err != nil {
		return fmt.Errorf("failed to parse bundle: %w", err)
	}

	fmt.Printf("🐦 Running container from %s\n", filename)
	fmt.Printf("📦 Bundle version: %s\n", bun.Version)
	fmt.Printf("🏷️  Name: %s\n", bun.Name)
	fmt.Printf("🖼️  Image: %s\n", bun.Image.DockerImage)
	fmt.Printf("🔌 Ports: %d mappings\n", len(bun.Ports))
	fmt.Printf("📁 Volumes: %d mounts\n", len(bun.Volumes))
	fmt.Printf("🔧 Environment: %d variables\n", len(bun.Env))

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

	ctr := bun.ToContainer(filename)
	ctx := context.Background()

	fmt.Printf("\nCreating container...\n")
	if err := manager.Create(ctx, ctr); err != nil {
		return err
	}

	fmt.Printf("\nStarting container...\n")
	if err := manager.Start(ctx, ctr.ID); err != nil {
		return err
	}

	fmt.Printf("\n✅ Container %s is now running\n", ctr.ShortID())

	if !detach {
		fmt.Println("\nPress Ctrl+C to stop container...")
		<-ctx.Done()
		fmt.Println("\nStopping container...")

		timeout := 10 * time.Second
		if err := manager.Stop(context.Background(), ctr.ID, timeout); err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}

		fmt.Println("✅ Container stopped")
	}

	return nil
}

func GetRunCmd() *cobra.Command {
	return runCmd
}

func init() {
	runCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Run container in background")
	runCmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the container")
}
