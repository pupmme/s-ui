package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/service"
	"github.com/pupmme/sub/web"
)

var (
	configPath = "/etc/sub/config.json"
	dataPath   = "/etc/sub/singbox.json"
	coreServiceInstance *service.Core
	xboardDaemonInstance *service.XboardDaemon
)

func Start() error {
	return startWithPaths(configPath, dataPath)
}

func StartWithPaths(cfgPath, dPath string) error {
	configPath = cfgPath
	dataPath = dPath
	return startWithPaths(configPath, dataPath)
}

func startWithPaths(cfgPath, dPath string) error {
	logger.InitLogger()
	logger.Info("sub starting...")

	config.SetPath(cfgPath)
	if err := config.Load(); err != nil {
		logger.Error("load config: ", err)
	}
	if err := db.Load(dPath); err != nil {
		logger.Error("load db: ", err)
	}

	coreServiceInstance = service.NewCore()
	if err := coreServiceInstance.Start(); err != nil {
		logger.Error("start core: ", err)
	}

	// Start xboard daemon if node mode is enabled
	if config.Get().Node {
		xboardDaemonInstance = service.NewXboardDaemon()
		xboardDaemonInstance.Start()
	}

	webServer := web.NewServer()
	if err := webServer.Start(); err != nil {
		logger.Error("start web server: ", err)
		return err
	}

	logger.Info("sub is running. PID: ", os.Getpid())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	logger.Info("shutting down...")
	if xboardDaemonInstance != nil {
		xboardDaemonInstance.Stop()
	}
	coreServiceInstance.Close()
	webServer.Stop()
	return nil
}
