package cmd

import (
	"github.com/pupmme/pupmsub/config"
	"github.com/pupmme/pupmsub/logger"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Node management (xboard agent)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Load(); err != nil {
			logger.Error("load config: ", err)
			return
		}
		cfg := config.Get()
		logger.Info("Node mode: ", cfg.Node)
		if cfg.Node {
			logger.Info("API Host: ", cfg.Xboard.ApiHost)
			logger.Info("Node ID: ", cfg.Xboard.NodeID)
			logger.Info("Node Type: ", cfg.Xboard.NodeType)
		} else {
			logger.Info("Not in node mode")
		}
	},
}
