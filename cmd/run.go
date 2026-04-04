package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"
	"github.com/pupmme/sub/web"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sub panel",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		logger.Info("sub @ ", version, " starting...")

		err := config.Load()
		if err != nil {
			logger.Error("load config failed: ", err)
			os.Exit(1)
		}
		subCfg := config.Get()
		logger.Info("node mode: ", subCfg.Node)
		if subCfg.Node {
			logger.Info("xboard api: ", subCfg.Xboard.ApiHost)
		}

		core := service.NewCore()
		if err := core.Start(); err != nil {
			logger.Error("start core failed: ", err)
			os.Exit(1)
		}
		defer core.Close()

		go core.Ticker()
		logger.Info("core started")

		web.Start()

		logger.Info("sub is running. PID: ", os.Getpid())
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		logger.Info("shutting down...")
	},
}
