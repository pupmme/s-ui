package service

import (
	"encoding/json"

	"github.com/pupmme/sub/core"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) GetConfig() ([]byte, error) {
	cfg := db.Get()
	return json.MarshalIndent(cfg, "", "  ")
}

func (s *ConfigService) StartCore() error {
	c := core.GetCore()
	if c.IsRunning() {
		return nil
	}
	data, err := buildSingboxConfig()
	if err != nil {
		return err
	}
	return c.Start(data)
}

func (s *ConfigService) StopCore() error {
	return core.GetCore().Stop()
}

func (s *ConfigService) RestartCore() error {
	c := core.GetCore()
	c.Stop()
	data, err := buildSingboxConfig()
	if err != nil {
		return err
	}
	return c.Start(data)
}

func (s *ConfigService) Save(obj string, act string, data json.RawMessage) error {
	is := &InboundService{}
	cs := &ClientService{}
	switch obj {
	case "inbound":
		return is.Save(nil, act, data, "", "")
	case "client":
		_, err := cs.Save(nil, act, data, "")
		return err
	default:
		logger.Debug("Config.Save: unhandled object ", obj, " ", act)
		return nil
	}
}
