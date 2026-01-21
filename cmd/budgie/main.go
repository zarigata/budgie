package root

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "budgie",
	Short: "Distributed container orchestration tool",
	Long: `Budgie is a distributed container orchestration tool that simplifies
running and replicating containers across machines in a local network.`,
	Version: "0.1",
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
