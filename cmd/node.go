package cmd

import (
	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/logger"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Node management (xboard agent)",
	Run: func(cmd *cobra.Command, args []string) {
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
