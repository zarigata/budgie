package ps

import (
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/spf13/cobra"
)

var (
	all bool
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers",
	Long: `List all containers running on this machine.

Use --all to show stopped containers as well.`,
	RunE: listContainers,
}

func listContainers(cmd *cobra.Command, args []string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "CONTAINER ID\tNAME\tSTATUS\tPORTS\tCREATED\n")

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		"abc123456789def456789abc123456789",
		"example-app",
		"Running",
		"8080->80/tcp",
		"2 minutes ago")

	if all {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			"def456789abc123456789def456789abc",
			"old-app",
			"Stopped",
			"-",
			"1 hour ago")
	}

	return nil
}

func GetPsCmd() *cobra.Command {
	return psCmd
}

func init() {
	psCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all containers (including stopped)")
}
