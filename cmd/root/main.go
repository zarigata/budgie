package main

import (
	"github.com/zarigata/budgie/cmd/chirp"
	"github.com/zarigata/budgie/cmd/nest"
	"github.com/zarigata/budgie/cmd/ps"
	"github.com/zarigata/budgie/cmd/root"
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

	rootCmd.Execute()
}
