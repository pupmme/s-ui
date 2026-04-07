package cmd

import (
	"github.com/pupmme/pupmsub/app"
	"github.com/spf13/cobra"
)

var configPath string
var dataPath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sub panel",
	Run: func(cmd *cobra.Command, args []string) {
		app.StartWithPaths(configPath, dataPath)
	},
}

func init() {
	runCmd.Flags().StringVarP(&configPath, "config", "c", "/etc/sub/config.json", "path to config file")
	runCmd.Flags().StringVarP(&dataPath, "data", "d", "/etc/sub/singbox.json", "path to sing-box data file")
}
