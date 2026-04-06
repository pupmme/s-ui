package app

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pupmme/pupmmesub/config"
	"github.com/pupmme/pupmmesub/cronjob"
	"github.com/pupmme/pupmmesub/db"
	"github.com/pupmme/pupmmesub/logger"
	"github.com/pupmme/pupmmesub/service"
	"github.com/pupmme/pupmmesub/web"
)

var (
	configPath = "/etc/sub/config.json"
	dataPath   = "/etc/sub/singbox.json"
	coreServiceInstance *service.Core
	xboardDaemonInstance *service.XboardDaemon
	cronJob             *cronjob.CronJob
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

	// Start cron jobs for traffic cleanup and core health
	settings := db.Get().Settings
	tz := settings["timeLoc"]
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, _ := time.LoadLocation(tz)
	trafficAge := 0
	if v, ok := settings["trafficAge"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			trafficAge = n
		}
	}
	cronJob = cronjob.NewCronJob()
	if err := cronJob.Start(loc, trafficAge); err != nil {
		logger.Error("start cron: ", err)
	} else {
		logger.Info("cron started")
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
	if cronJob != nil {
		cronJob.Stop()
	}
	coreServiceInstance.Close()
	webServer.Stop()
	return nil
}
