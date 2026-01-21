package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
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

	// Initialize command context
	cmdCtx, err := cmdutil.NewCommandContext()
	if err != nil {
		return err
	}

	// Find container by ID prefix or name
	ctr, err := cmdutil.FindContainer(cmdCtx.Manager, containerID)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Parse --since flag if provided
	var sinceTime time.Time
	if since != "" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return fmt.Errorf("invalid --since value %q: %w", since, err)
		}
		sinceTime = time.Now().Add(-duration)
	}

	// Get logs reader
	reader, err := cmdCtx.Runtime.Logs(ctx, ctr.ID, follow, tail)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	// Stream logs to stdout
	if follow {
		fmt.Fprintf(os.Stderr, "Streaming logs for %s (Ctrl+C to stop)...\n", ctr.ShortID())
	}

	// Use buffered reader for line-by-line processing if timestamps or since is set
	if timestamps || !sinceTime.IsZero() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			// If --since is set, we'd need actual log timestamps to filter
			// For now, we just output with optional timestamp prefix
			if timestamps {
				fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), line)
			} else {
				fmt.Println(line)
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			return fmt.Errorf("error reading logs: %w", err)
		}
	} else {
		_, err = io.Copy(os.Stdout, reader)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading logs: %w", err)
		}
	}

	return nil
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
