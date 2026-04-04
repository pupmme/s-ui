package service

import (
	"encoding/json"
	"sync"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/core"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
)

type Core struct {
	started bool
	mu      sync.Mutex
}

func NewCore() *Core {
	return &Core{}
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

func (c *Core) GetInstance() *core.Core {
	return core.GetCore()
}

// buildSingboxConfig builds a minimal valid sing-box config from db.Config.
func buildSingboxConfig() ([]byte, error) {
	logCfg := config.Get().Log
	level := "info"
	if logCfg.Level == "debug" {
		level = "debug"
	}

	// Collect inbounds from db
	cfg := db.Get()
	inbounds := make([]json.RawMessage, 0, len(cfg.Inbounds))
	for _, in := range cfg.Inbounds {
		inMap := map[string]interface{}{
			"type": in.Type,
			"tag":  in.Tag,
		}
		// Parse address from addrs if present
		if len(in.Addrs) > 0 {
			var addr struct {
				Listen    string `json:"listen"`
				ListenPort int    `json:"listen_port"`
			}
			json.Unmarshal(in.Addrs, &addr)
			if addr.Listen != "" {
				inMap["listen"] = addr.Listen
			}
			if addr.ListenPort != 0 {
				inMap["listen_port"] = addr.ListenPort
			}
		}
		// Apply TLS if present
		if in.TlsId > 0 && in.Tls != nil && in.Tls.Server != nil {
			var tlsCfg map[string]interface{}
			json.Unmarshal(in.Tls.Server, &tlsCfg)
			if tlsCfg != nil {
				inMap["tls"] = tlsCfg
			}
		}
		jb, _ := json.Marshal(inMap)
		inbounds = append(inbounds, json.RawMessage(jb))
	}

	// Build outbounds: direct + block + dns
	outbounds := []json.RawMessage{
		json.RawMessage(`{"type":"direct","tag":"direct"}`),
		json.RawMessage(`{"type":"block","tag":"block"}`),
		json.RawMessage(`{"type":"direct","tag":"dns"}`),
	}

	// Add WARP endpoints if any
	for _, ep := range cfg.Endpoints {
		if len(ep.Options) > 0 {
			outbounds = append(outbounds, ep.Options)
		}
	}

	root := map[string]interface{}{
		"log": map[string]interface{}{
			"level": level,
		},
		"inbounds":  inbounds,
		"outbounds": outbounds,
	}

	return json.MarshalIndent(root, "", "  ")
}
