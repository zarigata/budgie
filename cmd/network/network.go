package network

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
	"github.com/zarigata/budgie/internal/network"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage networks",
	Long:  `Manage container networks for network isolation.`,
}

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List networks",
	RunE:    listNetworks,
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a network",
	Args:  cobra.ExactArgs(1),
	RunE:  createNetwork,
}

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a network",
	Args:    cobra.ExactArgs(1),
	RunE:    removeNetwork,
}

var inspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Display detailed network information",
	Args:  cobra.ExactArgs(1),
	RunE:  inspectNetwork,
}

var (
	subnet  string
	gateway string
	driver  string
)

func listNetworks(cmd *cobra.Command, args []string) error {
	dataDir := cmdutil.GetDataDir()
	nm, err := network.NewNetworkManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}

	networks := nm.ListNetworks()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER\tSUBNET")

	for _, net := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			net.ID[:12], net.Name, net.Driver, net.Subnet)
	}

	return w.Flush()
}

func createNetwork(cmd *cobra.Command, args []string) error {
	name := args[0]

	dataDir := cmdutil.GetDataDir()
	nm, err := network.NewNetworkManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}

	// Use defaults if not specified
	if subnet == "" {
		subnet = "172.21.0.0/16"
	}
	if gateway == "" {
		gateway = "172.21.0.1"
	}
	if driver == "" {
		driver = "bridge"
	}

	if err := nm.CreateNetwork(name, driver, subnet, gateway); err != nil {
		return err
	}

	fmt.Println(name)
	return nil
}

func removeNetwork(cmd *cobra.Command, args []string) error {
	name := args[0]

	dataDir := cmdutil.GetDataDir()
	nm, err := network.NewNetworkManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}

	if err := nm.RemoveNetwork(name); err != nil {
		return err
	}

	fmt.Println(name)
	return nil
}

func inspectNetwork(cmd *cobra.Command, args []string) error {
	name := args[0]

	dataDir := cmdutil.GetDataDir()
	nm, err := network.NewNetworkManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}

	net, err := nm.GetNetwork(name)
	if err != nil {
		return err
	}

	fmt.Printf("ID: %s\n", net.ID)
	fmt.Printf("Name: %s\n", net.Name)
	fmt.Printf("Driver: %s\n", net.Driver)
	fmt.Printf("Subnet: %s\n", net.Subnet)
	fmt.Printf("Gateway: %s\n", net.Gateway)
	fmt.Printf("Containers: %d\n", len(net.Containers))

	if len(net.Containers) > 0 {
		fmt.Println("Connected containers:")
		for _, ctr := range net.Containers {
			fmt.Printf("  - %s\n", ctr)
		}
	}

	return nil
}

func GetNetworkCmd() *cobra.Command {
	return networkCmd
}

func init() {
	networkCmd.AddCommand(lsCmd)
	networkCmd.AddCommand(createCmd)
	networkCmd.AddCommand(rmCmd)
	networkCmd.AddCommand(inspectCmd)

	createCmd.Flags().StringVar(&subnet, "subnet", "", "Subnet in CIDR format (default: 172.21.0.0/16)")
	createCmd.Flags().StringVar(&gateway, "gateway", "", "Gateway IP (default: 172.21.0.1)")
	createCmd.Flags().StringVarP(&driver, "driver", "d", "bridge", "Network driver")
}
