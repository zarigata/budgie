package chirp

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/discovery"
	"github.com/zarigata/budgie/internal/runtime"
	budgiesync "github.com/zarigata/budgie/internal/sync"
	"github.com/zarigata/budgie/pkg/types"
)

var chirpCmd = &cobra.Command{
	Use:   "chirp [container-id]",
	Short: "List containers on LAN or join as a peer",
	Long: `Without arguments, chirp lists all containers discoverable on local network.

With a container ID argument, chirp joins that container as a peer/replica,
downloading its image and data to provide redundancy.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runChirp,
}

func runChirp(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return listContainers()
	}
	return joinContainer(args[0])
}

func listContainers() error {
	fmt.Println("Scanning local network for budgie containers...")

	disc := discovery.NewDiscoveryService()
	timeout := 10 * time.Second

	containers, err := disc.DiscoverContainers(timeout)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("No containers found on network")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "CONTAINER ID\tNAME\tIP\tPORT\tIMAGE\tNODE\n")

	for _, ctr := range containers {
		shortID := ctr.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		ips := strings.Join(ctr.IPs, ",")
		if len(ips) > 20 {
			ips = ips[:20] + "..."
		}

		image := ctr.Image
		if len(image) > 20 {
			image = image[:20] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			shortID,
			ctr.Name,
			ips,
			ctr.Port,
			image,
			ctr.NodeID)
	}

	fmt.Printf("\nFound %d container(s) on network\n", len(containers))
	return nil
}

var (
	syncVolumes bool
	dryRun      bool
)

func joinContainer(containerID string) error {
	fmt.Printf("Joining container %s as peer...\n", containerID)

	// Step 1: Discover the target container
	fmt.Println("\n[1/5] Discovering container on network...")
	disc := discovery.NewDiscoveryService()
	containers, err := disc.DiscoverContainers(10 * time.Second)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	var target *discovery.DiscoveredContainer
	for i, ctr := range containers {
		if strings.HasPrefix(ctr.ID, containerID) {
			target = &containers[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("container %s not found on network", containerID)
	}

	fmt.Printf("    Found: %s (%s)\n", target.Name, target.Image)
	fmt.Printf("    Primary node: %s\n", target.NodeID)

	if len(target.IPs) == 0 {
		return fmt.Errorf("no IP addresses found for container")
	}

	remoteIP := target.IPs[0]
	fmt.Printf("    IP: %s:%d\n", remoteIP, target.Port)

	if dryRun {
		fmt.Println("\n[Dry run] Would perform the following:")
		fmt.Printf("  - Pull image: %s\n", target.Image)
		fmt.Printf("  - Create replica container\n")
		fmt.Printf("  - Sync volumes from %s:18733\n", remoteIP)
		return nil
	}

	// Step 2: Initialize runtime and manager
	fmt.Println("\n[2/5] Initializing runtime...")
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

	// Step 3: Create replica container configuration
	fmt.Println("\n[3/5] Creating replica container...")

	// Create local volume directories
	localVolumePath := filepath.Join(dataDir, "volumes", target.ID[:12])
	if err := os.MkdirAll(localVolumePath, 0755); err != nil {
		return fmt.Errorf("failed to create volume directory: %w", err)
	}

	hostname, _ := os.Hostname()

	replica := &types.Container{
		ID:        types.GenerateContainerID(),
		Name:      fmt.Sprintf("%s-replica", target.Name),
		State:     types.StateCreating,
		Image: types.ImageConfig{
			DockerImage: target.Image,
		},
		Ports: []types.PortMapping{
			{
				ContainerPort: target.Port,
				HostPort:      target.Port,
				Protocol:      "tcp",
			},
		},
		Volumes: []types.VolumeMapping{
			{
				Source: localVolumePath,
				Target: "/data",
				Mode:   "rw",
			},
		},
		NodeID:    hostname,
		Peers:     []string{target.NodeID},
		CreatedAt: time.Now(),
	}

	fmt.Printf("    Replica ID: %s\n", replica.ShortID())

	// Step 4: Create container (pulls image)
	fmt.Println("\n[4/5] Pulling image and creating container...")
	ctx := context.Background()

	if err := manager.Create(ctx, replica); err != nil {
		return fmt.Errorf("failed to create replica container: %w", err)
	}
	fmt.Printf("    Image pulled: %s\n", target.Image)

	// Step 5: Sync volumes if enabled
	if syncVolumes {
		fmt.Println("\n[5/5] Syncing volumes from primary...")
		syncAddr := fmt.Sprintf("%s:%d", remoteIP, budgiesync.DefaultSyncPort)

		conn, err := net.DialTimeout("tcp", syncAddr, 10*time.Second)
		if err != nil {
			fmt.Printf("    Warning: Could not connect to sync server at %s: %v\n", syncAddr, err)
			fmt.Println("    Volume sync skipped. Container created without data.")
		} else {
			defer conn.Close()

			syncMgr, err := budgiesync.NewSyncManager(localVolumePath)
			if err != nil {
				fmt.Printf("    Warning: Failed to create sync manager: %v\n", err)
			} else {
				if err := syncMgr.ReceiveVolume(conn); err != nil {
					fmt.Printf("    Warning: Volume sync failed: %v\n", err)
				} else {
					fmt.Println("    Volume data synchronized successfully")
				}
			}
		}
	} else {
		fmt.Println("\n[5/5] Volume sync skipped (use --sync to enable)")
	}

	// Start the replica
	fmt.Println("\nStarting replica container...")
	if err := manager.Start(ctx, replica.ID); err != nil {
		return fmt.Errorf("failed to start replica: %w", err)
	}

	// Announce replica on network
	if err := disc.AnnounceContainer(replica); err != nil {
		fmt.Printf("Warning: Failed to announce replica: %v\n", err)
	}

	fmt.Printf("\n✅ Replica container %s is now running\n", replica.ShortID())
	fmt.Printf("   Name: %s\n", replica.Name)
	fmt.Printf("   Image: %s\n", replica.Image.DockerImage)
	fmt.Printf("   Primary: %s\n", target.NodeID)

	return nil
}

func GetChirpCmd() *cobra.Command {
	return chirpCmd
}

func init() {
	chirpCmd.Aliases = []string{"discover", "join"}
	chirpCmd.Flags().BoolVarP(&syncVolumes, "sync", "s", false, "Sync volumes from primary node")
	chirpCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
}
