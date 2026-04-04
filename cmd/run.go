package cmd

import (
	"github.com/pupmme/sub/app"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sub panel",
	Run: func(cmd *cobra.Command, args []string) {
		app.Start()
	},
}
