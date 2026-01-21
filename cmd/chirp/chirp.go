package chirp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/discovery"
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

func joinContainer(containerID string) error {
	fmt.Printf("Joining container %s as peer...\n", containerID)

	// Discover the target container
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

	fmt.Printf("Found container: %s (%s)\n", target.Name, target.Image)

	if len(target.IPs) == 0 {
		return fmt.Errorf("no IP addresses found for container")
	}

	remoteIP := target.IPs[0]
	fmt.Printf("Connecting to primary node at %s:%d...\n", remoteIP, target.Port)

	// TODO: Implement actual container pulling and volume sync
	// For now, provide instructions
	fmt.Println("\nTo complete the replication:")
	fmt.Printf("  1. Pull the image: docker pull %s\n", target.Image)
	fmt.Printf("  2. Create a matching .bun file\n")
	fmt.Printf("  3. Run: budgie run <your-file>.bun\n")
	fmt.Println("\nFull automated replication coming in a future release.")

	return nil
}

func GetChirpCmd() *cobra.Command {
	return chirpCmd
}

func init() {
	chirpCmd.Aliases = []string{"discover"}
}
