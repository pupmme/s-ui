package cmd

import (
	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync node data from xboard (node mode)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		if !cfg.Node {
			logger.Error("sync is only available in node mode")
			return
		}
		sync := service.NewXboardSync()
		if err := sync.DoFullSync(); err != nil {
			logger.Error("sync failed: ", err)
			return
		}
		logger.Info("sync completed")
		// Reload core with new config
		_ = service.NewCore()
	},
}
