package cmd

import (
	"github.com/pupmme/pupmmesub/logger"
	"github.com/pupmme/pupmmesub/service"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart sing-box",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("restarting sing-box...")
		core := service.NewCore()
		core.Close()
		if err := core.Start(); err != nil {
			logger.Error("restart failed: ", err)
			return
		}
		logger.Info("sing-box restarted")
	},
}
