	package service

import (
	"github.com/pupmme/pupmsub/core"
	"github.com/pupmme/pupmsub/util/common"
	"github.com/pupmme/pupmsub/db"
	"encoding/json"

)

type OutboundService struct{}

func (o *OutboundService) GetAll() (*[]map[string]interface{}, error) {
	cfg := db.Get()
	var data []map[string]interface{}
	for _, outbound := range cfg.Outbounds {
		outData := map[string]interface{}{
			"id":   outbound.Id,
			"type": outbound.Type,
			"tag":  outbound.Tag,
		}
		if outbound.Options != nil {
			var restFields map[string]json.RawMessage
			if err := json.Unmarshal(outbound.Options, &restFields); err == nil {
				for k, v := range restFields {
					outData[k] = v
				}
			}
		}
		data = append(data, outData)
	}
	return &data, nil
}

// GetAllConfig returns all outbounds as sing-box JSON configs.
func (o *OutboundService) GetAllConfig() ([]json.RawMessage, error) {
	cfg := db.Get()
	var outboundsJson []json.RawMessage
	for _, outbound := range cfg.Outbounds {
		outModel := db.Outbound{
			Id:      outbound.Id,
			Type:    outbound.Type,
			Tag:     outbound.Tag,
			Options: outbound.Options,
		}
		outboundJson, err := json.Marshal(outModel)
		if err != nil {
			return nil, err
		}
		outboundsJson = append(outboundsJson, outboundJson)
	}
	return outboundsJson, nil
}

// Save handles CRUD for outbounds. tx is ignored in JSON mode.
func (s *OutboundService) Save(tx interface{}, act string, data json.RawMessage) error {
	cfg := db.Get()

	switch act {
	case "new", "edit":
		var outbound db.Outbound
		if err := json.Unmarshal(data, &outbound); err != nil {
			return err
		}

		if core.GetCore().IsRunning() {
			configData, err := json.Marshal(outbound)
			if err != nil {
				return err
			}
			if act == "edit" {
				var oldTag string
				for _, o := range cfg.Outbounds {
					if o.Id == outbound.Id {
						oldTag = o.Tag
						break
					}
				}
				if oldTag != "" {
					if err := core.GetCore().RemoveOutbound(oldTag); err != nil && err != nil && err.Error() != "invalid" {
						return err
					}
				}
			}
			if err := core.GetCore().AddOutbound(configData); err != nil {
				return err
			}
		}

		outJSON := db.Outbound{
			Id:      outbound.Id,
			Type:    outbound.Type,
			Tag:     outbound.Tag,
			Options: outbound.Options,
		}
		if act == "edit" {
			found := false
			for i := range cfg.Outbounds {
				if cfg.Outbounds[i].Id == outbound.Id {
					cfg.Outbounds[i] = outJSON
					found = true
					break
				}
			}
			if !found {
				cfg.Outbounds = append(cfg.Outbounds, outJSON)
			}
		} else {
			maxId := uint(0)
			for _, o := range cfg.Outbounds {
				if o.Id > maxId {
					maxId = o.Id
				}
			}
			outJSON.Id = maxId + 1
			cfg.Outbounds = append(cfg.Outbounds, outJSON)
		}
		db.Set(cfg)
		return db.SaveConfig()

	case "del":
		var tag string
		if err := json.Unmarshal(data, &tag); err != nil {
			return err
		}
		if core.GetCore().IsRunning() {
			if err := core.GetCore().RemoveOutbound(tag); err != nil && err != nil && err.Error() != "invalid" {
				return err
			}
		}
		newOutbounds := make([]db.Outbound, 0, len(cfg.Outbounds))
		for _, o := range cfg.Outbounds {
			if o.Tag != tag {
				newOutbounds = append(newOutbounds, o)
			}
		}
		cfg.Outbounds = newOutbounds
		db.Set(cfg)
		return db.SaveConfig()

	default:
		return common.NewErrorf("unknown action: %s", act)
	}
}
