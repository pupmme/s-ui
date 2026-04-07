package cmd

import (
	"github.com/pupmme/pupmsub/config"
	"github.com/pupmme/pupmsub/logger"
	"github.com/pupmme/pupmsub/network"
	"github.com/pupmme/pupmsub/service"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync node data from xboard (node mode)",
	Run: func(cmd *cobra.Command, args []string) {
		// Must load config before reading it
		if err := config.Load(); err != nil {
			logger.Error("load config: ", err)
			return
		}
		cfg := config.Get()
		if !cfg.Node {
			logger.Error("sync is only available in node mode")
			return
		}
		sync := service.NewXboardSync(network.NewXboardClient())
		if err := sync.Sync(true); err != nil {
			logger.Error("sync failed: ", err)
			return
		}
		logger.Info("sync completed")
		// Reload core with new config
		if cs := service.GetCoreService(); cs != nil {
			_ = cs.Restart()
		}
	},
}
