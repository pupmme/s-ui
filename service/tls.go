package service

import (
	"encoding/json"

	"github.com/pupmme/sub/database"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/util/common"
)

type TlsService struct {
	InboundService
	ServicesService
}

func (s *TlsService) GetAll() ([]db.Tls, error) {
	cfg := db.Get()
	result := make([]db.Tls, 0, len(cfg.TLS))
	for _, tls := range cfg.TLS {
		result = append(result, db.Tls{
			Id:     tls.Id,
			Name:   tls.Name,
			Server: tls.Server,
			Client: tls.Client,
		})
	}
	return result, nil
}

// Save handles CRUD for TLS. tx is ignored in JSON mode.
func (s *TlsService) Save(tx interface{}, action string, data json.RawMessage, hostname string) error {
	cfg := db.Get()

	switch action {
	case "new", "edit":
		var tls db.Tls
		if err := json.Unmarshal(data, &tls); err != nil {
			return err
		}

		tlsJSON := db.TLS{
			Id:     tls.Id,
			Name:   tls.Name,
			Server: tls.Server,
			Client: tls.Client,
		}
		if action == "edit" {
			found := false
			for i := range cfg.TLS {
				if cfg.TLS[i].Id == tls.Id {
					cfg.TLS[i] = tlsJSON
					found = true
					break
				}
			}
			if !found {
				cfg.TLS = append(cfg.TLS, tlsJSON)
			}

			// Find inbounds using this TLS and restart them
			var inboundIds []uint
			for _, inb := range cfg.Inbounds {
				if inb.TlsId == tls.Id {
					inboundIds = append(inboundIds, inb.Id)
				}
			}
			if len(inboundIds) > 0 {
				err := s.InboundService.UpdateOutJsons(nil, inboundIds, hostname)
				if err != nil {
					return common.NewError("unable to update out_json of inbounds: ", err.Error())
				}
				err = s.InboundService.RestartInbounds(nil, inboundIds)
				if err != nil {
					return err
				}
			}

			// Find services using this TLS
			var serviceIds []uint
			for _, srv := range cfg.Services {
				if srv.TlsId == tls.Id {
					serviceIds = append(serviceIds, srv.Id)
				}
			}
			if len(serviceIds) > 0 {
				err := s.ServicesService.RestartServices(nil, serviceIds)
				if err != nil {
					return err
				}
			}

		} else {
			maxId := uint(0)
			for _, t := range cfg.TLS {
				if t.Id > maxId {
					maxId = t.Id
				}
			}
			tlsJSON.Id = maxId + 1
			cfg.TLS = append(cfg.TLS, tlsJSON)
		}
		db.Set(cfg)
		return database.SaveConfig()

	case "del":
		var id uint
		if err := json.Unmarshal(data, &id); err != nil {
			return err
		}
		inboundCount := 0
		for _, inb := range cfg.Inbounds {
			if inb.TlsId == id {
				inboundCount++
			}
		}
		serviceCount := 0
		for _, srv := range cfg.Services {
			if srv.TlsId == id {
				serviceCount++
			}
		}
		if inboundCount > 0 || serviceCount > 0 {
			return common.NewError("tls in use")
		}
		newTLS := make([]db.TLS, 0, len(cfg.TLS))
		for _, t := range cfg.TLS {
			if t.Id != id {
				newTLS = append(newTLS, t)
			}
		}
		cfg.TLS = newTLS
		db.Set(cfg)
		return database.SaveConfig()

	default:
		return common.NewErrorf("unknown action: %s", action)
	}
}
