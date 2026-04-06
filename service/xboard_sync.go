package service

import (
	"encoding/json"
	"fmt"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/network"
)

// XboardSync handles data synchronization between s-ui (node) and xboard (server).
type XboardSync struct {
	client *network.XboardClient
}

// NewXboardSync creates a new sync service with a shared xboard client.
func NewXboardSync(client *network.XboardClient) *XboardSync {
	return &XboardSync{client: client}
}

// Sync performs config + user sync. Pass initial=true to also do handshake.
// Uses Config/Users from handshake response when available to avoid extra round-trips.
func (s *XboardSync) Sync(initial bool) error {
	var hs *network.HandshakeResponse

	if initial {
		cfg := config.Get()
		if !cfg.Node {
			return fmt.Errorf("sync is only available in node mode")
		}
		var err error
		hs, err = s.client.Handshake()
		if err != nil {
			return fmt.Errorf("handshake failed: %w", err)
		}
		logger.Info("[xboard-sync] connected to xboard ", hs.Version)
	}

	// Apply config from handshake response if available (avoids extra round-trip)
	if hs != nil && len(hs.Config) > 0 {
		if err := s.applyInboundConfigRaw(hs.Config); err != nil {
			logger.Error("[xboard-sync] apply inbound config from handshake: ", err)
		} else {
			logger.Info("[xboard-sync] inbound config applied from handshake")
		}
	} else {
		// Fallback: explicit fetch
		nodeCfg, err := s.client.GetConfig()
		if err != nil {
			logger.Info("[xboard-sync] get config: ", err)
		} else if nodeCfg != nil {
			if err := s.applyInboundConfig(nodeCfg); err != nil {
				logger.Error("[xboard-sync] apply inbound config: ", err)
			} else {
				logger.Info("[xboard-sync] inbound config applied")
			}
		}
	}

	// Apply users from handshake response if available
	if hs != nil && len(hs.Users) > 0 {
		if err := s.applyUsersRaw(hs.Users); err != nil {
			logger.Error("[xboard-sync] apply users from handshake: ", err)
		} else {
			logger.Info("[xboard-sync] users applied from handshake")
		}
	} else {
		// Fallback: explicit fetch
		users, err := s.client.GetUsers()
		if err != nil {
			logger.Info("[xboard-sync] get users: ", err)
		} else if users != nil {
			if err := s.applyUsers(users); err != nil {
				logger.Error("[xboard-sync] apply users: ", err)
			} else {
				logger.Info("[xboard-sync] ", len(users), " users applied")
			}
		}
	}

	logger.Info("[xboard-sync] sync completed")
	return nil
}

// applyInboundConfigRaw parses raw config JSON from handshake and applies it.
func (s *XboardSync) applyInboundConfigRaw(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var cfg network.NodeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	return s.applyInboundConfig(&cfg)
}

// applyUsersRaw parses raw users JSON array from handshake and applies it.
func (s *XboardSync) applyUsersRaw(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var users []network.User
	if err := json.Unmarshal(raw, &users); err != nil {
		return fmt.Errorf("parse users: %w", err)
	}
	return s.applyUsers(users)
}

// applyInboundConfig applies the inbound configuration from xboard to local db.
func (s *XboardSync) applyInboundConfig(nodeCfg *network.NodeConfig) error {
	cfg := db.Get()

	// Find or create inbound matching this tag
	var targetIdx int = -1
	for i := range cfg.Inbounds {
		if cfg.Inbounds[i].Tag == nodeCfg.Tag {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		// Create new inbound
		newID := uint(1)
		for _, ib := range cfg.Inbounds {
			if ib.Id >= newID {
				newID = ib.Id + 1
			}
		}
		listen := "0.0.0.0"
		if nodeCfg.Listen != "" {
			listen = nodeCfg.Listen
		}
		cfg.Inbounds = append(cfg.Inbounds, db.Inbound{
			Id:     newID,
			Type:   nodeCfg.Protocol,
			Tag:    nodeCfg.Tag,
			Addrs:  []byte(fmt.Sprintf(`{"listen":"%s","listen_port":%d}`, listen, nodeCfg.Port)),
			OutJson: nodeCfg.Settings,
		})
		logger.Info("[xboard-sync] created inbound: ", nodeCfg.Tag)
	} else {
		// Update existing by slice index, not by business Id
		if nodeCfg.Listen != "" {
			cfg.Inbounds[targetIdx].Addrs = []byte(fmt.Sprintf(`{"listen":"%s","listen_port":%d}`, nodeCfg.Listen, nodeCfg.Port))
		}
		cfg.Inbounds[targetIdx].OutJson = nodeCfg.Settings
		logger.Info("[xboard-sync] updated inbound: ", nodeCfg.Tag)
	}

	db.Set(cfg)
	return db.SaveConfig()
}

// applyUsers applies the user list from xboard to local db.
// In node mode, users are managed by xboard — we mirror them locally.
func (s *XboardSync) applyUsers(users []network.User) error {
	cfg := db.Get()

	for _, u := range users {
		// Build xboard metadata stored in Config JSON
		meta := map[string]interface{}{
			"uuid":   u.UUID,
			"email":  u.Email,
			"flow":   u.Flow,
			"tg_id":  u.TgId,
			"sub_id": u.SubID,
		}
		metaJSON, _ := json.Marshal(meta)

		// Build description for display
		desc := u.Email
		if desc == "" {
			desc = fmt.Sprintf("xboard:%s", u.Username)
		}

		found := false
		for i := range cfg.Clients {
			if int64(cfg.Clients[i].Id) == u.ID {
				// Update existing user
				cfg.Clients[i].Enable = u.Enable
				cfg.Clients[i].Up = u.Upload
				cfg.Clients[i].Down = u.Download
				cfg.Clients[i].Volume = u.Total
				cfg.Clients[i].Expiry = u.ExpiryTime
				cfg.Clients[i].Desc = desc
				cfg.Clients[i].Config = metaJSON
				found = true
				logger.Debug("[xboard-sync] updated user: ", u.Username)
				break
			}
		}
		if !found {
			// Create new client
			cfg.Clients = append(cfg.Clients, db.Client{
				Id:       uint(u.ID),
				Name:     u.Username,
			Enable:   u.Enable,
				Up:       u.Upload,
				Down:     u.Download,
				Volume:   u.Total,
				Expiry:   u.ExpiryTime,
				Desc:     desc,
				Config:   metaJSON,
				Inbounds: []byte("[]"),
			})
			logger.Debug("[xboard-sync] added user: ", u.Username)
		}
	}

	db.Set(cfg)
	return db.SaveConfig()
}

// SyncWithHandshake applies the handshake response directly (no extra fetch needed).
func (s *XboardSync) SyncWithHandshake(hs *network.HandshakeResponse) error {
	if hs == nil {
		return fmt.Errorf("nil handshake response")
	}

	if len(hs.Config) > 0 {
		if err := s.applyInboundConfigRaw(hs.Config); err != nil {
			logger.Error("[xboard-sync] apply inbound config from handshake: ", err)
		} else {
			logger.Info("[xboard-sync] inbound config applied from handshake")
		}
	}

	if len(hs.Users) > 0 {
		if err := s.applyUsersRaw(hs.Users); err != nil {
			logger.Error("[xboard-sync] apply users from handshake: ", err)
		} else {
			logger.Info("[xboard-sync] ", len(hs.Users), " users applied from handshake")
		}
	}

	logger.Info("[xboard-sync] sync completed")
	return nil
}

// ReportTraffic sends traffic data to xboard.
func (s *XboardSync) ReportTraffic(traffic map[int64][2]int64) error {
	return s.client.Report(traffic, nil, nil, 0, [2]uint64{}, [2]uint64{}, [2]uint64{})
}
