	package service

import (
	"github.com/pupmme/pupmsub/core"
	"github.com/pupmme/pupmsub/util/common"
	"github.com/pupmme/pupmsub/db"
	"encoding/json"

)

type ServicesService struct{}

// GetAll returns all services as maps.
func (s *ServicesService) GetAll() (*[]map[string]interface{}, error) {
	cfg := db.Get()
	var data []map[string]interface{}
	for _, srv := range cfg.Services {
		srvData := map[string]interface{}{
			"id":     srv.Id,
			"type":   srv.Type,
			"tag":    srv.Tag,
			"tls_id": srv.TlsId,
		}
		if srv.Options != nil {
			var restFields map[string]json.RawMessage
			if err := json.Unmarshal(srv.Options, &restFields); err == nil {
				for k, v := range restFields {
					srvData[k] = v
				}
			}
		}
		data = append(data, srvData)
	}
	return &data, nil
}

// GetAllConfig returns all services as sing-box JSON configs.
func (s *ServicesService) GetAllConfig() ([]json.RawMessage, error) {
	cfg := db.Get()
	var servicesJson []json.RawMessage
	for _, srv := range cfg.Services {
		srvModel := db.Service{
			Id:      srv.Id,
			Type:    srv.Type,
			Tag:     srv.Tag,
			TlsId:   srv.TlsId,
			Options: srv.Options,
		}
		if srv.TlsId > 0 {
			for _, tls := range cfg.TLS {
				if tls.Id == srv.TlsId {
					srvModel.Tls = &db.TLS{
						Id:     tls.Id,
						Name:   tls.Name,
						Server: tls.Server,
						Client: tls.Client,
					}
					break
				}
			}
		}
		srvJson, err := json.Marshal(srvModel)
		if err != nil {
			return nil, err
		}
		servicesJson = append(servicesJson, srvJson)
	}
	return servicesJson, nil
}

// Save handles CRUD for services. tx is ignored in JSON mode.
func (s *ServicesService) Save(tx interface{}, act string, data json.RawMessage) error {
	cfg := db.Get()

	switch act {
	case "new", "edit":
		var srv db.Service
		if err := json.Unmarshal(data, &srv); err != nil {
			return err
		}

		if srv.TlsId > 0 {
			for i := range cfg.TLS {
				if cfg.TLS[i].Id == srv.TlsId {
					tls := cfg.TLS[i]
					srv.Tls = &db.TLS{
						Id:     tls.Id,
						Name:   tls.Name,
						Server: tls.Server,
						Client: tls.Client,
					}
					break
				}
			}
		}

		if core.GetCore().IsRunning() {
			configData, err := json.Marshal(srv)
			if err != nil {
				return err
			}
			if act == "edit" {
				var oldTag string
				for _, s := range cfg.Services {
					if s.Id == srv.Id {
						oldTag = s.Tag
						break
					}
				}
				if oldTag != "" {
					if err := core.GetCore().RemoveService(oldTag); err != nil && err != nil && err.Error() != "not found" {
						return err
					}
				}
			}
			if err := core.GetCore().AddService(configData); err != nil {
				return err
			}
		}

		srvJSON := db.Service{
			Id:      srv.Id,
			Type:    srv.Type,
			Tag:     srv.Tag,
			TlsId:   srv.TlsId,
			Options: srv.Options,
		}
		if act == "edit" {
			found := false
			for i := range cfg.Services {
				if cfg.Services[i].Id == srv.Id {
					cfg.Services[i] = srvJSON
					found = true
					break
				}
			}
			if !found {
				cfg.Services = append(cfg.Services, srvJSON)
			}
		} else {
			maxId := uint(0)
			for _, s := range cfg.Services {
				if s.Id > maxId {
					maxId = s.Id
				}
			}
			srvJSON.Id = maxId + 1
			cfg.Services = append(cfg.Services, srvJSON)
		}
		db.Set(cfg)
		return db.SaveConfig()

	case "del":
		var tag string
		if err := json.Unmarshal(data, &tag); err != nil {
			return err
		}
		if core.GetCore().IsRunning() {
			if err := core.GetCore().RemoveService(tag); err != nil && err != nil && err.Error() != "not found" {
				return err
			}
		}
		newServices := make([]db.Service, 0, len(cfg.Services))
		for _, s := range cfg.Services {
			if s.Tag != tag {
				newServices = append(newServices, s)
			}
		}
		cfg.Services = newServices
		db.Set(cfg)
		return db.SaveConfig()

	default:
		return common.NewErrorf("unknown action: %s", act)
	}
}

// RestartServices restarts specific services by IDs.
func (s *ServicesService) RestartServices(tx interface{}, ids []uint) error {
	if !core.GetCore().IsRunning() {
		return nil
	}
	cfg := db.Get()
	for _, id := range ids {
		for _, srv := range cfg.Services {
			if srv.Id != id {
				continue
			}
			if err := core.GetCore().RemoveService(srv.Tag); err != nil && err != nil && err.Error() != "not found" {
				return err
			}
			srvModel := db.Service{
				Id:      srv.Id,
				Type:    srv.Type,
				Tag:     srv.Tag,
				TlsId:   srv.TlsId,
				Options: srv.Options,
			}
			if srv.TlsId > 0 {
				for _, tls := range cfg.TLS {
					if tls.Id == srv.TlsId {
						srvModel.Tls = &db.TLS{Id: tls.Id, Name: tls.Name, Server: tls.Server, Client: tls.Client}
						break
					}
				}
			}
			srvConfig, err := json.Marshal(srvModel)
			if err != nil {
				return err
			}
			if err := core.GetCore().AddService(srvConfig); err != nil {
				return err
			}
		}
	}
	return nil
}
