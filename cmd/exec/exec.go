package exec

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
	"github.com/zarigata/budgie/internal/runtime"
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
Use -t to allocate a pseudo-TTY.
Use -u to specify the user to run as.
Use -w to specify the working directory.
Use -e to set environment variables.`,
	Args: cobra.MinimumNArgs(2),
	RunE: execCommand,
}

func execCommand(cmd *cobra.Command, args []string) error {
	containerID := args[0]
	execArgs := args[1:]

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

	// Ensure container is running
	if err := cmdutil.RequireRunning(ctr); err != nil {
		return err
	}

	ctx := context.Background()

	// Build exec options
	opts := runtime.ExecOptions{
		Cmd:         execArgs,
		Interactive: interactive || tty,
		TTY:         tty,
		Detach:      detach,
		User:        user,
		WorkDir:     workdir,
		Env:         envVars,
	}

	// Execute command
	exitCode, err := cmdCtx.Runtime.ExecWithOptions(ctx, ctr.ID, opts)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
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
