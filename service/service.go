package service

import (
	"encoding/json"
	"sync"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/core"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
)

var (
	corePtr  *core.Core
	coreOnce sync.Once
)

func GetCore() *core.Core {
	coreOnce.Do(func() {
		corePtr = core.GetCore()
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
	cfg := db.Get()
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return GetCore().Start(data)
}

func (c *Core) Close() {
	GetCore().Stop()
}

func (c *Core) GetInstance() *core.Core {
	return GetCore()
}
