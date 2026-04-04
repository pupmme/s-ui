package service

import (
	"sync"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/core"
	"github.com/pupmme/sub/cronjob"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
	"time"
)

var (
	corePtr  *core.Core
	coreOnce sync.Once
)

func GetCore() *core.Core {
	coreOnce.Do(func() {
		corePtr = core.New("/etc/sub/singbox.json")
	})
	return corePtr
}

type Core struct{}

func NewCore() *Core {
	return &Core{}
}

func (c *Core) Start() error {
	if err := config.Load(); err != nil {
		logger.Warning("load config: ", err)
	}
	if err := db.Load("/etc/sub/singbox.json"); err != nil {
		logger.Warning("load db: ", err)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.Local
	}
	cronjob.NewCronJob().Start(loc, 30)
	return GetCore().Start()
}

func (c *Core) Close() {
	GetCore().Close()
}

func (c *Core) Ticker() {
	GetCore().Ticker()
}

func (c *Core) GetInstance() *core.Core {
	return GetCore()
}
