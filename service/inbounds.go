	package service

import (
	"github.com/pupmme/pupmmesub/core"
	"github.com/pupmme/pupmmesub/util"
	"github.com/pupmme/pupmmesub/util/common"
	"github.com/pupmme/pupmmesub/db"
	"encoding/json"
	"strings"

)

type InboundService struct {
	ClientService
}

func (s *InboundService) Get(ids string) (*[]map[string]interface{}, error) {
	if ids == "" {
		return s.GetAll()
	}
	return s.getById(ids)
}

func (s *InboundService) getById(ids string) (*[]map[string]interface{}, error) {
	cfg := db.Get()
	idMap := make(map[string]bool)
	for _, id := range strings.Split(ids, ",") {
		idMap[id] = true
	}
	var result []map[string]interface{}
	for _, inb := range cfg.Inbounds {
		if !idMap[common.Itoa(int(inb.Id))] {
			continue
		}
		data := s.inboundToMap(inb)
		if data == nil {
			continue
		}
		result = append(result, *data)
	}
	return &result, nil
}

func (s *InboundService) GetAll() (*[]map[string]interface{}, error) {
	cfg := db.Get()
	var data []map[string]interface{}
	for _, inbound := range cfg.Inbounds {
		m := s.inboundToMap(inbound)
		if m == nil {
			continue
		}
		data = append(data, *m)
	}
	return &data, nil
}

func (s *InboundService) inboundToMap(inb db.Inbound) *map[string]interface{} {
	var shadowtls_version uint
	ss_managed := false
	inbData := map[string]interface{}{
		"id":     inb.Id,
		"type":   inb.Type,
		"tag":    inb.Tag,
		"tls_id": inb.TlsId,
	}
	if inb.Options != nil {
		var restFields map[string]json.RawMessage
		if err := json.Unmarshal(inb.Options, &restFields); err == nil {
			inbData["listen"] = restFields["listen"]
			inbData["listen_port"] = restFields["listen_port"]
			if inb.Type == "shadowtls" {
				json.Unmarshal(restFields["version"], &shadowtls_version)
			}
			if inb.Type == "shadowsocks" {
				json.Unmarshal(restFields["managed"], &ss_managed)
			}
		}
	}
	if s.hasUser(inb.Type) &&
		!(inb.Type == "shadowtls" && shadowtls_version < 3) &&
		!(inb.Type == "shadowsocks" && ss_managed) {
		users := s.getClientNamesByInbound(inb.Id)
		inbData["users"] = users
	}
	return &inbData
}

func (s *InboundService) getClientNamesByInbound(inboundId uint) []string {
	cfg := db.Get()
	var names []string
	for _, client := range cfg.Clients {
		var ids []uint
		if err := json.Unmarshal(client.Inbounds, &ids); err != nil {
			continue
		}
		for _, id := range ids {
			if id == inboundId {
				names = append(names, client.Name)
				break
			}
		}
	}
	return names
}

func (s *InboundService) FromIds(ids []uint) ([]*db.Inbound, error) {
	cfg := db.Get()
	var result []*db.Inbound
	for _, inb := range cfg.Inbounds {
		for _, id := range ids {
			if inb.Id == id {
				result = append(result, &db.Inbound{
					Id:      inb.Id,
					Type:    inb.Type,
					Tag:     inb.Tag,
					TlsId:   inb.TlsId,
					Addrs:   inb.Addrs,
					OutJson: inb.OutJson,
					Options: inb.Options,
				})
				break
			}
		}
	}
	return result, nil
}

// Save handles CRUD for inbounds. tx parameter is ignored in JSON mode.
func (s *InboundService) Save(tx interface{}, act string, data json.RawMessage, initUserIds string, hostname string) error {
	if s.ClientService.IsNodeMode() {
		return common.NewError("inbounds are read-only in node mode")
	}
	cfg := db.Get()

	switch act {
	case "new", "edit":
		var inbound db.Inbound
		if err := json.Unmarshal(data, &inbound); err != nil {
			return err
		}

		if inbound.TlsId > 0 {
			for i := range cfg.TLS {
				if cfg.TLS[i].Id == inbound.TlsId {
					tls := cfg.TLS[i]
					inbound.Tls = &db.TLS{
						Id:     tls.Id,
						Name:   tls.Name,
						Server: tls.Server,
						Client: tls.Client,
					}
					break
				}
			}
		}

		var oldTag string
		if act == "edit" {
			for _, existing := range cfg.Inbounds {
				if existing.Id == inbound.Id {
					oldTag = existing.Tag
					break
				}
			}
		}

		if core.GetCore().IsRunning() {
			if act == "edit" && oldTag != "" {
				if err := core.GetCore().RemoveInbound(oldTag); err != nil && err != nil && err.Error() != "invalid" {
					return err
				}
			}
			inboundConfig, err := json.Marshal(inbound)
			if err != nil {
				return err
			}
			if act == "edit" {
				inboundConfig, err = s.addUsersJSON(inboundConfig, inbound.Id, inbound.Type)
			} else {
				inboundConfig, err = s.initUsersJSON(inboundConfig, initUserIds, inbound.Type)
			}
			if err != nil {
				return err
			}
			if err := core.GetCore().AddInbound(inboundConfig); err != nil {
				return err
			}
		}

		if err := util.FillOutJson(&inbound, hostname); err != nil {
			return err
		}

		inbJSON := db.Inbound{
			Id:      inbound.Id,
			Type:    inbound.Type,
			Tag:     inbound.Tag,
			TlsId:   inbound.TlsId,
			Addrs:   inbound.Addrs,
			OutJson: inbound.OutJson,
			Options: inbound.Options,
		}
		if act == "edit" {
			found := false
			for i := range cfg.Inbounds {
				if cfg.Inbounds[i].Id == inbound.Id {
					cfg.Inbounds[i] = inbJSON
					found = true
					break
				}
			}
			if !found {
				cfg.Inbounds = append(cfg.Inbounds, inbJSON)
			}
		} else {
			maxId := uint(0)
			for _, ib := range cfg.Inbounds {
				if ib.Id > maxId {
					maxId = ib.Id
				}
			}
			inbJSON.Id = maxId + 1
			cfg.Inbounds = append(cfg.Inbounds, inbJSON)
		}
		db.Set(cfg)
		if err := db.SaveConfig(); err != nil {
			return err
		}

		if act == "new" {
			err := s.ClientService.UpdateClientsOnInboundAdd(initUserIds, inbJSON.Id, hostname)
			if err != nil {
				return err
			}
		} else {
			err := s.ClientService.UpdateLinksByInboundChange(inbJSON, oldTag, hostname)
			if err != nil {
				return err
			}
		}

	case "del":
		var tag string
		if err := json.Unmarshal(data, &tag); err != nil {
			return err
		}
		if core.GetCore().IsRunning() {
			if err := core.GetCore().RemoveInbound(tag); err != nil && err != nil && err.Error() != "invalid" {
				return err
			}
		}
		var id uint
		for _, inb := range cfg.Inbounds {
			if inb.Tag == tag {
				id = inb.Id
				break
			}
		}
		if err := s.ClientService.UpdateClientsOnInboundDelete(id, tag); err != nil {
			return err
		}
		newInbounds := make([]db.Inbound, 0, len(cfg.Inbounds))
		for _, inb := range cfg.Inbounds {
			if inb.Tag != tag {
				newInbounds = append(newInbounds, inb)
			}
		}
		cfg.Inbounds = newInbounds
		db.Set(cfg)
		if err := db.SaveConfig(); err != nil {
			return err
		}
	default:
		return common.NewErrorf("unknown action: %s", act)
	}
	return nil
}

func (s *InboundService) UpdateOutJsons(tx interface{}, inboundIds []uint, hostname string) error {
	cfg := db.Get()
	for _, inboundId := range inboundIds {
		for i := range cfg.Inbounds {
			if cfg.Inbounds[i].Id != inboundId {
				continue
			}
			inb := &cfg.Inbounds[i]
			inbModel := &db.Inbound{
				Id:      inb.Id,
				Type:    inb.Type,
				Tag:     inb.Tag,
				TlsId:   inb.TlsId,
				Addrs:   inb.Addrs,
				OutJson: inb.OutJson,
				Options: inb.Options,
			}
			if inb.TlsId > 0 {
				for _, tls := range cfg.TLS {
					if tls.Id == inb.TlsId {
						inbModel.Tls = &db.TLS{
							Id:     tls.Id,
							Name:   tls.Name,
							Server: tls.Server,
							Client: tls.Client,
						}
						break
					}
				}
			}
			if err := util.FillOutJson(inbModel, hostname); err != nil {
				return err
			}
			inb.OutJson = inbModel.OutJson
		}
	}
	db.Set(cfg)
	return db.SaveConfig()
}

// GetAllConfig returns all inbounds as sing-box JSON configs.
func (s *InboundService) GetAllConfig() ([]json.RawMessage, error) {
	cfg := db.Get()
	var inboundsJson []json.RawMessage
	for _, inp := range cfg.Inbounds {
		inbModel := db.Inbound{
			Id:      inp.Id,
			Type:    inp.Type,
			Tag:     inp.Tag,
			TlsId:   inp.TlsId,
			Addrs:   inp.Addrs,
			OutJson: inp.OutJson,
			Options: inp.Options,
		}
		if inp.TlsId > 0 {
			for _, tls := range cfg.TLS {
				if tls.Id == inp.TlsId {
					inbModel.Tls = &db.TLS{
						Id:     tls.Id,
						Name:   tls.Name,
						Server: tls.Server,
						Client: tls.Client,
					}
					break
				}
			}
		}
		inboundJson, err := json.Marshal(inbModel)
		if err != nil {
			return nil, err
		}
		inboundJson, err = s.addUsersJSON(inboundJson, inp.Id, inp.Type)
		if err != nil {
			return nil, err
		}
		inboundsJson = append(inboundsJson, inboundJson)
	}
	return inboundsJson, nil
}

func (s *InboundService) hasUser(inboundType string) bool {
	switch inboundType {
	case "mixed", "socks", "http", "shadowsocks", "vmess", "trojan", "naive", "hysteria", "shadowtls", "tuic", "hysteria2", "vless", "anytls":
		return true
	}
	return false
}

func (s *InboundService) fetchUsersJSON(inboundId uint, inboundType string, inbound map[string]interface{}) ([]json.RawMessage, error) {
	if inboundType == "shadowtls" {
		version, _ := inbound["version"].(float64)
		if int(version) < 3 {
			return nil, nil
		}
	}
	if inboundType == "shadowsocks" {
		method, _ := inbound["method"].(string)
		if method == "2022-blake3-aes-128-gcm" {
			inboundType = "shadowsocks16"
		}
	}
	var usersJson []json.RawMessage
	cfg := db.Get()
	for _, client := range cfg.Clients {
		if !client.Enable {
			continue
		}
		var ids []uint
		if err := json.Unmarshal(client.Inbounds, &ids); err != nil {
			continue
		}
		hasInbound := false
		for _, id := range ids {
			if id == inboundId {
				hasInbound = true
				break
			}
		}
		if !hasInbound {
			continue
		}
		// client.Config is a JSON object keyed by inbound type
		if client.Config == nil {
			continue
		}
		var cfgMap map[string]json.RawMessage
		if err := json.Unmarshal(client.Config, &cfgMap); err != nil {
			continue
		}
		cfgRaw, ok := cfgMap[inboundType]
		if !ok {
			continue
		}
		cfgBytes := []byte(cfgRaw)
		if inboundType == "vless" && inbound["tls"] == nil {
			cfgStr := strings.Replace(string(cfgBytes), "xtls-rprx-vision", "", -1)
			cfgBytes = []byte(cfgStr)
		}
		usersJson = append(usersJson, json.RawMessage(cfgBytes))
	}
	return usersJson, nil
}

func (s *InboundService) addUsersJSON(inboundJson []byte, inboundId uint, inboundType string) ([]byte, error) {
	if !s.hasUser(inboundType) {
		return inboundJson, nil
	}
	var inbound map[string]interface{}
	if err := json.Unmarshal(inboundJson, &inbound); err != nil {
		return nil, err
	}
	users, err := s.fetchUsersJSON(inboundId, inboundType, inbound)
	if err != nil {
		return nil, err
	}
	inbound["users"] = users
	return json.Marshal(inbound)
}

func (s *InboundService) initUsersJSON(inboundJson []byte, clientIds string, inboundType string) ([]byte, error) {
	if !s.hasUser(inboundType) {
		return inboundJson, nil
	}
	var inbound map[string]interface{}
	if err := json.Unmarshal(inboundJson, &inbound); err != nil {
		return nil, err
	}
	clientIdList := strings.Split(clientIds, ",")
	clientIdSet := make(map[string]bool)
	for _, c := range clientIdList {
		clientIdSet[c] = true
	}
	cfg := db.Get()
	var usersJson []json.RawMessage
	for _, client := range cfg.Clients {
		if !client.Enable {
			continue
		}
		if !clientIdSet[common.Itoa(int(client.Id))] {
			continue
		}
		if client.Config == nil {
			continue
		}
		var cfgMap map[string]json.RawMessage
		if err := json.Unmarshal(client.Config, &cfgMap); err != nil {
			continue
		}
		cfgRaw, ok := cfgMap[inboundType]
		if !ok {
			continue
		}
		usersJson = append(usersJson, cfgRaw)
	}
	inbound["users"] = usersJson
	return json.Marshal(inbound)
}

func (s *InboundService) RestartInbounds(tx interface{}, ids []uint) error {
	if !core.GetCore().IsRunning() {
		return nil
	}
	for _, id := range ids {
		for _, inp := range db.Get().Inbounds {
			if inp.Id != id {
				continue
			}
			if err := core.GetCore().RemoveInbound(inp.Tag); err != nil && err != nil && err.Error() != "invalid" {
				return err
			}
			core.GetCore().GetInstance().ConnTracker().CloseConnByInbound(inp.Tag)

			inbModel := &db.Inbound{
				Id:      inp.Id,
				Type:    inp.Type,
				Tag:     inp.Tag,
				TlsId:   inp.TlsId,
				Addrs:   inp.Addrs,
				OutJson: inp.OutJson,
				Options: inp.Options,
			}
			inboundConfig, err := json.Marshal(inbModel)
			if err != nil {
				return err
			}
			inboundConfig, err = s.addUsersJSON(inboundConfig, inp.Id, inp.Type)
			if err != nil {
				return err
			}
			if err := core.GetCore().AddInbound(inboundConfig); err != nil {
				return err
			}
		}
	}
	return nil
}
