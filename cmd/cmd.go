package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sub",
		Short: "sub - sing-box management panel",
		Long:  "sub - sing-box management panel with xboard node agent",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(versionCmd)
}
