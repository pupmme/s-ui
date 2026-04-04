package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/db"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util"
	"github.com/alireza0/s-ui/util/common"
)

type ClientService struct{}

func (s *ClientService) Get(id string) (*[]model.Client, error) {
	if id == "" {
		return s.GetAll()
	}
	return s.getById(id)
}

func (s *ClientService) getById(id string) (*[]model.Client, error) {
	cfg := db.Get()
	idSet := make(map[string]bool)
	for _, i := range strings.Split(id, ",") {
		idSet[i] = true
	}
	var result []model.Client
	for _, c := range cfg.Clients {
		if idSet[common.Itoa(int(c.Id))] {
			result = append(result, s.dbClientToModel(c))
		}
	}
	return &result, nil
}

func (s *ClientService) GetAll() (*[]model.Client, error) {
	cfg := db.Get()
	result := make([]model.Client, 0, len(cfg.Clients))
	for _, c := range cfg.Clients {
		result = append(result, s.dbClientToModel(c))
	}
	return &result, nil
}

func (s *ClientService) dbClientToModel(c db.Client) model.Client {
	return model.Client{
		Id:         c.Id,
		Enable:     c.Enable,
		Name:       c.Name,
		Config:     c.Config,
		Inbounds:   c.Inbounds,
		Links:      c.Links,
		Volume:     c.Volume,
		Expiry:     c.Expiry,
		Up:         c.Up,
		Down:       c.Down,
		Desc:       c.Desc,
		Group:      c.Group,
		DelayStart: c.DelayStart,
		AutoReset:  c.AutoReset,
		ResetDays:  c.ResetDays,
		NextReset:  c.NextReset,
		TotalUp:    c.TotalUp,
		TotalDown:  c.TotalDown,
	}
}

func (s *ClientService) modelToDbClient(m model.Client) db.Client {
	return db.Client{
		Id:         m.Id,
		Enable:     m.Enable,
		Name:       m.Name,
		Config:     m.Config,
		Inbounds:   m.Inbounds,
		Links:      m.Links,
		Volume:     m.Volume,
		Expiry:     m.Expiry,
		Up:         m.Up,
		Down:       m.Down,
		Desc:       m.Desc,
		Group:      m.Group,
		DelayStart: m.DelayStart,
		AutoReset:  m.AutoReset,
		ResetDays:  m.ResetDays,
		NextReset:  m.NextReset,
		TotalUp:    m.TotalUp,
		TotalDown:  m.TotalDown,
	}
}

// Save handles CRUD for clients. tx is ignored in JSON mode.
func (s *ClientService) Save(tx interface{}, act string, data json.RawMessage, hostname string) ([]uint, error) {
	var err error
	var inboundIds []uint

	switch act {
	case "new", "edit":
		var client model.Client
		if err = json.Unmarshal(data, &client); err != nil {
			return nil, err
		}
		err = s.updateLinksWithFixedInboundsJSON([]*model.Client{&client}, hostname)
		if err != nil {
			return nil, err
		}
		if act == "edit" {
			inboundIds, err = s.findInboundsChangesJSON(&client, false)
			if err != nil {
				return nil, err
			}
		} else {
			_ = json.Unmarshal(client.Inbounds, &inboundIds)
		}
		cfg := db.Get()
		if act == "edit" {
			found := false
			for i := range cfg.Clients {
				if cfg.Clients[i].Id == client.Id {
					cfg.Clients[i] = s.modelToDbClient(client)
					found = true
					break
				}
			}
			if !found {
				cfg.Clients = append(cfg.Clients, s.modelToDbClient(client))
			}
		} else {
			maxId := uint(0)
			for _, c := range cfg.Clients {
				if c.Id > maxId {
					maxId = c.Id
				}
			}
			client.Id = maxId + 1
			cfg.Clients = append(cfg.Clients, s.modelToDbClient(client))
		}
		db.Set(cfg)
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}

	case "addbulk":
		var clients []*model.Client
		if err = json.Unmarshal(data, &clients); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(clients[0].Inbounds, &inboundIds)
		err = s.updateLinksWithFixedInboundsJSON(clients, hostname)
		if err != nil {
			return nil, err
		}
		cfg := db.Get()
		for _, client := range clients {
			maxId := uint(0)
			for _, c := range cfg.Clients {
				if c.Id > maxId {
					maxId = c.Id
				}
			}
			client.Id = maxId + 1
			cfg.Clients = append(cfg.Clients, s.modelToDbClient(*client))
		}
		db.Set(cfg)
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}

	case "editbulk":
		var clients []*model.Client
		if err = json.Unmarshal(data, &clients); err != nil {
			return nil, err
		}
		for _, client := range clients {
			changedInboundIds, err := s.findInboundsChangesJSON(client, true)
			if err != nil {
				return nil, err
			}
			if len(changedInboundIds) > 0 {
				inboundIds = common.UnionUintArray(inboundIds, changedInboundIds)
			}
		}
		if len(inboundIds) > 0 {
			err = s.updateLinksWithFixedInboundsJSON(clients, hostname)
			if err != nil {
				return nil, err
			}
		}
		cfg := db.Get()
		for _, client := range clients {
			found := false
			for i := range cfg.Clients {
				if cfg.Clients[i].Id == client.Id {
					cfg.Clients[i] = s.modelToDbClient(*client)
					found = true
					break
				}
			}
			if !found {
				cfg.Clients = append(cfg.Clients, s.modelToDbClient(*client))
			}
		}
		db.Set(cfg)
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}

	case "delbulk":
		var ids []uint
		if err = json.Unmarshal(data, &ids); err != nil {
			return nil, err
		}
		cfg := db.Get()
		for _, id := range ids {
			for _, c := range cfg.Clients {
				if c.Id == id {
					var clientInbounds []uint
					_ = json.Unmarshal(c.Inbounds, &clientInbounds)
					inboundIds = common.UnionUintArray(inboundIds, clientInbounds)
					break
				}
			}
		}
		newClients := make([]db.Client, 0, len(cfg.Clients))
		for _, c := range cfg.Clients {
			skip := false
			for _, id := range ids {
				if c.Id == id {
					skip = true
					break
				}
			}
			if !skip {
				newClients = append(newClients, c)
			}
		}
		cfg.Clients = newClients
		db.Set(cfg)
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}

	case "del":
		var id uint
		if err = json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		cfg := db.Get()
		for _, c := range cfg.Clients {
			if c.Id == id {
				_ = json.Unmarshal(c.Inbounds, &inboundIds)
				break
			}
		}
		newClients := make([]db.Client, 0, len(cfg.Clients))
		for _, c := range cfg.Clients {
			if c.Id != id {
				newClients = append(newClients, c)
			}
		}
		cfg.Clients = newClients
		db.Set(cfg)
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}

	default:
		return nil, common.NewErrorf("unknown action: %s", act)
	}

	return inboundIds, nil
}

func (s *ClientService) updateLinksWithFixedInboundsJSON(clients []*model.Client, hostname string) error {
	if len(clients) == 0 {
		return nil
	}

	// Collect inbound IDs from first client
	var inboundIds []uint
	if err := json.Unmarshal(clients[0].Inbounds, &inboundIds); err != nil {
		return err
	}

	// Load inbounds that support subscription links
	cfg := db.Get()
	var inbounds []model.Inbound
	for _, inb := range cfg.Inbounds {
		for _, id := range inboundIds {
			if inb.Id == id && slicesContains(util.InboundTypeWithLink, inb.Type) {
				inbounds = append(inbounds, model.Inbound{
					Id:      inb.Id,
					Type:    inb.Type,
					Tag:     inb.Tag,
					TlsId:   inb.TlsId,
					Options: inb.Options,
				})
				break
			}
		}
	}

	for _, client := range clients {
		var clientLinks []map[string]string
		_ = json.Unmarshal(client.Links, &clientLinks)
		newClientLinks := []map[string]string{}

		for _, inbound := range inbounds {
			newLinks := util.LinkGenerator(client.Config, &inbound, hostname)
			for _, newLink := range newLinks {
				newClientLinks = append(newClientLinks, map[string]string{
					"remark": inbound.Tag,
					"type":   "local",
					"uri":    newLink,
				})
			}
		}

		// Preserve non-local links
		for _, clientLink := range clientLinks {
			if clientLink["type"] != "local" {
				newClientLinks = append(newClientLinks, clientLink)
			}
		}

		clients[0].Links, _ = json.MarshalIndent(newClientLinks, "", "  ")
	}
	return nil
}

// UpdateClientsOnInboundAddJSON updates clients when a new inbound is added.
func (s *ClientService) UpdateClientsOnInboundAddJSON(initIds string, inboundId uint, hostname string) error {
	clientIds := strings.Split(initIds, ",")
	clientIdSet := make(map[string]bool)
	for _, c := range clientIds {
		clientIdSet[c] = true
	}

	// Load inbound
	cfg := db.Get()
	var inbound model.Inbound
	var found bool
	for _, inb := range cfg.Inbounds {
		if inb.Id == inboundId {
			inbound = model.Inbound{
				Id:      inb.Id,
				Type:    inb.Type,
				Tag:     inb.Tag,
				TlsId:   inb.TlsId,
				Options: inb.Options,
			}
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if !clientIdSet[common.Itoa(int(c.Id))] {
			continue
		}

		// Add inbound to client
		var clientInbounds []uint
		_ = json.Unmarshal(c.Inbounds, &clientInbounds)
		clientInbounds = append(clientInbounds, inboundId)
		c.Inbounds, _ = json.MarshalIndent(clientInbounds, "", "  ")

		// Add links
		var clientLinks, newClientLinks []map[string]string
		_ = json.Unmarshal(c.Links, &clientLinks)
		inbForLink := model.Inbound{
			Id:      inbound.Id,
			Type:    inbound.Type,
			Tag:     inbound.Tag,
			TlsId:   inbound.TlsId,
			Options: inbound.Options,
		}
		newLinks := util.LinkGenerator(c.Config, &inbForLink, hostname)
		for _, newLink := range newLinks {
			newClientLinks = append(newClientLinks, map[string]string{
				"remark": inbound.Tag,
				"type":   "local",
				"uri":    newLink,
			})
		}
		for _, clientLink := range clientLinks {
			if clientLink["remark"] != inbound.Tag {
				newClientLinks = append(newClientLinks, clientLink)
			}
		}
		c.Links, _ = json.MarshalIndent(newClientLinks, "", "  ")
	}

	db.Set(cfg)
	return database.SaveConfig()
}

// UpdateClientsOnInboundDeleteJSON updates clients when an inbound is deleted.
func (s *ClientService) UpdateClientsOnInboundDeleteJSON(id uint, tag string) error {
	cfg := db.Get()
	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		var clientInbounds []uint
		_ = json.Unmarshal(c.Inbounds, &clientInbounds)

		// Remove inbound ID
		var newInbounds []uint
		for _, ib := range clientInbounds {
			if ib != id {
				newInbounds = append(newInbounds, ib)
			}
		}
		c.Inbounds, _ = json.MarshalIndent(newInbounds, "", "  ")

		// Remove local links for this inbound tag
		var clientLinks, newClientLinks []map[string]string
		_ = json.Unmarshal(c.Links, &clientLinks)
		for _, cl := range clientLinks {
			if cl["remark"] != tag {
				newClientLinks = append(newClientLinks, cl)
			}
		}
		c.Links, _ = json.MarshalIndent(newClientLinks, "", "  ")
	}
	db.Set(cfg)
	return database.SaveConfig()
}

// UpdateLinksByInboundChangeJSON regenerates links for changed inbounds.
func (s *ClientService) UpdateLinksByInboundChangeJSON(inbound db.Inbound, oldTag string, hostname string) error {
	cfg := db.Get()
	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		var ids []uint
		if err := json.Unmarshal(c.Inbounds, &ids); err != nil {
			continue
		}
		hasInbound := false
		for _, id := range ids {
			if id == inbound.Id {
				hasInbound = true
				break
			}
		}
		if !hasInbound {
			continue
		}

		var clientLinks, newClientLinks []map[string]string
		_ = json.Unmarshal(c.Links, &clientLinks)

		inbForLink := model.Inbound{
			Id:      inbound.Id,
			Type:    inbound.Type,
			Tag:     inbound.Tag,
			TlsId:   inbound.TlsId,
			Options: inbound.Options,
		}
		newLinks := util.LinkGenerator(c.Config, &inbForLink, hostname)
		for _, newLink := range newLinks {
			newClientLinks = append(newClientLinks, map[string]string{
				"remark": inbound.Tag,
				"type":   "local",
				"uri":    newLink,
			})
		}
		for _, cl := range clientLinks {
			if cl["type"] != "local" || (cl["remark"] != inbound.Tag && cl["remark"] != oldTag) {
				newClientLinks = append(newClientLinks, cl)
			}
		}
		c.Links, _ = json.MarshalIndent(newClientLinks, "", "  ")
	}
	db.Set(cfg)
	return database.SaveConfig()
}

func (s *ClientService) DepleteClients() ([]uint, error) {
	dt := time.Now().Unix()
	inboundIds, err := s.ResetClientsJSON(dt)
	if err != nil {
		return nil, err
	}

	cfg := db.Get()
	var changes []db.Change
	var affectedClients []db.Client

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
			Obj:      json.RawMessage("\"" + c.Name + "\""),
		})
		affectedClients = append(affectedClients, *c)
	}

	if len(changes) > 0 {
		cfg.Changes = append(cfg.Changes, changes...)
		common.LastUpdate = dt
		db.Set(cfg)
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}
	}

	return inboundIds, nil
}

func (s *ClientService) ResetClientsJSON(dt int64) ([]uint, error) {
	var inboundIds []uint
	cfg := db.Get()
	var changes []db.Change

	// Delay start without periodic reset
	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if !c.Enable || !c.DelayStart || c.AutoReset {
			continue
		}
		if (c.Up + c.Down) > 0 {
			c.Expiry = dt + int64(c.ResetDays)*86400
			c.DelayStart = false
			changes = append(changes, db.Change{
				DateTime: dt,
				Actor:    "ResetJob",
				Key:      "clients",
				Action:   "reset",
				Obj:      json.RawMessage("\"" + c.Name + "\""),
			})
		}
	}

	// Delay start with periodic reset
	for i := range cfg.Clients {
		c := &cfg.Clients[i]
		if !c.Enable || !c.DelayStart || !c.AutoReset {
			continue
		}
		if (c.Up + c.Down) > 0 {
			c.NextReset = dt + int64(c.ResetDays)*86400
			c.DelayStart = false
			changes = append(changes, db.Change{
				DateTime: dt,
				Actor:    "ResetJob",
				Key:      "clients",
				Action:   "reset",
				Obj:      json.RawMessage("\"" + c.Name + "\""),
			})
		}
	}

	// Periodic reset
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
		if err := database.SaveConfig(); err != nil {
			return nil, err
		}
	}

	return inboundIds, nil
}

func (s *ClientService) findInboundsChangesJSON(client *model.Client, fillOmitted bool) ([]uint, error) {
	cfg := db.Get()
	var oldClient *db.Client
	for _, c := range cfg.Clients {
		if c.Id == client.Id {
			oldClient = &c
			break
		}
	}
	if oldClient == nil {
		return nil, nil
	}
	if fillOmitted {
		client.Links = oldClient.Links
		client.Config = oldClient.Config
	}

	var oldInboundIds, newInboundIds []uint
	_ = json.Unmarshal(oldClient.Inbounds, &oldInboundIds)
	_ = json.Unmarshal(client.Inbounds, &newInboundIds)

	// Check config or name changes
	if !jsonBytesEqual(oldClient.Config, client.Config) ||
		oldClient.Name != client.Name ||
		oldClient.Enable != client.Enable {
		return common.UnionUintArray(oldInboundIds, newInboundIds), nil
	}

	return common.DiffUintArray(oldInboundIds, newInboundIds), nil
}

func jsonBytesEqual(a, b json.RawMessage) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return string(a) == string(b)
}
