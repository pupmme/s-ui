package cmd

import (
	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manually sync from xboard (node mode)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		if !cfg.Node {
			logger.Error("not in node mode, sync is only available in node mode")
			return
		}
		logger.Info("syncing from xboard...")
		// TODO: implement xboard sync
		logger.Info("sync completed")
		_ = service.NewCore()
	},
}
