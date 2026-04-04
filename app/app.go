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

func Start() error {
	logger.InitLogger()
	logger.Info("sub starting...")

	if err := config.Load(); err != nil {
		logger.Error("load config: ", err)
	}
	if err := db.Load("/etc/sub/singbox.json"); err != nil {
		logger.Error("load db: ", err)
	}

	coreSvc := service.NewCore()
	if err := coreSvc.Start(); err != nil {
		logger.Error("start core: ", err)
	}

	go web.NewServer()

	logger.Info("sub is running. PID: ", os.Getpid())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	logger.Info("shutting down...")
	return nil
}
