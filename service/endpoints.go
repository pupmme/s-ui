	package service

import (
	"github.com/pupmme/sub/util/common"
	"github.com/pupmme/sub/db"
	"encoding/json"

)

type EndpointService struct {
	WarpService
}

func (o *EndpointService) GetAll() (*[]map[string]interface{}, error) {
	cfg := db.Get()
	var data []map[string]interface{}
	for _, endpoint := range cfg.Endpoints {
		epData := map[string]interface{}{
			"id":   endpoint.Id,
			"type": endpoint.Type,
			"tag":  endpoint.Tag,
			"ext":  endpoint.Ext,
		}
		if endpoint.Options != nil {
			var restFields map[string]json.RawMessage
			if err := json.Unmarshal(endpoint.Options, &restFields); err == nil {
				for k, v := range restFields {
					epData[k] = v
				}
			}
		}
		data = append(data, epData)
	}
	return &data, nil
}

// GetAllConfig returns all endpoints as sing-box JSON configs.
func (o *EndpointService) GetAllConfig() ([]json.RawMessage, error) {
	cfg := db.Get()
	var endpointsJson []json.RawMessage
	for _, endpoint := range cfg.Endpoints {
		epModel := db.Endpoint{
			Id:      endpoint.Id,
			Type:    endpoint.Type,
			Tag:     endpoint.Tag,
			Options: endpoint.Options,
			Ext:     endpoint.Ext,
		}
		endpointJson, err := json.Marshal(epModel)
		if err != nil {
			return nil, err
		}
		endpointsJson = append(endpointsJson, endpointJson)
	}
	return endpointsJson, nil
}

// Save handles CRUD for endpoints. tx is ignored in JSON mode.
func (s *EndpointService) Save(tx interface{}, act string, data json.RawMessage) error {
	cfg := db.Get()

	switch act {
	case "new", "edit":
		var endpoint db.Endpoint
		if err := json.Unmarshal(data, &endpoint); err != nil {
			return err
		}

		if endpoint.Type == "warp" {
			if act == "new" {
				if err := s.WarpService.RegisterWarp(&endpoint); err != nil {
					return err
				}
			} else {
				// Retrieve old license from existing endpoint
				var oldLicense string
				for _, ep := range cfg.Endpoints {
					if ep.Id == endpoint.Id {
						var extMap map[string]string
						if ep.Ext != nil {
							json.Unmarshal(ep.Ext, &extMap)
							oldLicense = extMap["license_key"]
						}
						break
					}
				}
				if err := s.WarpService.SetWarpLicense(oldLicense, &endpoint); err != nil {
					return err
				}
			}
		}

		if corePtr.IsRunning() {
			configData, err := json.Marshal(endpoint)
			if err != nil {
				return err
			}
			if act == "edit" {
				var oldTag string
				for _, ep := range cfg.Endpoints {
					if ep.Id == endpoint.Id {
						oldTag = ep.Tag
						break
					}
				}
				if oldTag != "" {
					if err := corePtr.RemoveEndpoint(oldTag); err != nil && err != nil && err.Error() != "not found" {
						return err
					}
				}
			}
			if err := corePtr.AddEndpoint(configData); err != nil {
				return err
			}
		}

		epJSON := db.Endpoint{
			Id:      endpoint.Id,
			Type:    endpoint.Type,
			Tag:     endpoint.Tag,
			Options: endpoint.Options,
			Ext:     endpoint.Ext,
		}
		if act == "edit" {
			found := false
			for i := range cfg.Endpoints {
				if cfg.Endpoints[i].Id == endpoint.Id {
					cfg.Endpoints[i] = epJSON
					found = true
					break
				}
			}
			if !found {
				cfg.Endpoints = append(cfg.Endpoints, epJSON)
			}
		} else {
			maxId := uint(0)
			for _, ep := range cfg.Endpoints {
				if ep.Id > maxId {
					maxId = ep.Id
				}
			}
			epJSON.Id = maxId + 1
			cfg.Endpoints = append(cfg.Endpoints, epJSON)
		}
		db.Set(cfg)
		return db.SaveConfig()

	case "del":
		var tag string
		if err := json.Unmarshal(data, &tag); err != nil {
			return err
		}
		if corePtr.IsRunning() {
			if err := corePtr.RemoveEndpoint(tag); err != nil && err != nil && err.Error() != "not found" {
				return err
			}
		}
		newEndpoints := make([]db.Endpoint, 0, len(cfg.Endpoints))
		for _, ep := range cfg.Endpoints {
			if ep.Tag != tag {
				newEndpoints = append(newEndpoints, ep)
			}
		}
		cfg.Endpoints = newEndpoints
		db.Set(cfg)
		return db.SaveConfig()

	default:
		return common.NewErrorf("unknown action: %s", act)
	}
}
