package main

import (
	"github.com/budgie/budgie/cmd/chirp"
	"github.com/budgie/budgie/cmd/nest"
	"github.com/budgie/budgie/cmd/ps"
	"github.com/budgie/budgie/cmd/root"
	"github.com/budgie/budgie/cmd/run"
	"github.com/budgie/budgie/cmd/stop"
)

func main() {
	rootCmd := root.GetRootCmd()

	rootCmd.AddCommand(run.GetRunCmd())
	rootCmd.AddCommand(ps.GetPsCmd())
	rootCmd.AddCommand(chirp.GetChirpCmd())
	rootCmd.AddCommand(stop.GetStopCmd())
	rootCmd.AddCommand(nest.GetNestCmd())

	rootCmd.Execute()
}
