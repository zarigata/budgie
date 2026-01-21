package main

import (
	"github.com/zarigata/budgie/cmd/chirp"
	budgieconfig "github.com/zarigata/budgie/cmd/config"
	"github.com/zarigata/budgie/cmd/exec"
	"github.com/zarigata/budgie/cmd/inspect"
	"github.com/zarigata/budgie/cmd/logs"
	"github.com/zarigata/budgie/cmd/nest"
	"github.com/zarigata/budgie/cmd/ps"
	"github.com/zarigata/budgie/cmd/rm"
	root "github.com/zarigata/budgie/cmd/budgie"
	"github.com/zarigata/budgie/cmd/run"
	"github.com/zarigata/budgie/cmd/stop"
)

func main() {
	rootCmd := root.GetRootCmd()

	rootCmd.AddCommand(run.GetRunCmd())
	rootCmd.AddCommand(ps.GetPsCmd())
	rootCmd.AddCommand(chirp.GetChirpCmd())
	rootCmd.AddCommand(stop.GetStopCmd())
	rootCmd.AddCommand(nest.GetNestCmd())
	rootCmd.AddCommand(rm.GetRmCmd())
	rootCmd.AddCommand(logs.GetLogsCmd())
	rootCmd.AddCommand(exec.GetExecCmd())
	rootCmd.AddCommand(inspect.GetInspectCmd())
	rootCmd.AddCommand(budgieconfig.GetConfigCmd())

	rootCmd.Execute()
}
