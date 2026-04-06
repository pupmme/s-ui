package service

import (
	"encoding/json"
	"time"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/util/common"
	"strings"
)

// ClientService provides read-only access to clients.
// In Node mode, data is sourced from xboard and should not be modified locally.
type ClientService struct{}

func (s *ClientService) Get(id string) (*[]db.Client, error) {
	if id == "" {
		return s.GetAll()
	}
	return s.getById(id)
}

func (s *ClientService) getById(id string) (*[]db.Client, error) {
	cfg := db.Get()
	idSet := make(map[string]bool)
	for _, i := range strings.Split(id, ",") {
		idSet[i] = true
	}
	var result []db.Client
	for _, c := range cfg.Clients {
		if idSet[common.Itoa(int(c.Id))] {
			result = append(result, c)
		}
	}
	return &result, nil
}

func (s *ClientService) GetAll() (*[]db.Client, error) {
	cfg := db.Get()
	result := make([]db.Client, 0, len(cfg.Clients))
	for _, c := range cfg.Clients {
		result = append(result, c)
	}
	return &result, nil
}

// Save is a no-op in Node mode. Clients are managed by xboard.
func (s *ClientService) Save(tx interface{}, act string, data json.RawMessage, hostname string) ([]uint, error) {
	return nil, common.NewError("clients are read-only in node mode")
}

// IsNodeMode returns true when pupmsub is acting as a node proxy to xboard.
func (s *ClientService) IsNodeMode() bool {
	return config.Get().Node
}

// UpdateClientsOnInboundAdd is a stub — inbound write logic lives in InboundService.
func (s *ClientService) UpdateClientsOnInboundAdd(initIds string, inboundId uint, hostname string) error {
	return nil
}

// UpdateLinksByInboundChange is a stub — inbound write logic lives in InboundService.
func (s *ClientService) UpdateLinksByInboundChange(inbound db.Inbound, oldTag string, hostname string) error {
	return nil
}

// UpdateClientsOnInboundDelete is a stub — inbound write logic lives in InboundService.
func (s *ClientService) UpdateClientsOnInboundDelete(id uint, tag string) error {
	return nil
}

func (s *ClientService) DepleteClients() ([]uint, error) {
	if s.IsNodeMode() {
		logger.Debug("DepleteClients skipped: node mode")
		return nil, nil
	}
	return s.depleteClientsImpl()
}

func (s *ClientService) ResetClients(dt int64) ([]uint, error) {
	if s.IsNodeMode() {
		logger.Debug("ResetClients skipped: node mode")
		return nil, nil
	}
	return s.resetClientsImpl(dt)
}

// depleteClientsImpl performs actual depletion logic.
func (s *ClientService) depleteClientsImpl() ([]uint, error) {
	dt := time.Now().Unix()
	inboundIds, err := s.resetClientsImpl(dt)
	if err != nil {
		return nil, err
	}

	cfg := db.Get()
	var changes []db.Change

	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if !c.Enable {
			continue
		}
		exceeded := c.Volume > 0 && (c.Up+c.Down) > c.Volume
		expired := c.Expiry > 0 && c.Expiry < dt
		if !exceeded && !expired {
			continue
		}
		logger.Debug("Client ", c.Name, " is going to be disabled")
		var userInbounds []uint
		_ = json.Unmarshal(c.Inbounds, &userInbounds)
		inboundIds = common.UnionUintArray(inboundIds, userInbounds)
		c.Enable = false
		changes = append(changes, db.Change{
			DateTime: dt,
			Actor:    "DepleteJob",
			Key:      "clients",
			Action:   "disable",
			Obj:      []byte("\"" + c.Name + "\""),
		})
	}

	if len(changes) > 0 {
		cfg.Changes = append(cfg.Changes, changes...)
		common.LastUpdate = dt
		db.Set(cfg)
		if err := db.SaveConfig(); err != nil {
			return nil, err
		}
	}

	return inboundIds, nil
}

// resetClientsImpl performs actual reset logic.
func (s *ClientService) resetClientsImpl(dt int64) ([]uint, error) {
	var inboundIds []uint
	cfg := db.Get()
	var changes []db.Change

	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if !c.Enable || !c.DelayStart || c.AutoReset {
			continue
		}
		if (c.Up+c.Down) > 0 {
			c.Expiry = dt + int64(c.ResetDays)*86400
			c.DelayStart = false
			changes = append(changes, db.Change{
				DateTime: dt,
				Actor:    "ResetJob",
				Key:      "clients",
				Action:   "reset",
				Obj:      []byte("\"" + c.Name + "\""),
			})
		}
	}

	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if !c.Enable || !c.DelayStart || !c.AutoReset {
			continue
		}
		if (c.Up+c.Down) > 0 {
			c.NextReset = dt + int64(c.ResetDays)*86400
			c.DelayStart = false
			changes = append(changes, db.Change{
				DateTime: dt,
				Actor:    "ResetJob",
				Key:      "clients",
				Action:   "reset",
				Obj:      []byte("\"" + c.Name + "\""),
			})
		}
	}

	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if c.DelayStart || !c.AutoReset {
			continue
		}
		if c.NextReset < dt {
			c.NextReset = dt + int64(c.ResetDays)*86400
			c.TotalUp += c.Up
			c.TotalDown += c.Down
			c.Up = 0
			c.Down = 0
			if !c.Enable {
				c.Enable = true
				var ids []uint
				_ = json.Unmarshal(c.Inbounds, &ids)
				inboundIds = common.UnionUintArray(inboundIds, ids)
			}
		}
	}

	if len(changes) > 0 {
		cfg.Changes = append(cfg.Changes, changes...)
		common.LastUpdate = dt
		db.Set(cfg)
		if err := db.SaveConfig(); err != nil {
			return nil, err
		}
	}

	return inboundIds, nil
}
