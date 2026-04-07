package service

import (
	"encoding/json"
	"sync"

	"github.com/pupmme/pupmsub/config"
	"github.com/pupmme/pupmsub/core"
	"github.com/pupmme/pupmsub/db"
	"github.com/pupmme/pupmsub/logger"
)

type Core struct {
	started bool
	mu      sync.Mutex
}

func NewCore() *Core {
	coreServiceInstance = &Core{}
	return coreServiceInstance
}

func (c *Core) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return nil
	}
	if err := config.Load(); err != nil {
		logger.Warning("load config: ", err)
	}
	if err := db.Load(db.DataPath()); err != nil {
		logger.Warning("load db: ", err)
	}

	data, err := buildSingboxConfig()
	if err != nil {
		return err
	}
	c.started = true
	return core.GetCore().Start(data)
}

func (c *Core) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		core.GetCore().Stop()
		c.started = false
	}
}

func (c *Core) Restart() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	core.GetCore().Stop()
	c.started = false
	if err := config.Load(); err != nil {
		logger.Warning("reload config: ", err)
	}
	if err := db.Load(db.DataPath()); err != nil {
		logger.Warning("reload db: ", err)
	}
	data, err := buildSingboxConfig()
	if err != nil {
		return err
	}
	c.started = true
	return core.GetCore().Start(data)
}

var coreServiceInstance *Core

// GetCoreService returns the global CoreService instance.
func GetCoreService() *Core {
	return coreServiceInstance
}

func (c *Core) IsRunning() bool {
	return core.GetCore().IsRunning()
}

func (c *Core) GetInstance() *core.Core {
	return core.GetCore()
}

// buildSingboxConfig builds a complete sing-box config from db.Config.
// db.Inbound.Options contains the raw sing-box JSON for each inbound.
// db.Outbound.Options contains the raw sing-box JSON for each outbound.
func buildSingboxConfig() ([]byte, error) {
	logCfg := config.Get().Log
	level := "info"
	if logCfg.Level == "debug" {
		level = "debug"
	}

	cfg := db.Get()

	// Build inbounds: merge Addrs + TLS + Options (Options contains full sing-box fields)
	inbounds := make([]json.RawMessage, 0, len(cfg.Inbounds))
	for _, in := range cfg.Inbounds {
		// Start from Options (raw sing-box JSON for this inbound)
		var inMap map[string]interface{}
		if len(in.Options) > 0 {
			json.Unmarshal(in.Options, &inMap)
		}
		if inMap == nil {
			inMap = make(map[string]interface{})
		}

		// Always set type and tag from db.Inbound
		inMap["type"] = in.Type
		inMap["tag"] = in.Tag

		// Merge listen/port from Addrs
		if len(in.Addrs) > 0 {
			var addr struct {
				Listen    string `json:"listen"`
				ListenPort int    `json:"listen_port"`
			}
			json.Unmarshal(in.Addrs, &addr)
			if addr.Listen != "" {
				inMap["listen"] = addr.Listen
			} else {
				inMap["listen"] = "0.0.0.0"
			}
			if addr.ListenPort != 0 {
				inMap["listen_port"] = addr.ListenPort
			}
		} else {
			// Defaults
			if _, ok := inMap["listen"]; !ok {
				inMap["listen"] = "0.0.0.0"
			}
			if _, ok := inMap["listen_port"]; !ok {
				inMap["listen_port"] = 2053
			}
		}

		// Merge TLS config from db.TLS
		if in.TlsId > 0 && in.Tls != nil && in.Tls.Server != nil {
			var tlsCfg map[string]interface{}
			json.Unmarshal(in.Tls.Server, &tlsCfg)
			if tlsCfg != nil && len(tlsCfg) > 0 {
				inMap["tls"] = tlsCfg
			}
		}

		jb, _ := json.Marshal(inMap)
		inbounds = append(inbounds, json.RawMessage(jb))
	}

	// Build outbounds: use db.Outbound.Options (raw sing-box JSON)
	outbounds := []json.RawMessage{
		json.RawMessage(`{"type":"direct","tag":"direct"}`),
		json.RawMessage(`{"type":"block","tag":"block"}`),
	}

	for _, out := range cfg.Outbounds {
		if len(out.Options) > 0 {
			outbounds = append(outbounds, out.Options)
		}
	}

	// Add WARP endpoints
	for _, ep := range cfg.Endpoints {
		if len(ep.Options) > 0 {
			outbounds = append(outbounds, ep.Options)
		}
	}

	// Build root config
	root := map[string]interface{}{
		"log": map[string]interface{}{
			"level": level,
			"timestamp": true,
		},
		"inbounds":  inbounds,
		"outbounds": outbounds,
	}

	return json.MarshalIndent(root, "", "  ")
}
