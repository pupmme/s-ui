	package service

import (
	"github.com/pupmme/pupmmesub/config"
	"github.com/pupmme/pupmmesub/db"
	"github.com/pupmme/pupmmesub/logger"
	"github.com/pupmme/pupmmesub/util/common"
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

)

var defaultConfig = `{
  "log": { "level": "info" },
  "dns": { "servers": [], "rules": [] },
  "route": { "rules": [{ "action": "sniff" }, { "protocol": ["dns"], "action": "hijack-dns" }] },
  "experimental": {}
}`

var defaultValueMap = map[string]string{
	"webListen":     "",
	"webDomain":     "",
	"webPort":       "2053",
	"secret":        common.Random(32),
	"webCertFile":   "",
	"webKeyFile":    "",
	"webPath":       "/app/",
	"webURI":        "",
	"sessionMaxAge": "0",
	"trafficAge":    "30",
	"timeLocation":  "Asia/Shanghai",
	"subListen":     "",
	"subPort":       "",
	"subPath":       "",
	"subDomain":     "",
	"subCertFile":   "",
	"subKeyFile":    "",
	"subUpdates":    "",
	"subEncode":     "false",
	"subShowInfo":   "false",
	"subURI":        "",
	"subJsonExt":    "",
	"subClashExt":   "",
	"config":        defaultConfig,
	"version":       "sub",
	"nodeMode":      "false",
	"xboardApiHost": "",
	"xboardApiKey":  "",
}

type SettingService struct{}

func (s *SettingService) GetAllSetting() (*map[string]string, error) {
	cfg := db.Get()
	allSetting := make(map[string]string)

	for key, value := range cfg.Settings {
		allSetting[key] = value
	}

	// Batch-apply missing defaults, then persist once
	var missing []string
	for key, defaultValue := range defaultValueMap {
		if _, exists := allSetting[key]; !exists {
			if cfg.Settings == nil {
				cfg.Settings = make(map[string]string)
			}
			cfg.Settings[key] = defaultValue
			missing = append(missing, key)
			allSetting[key] = defaultValue
		}
	}
	if len(missing) > 0 {
		db.Set(cfg)
		if err := db.SaveConfig(); err != nil {
			logger.Warning("GetAllSetting: save defaults ", missing, ": ", err)
		}
	}

	delete(allSetting, "secret")
	delete(allSetting, "config")
	delete(allSetting, "version")

	return &allSetting, nil
}

func (s *SettingService) ResetSettings() error {
	cfg := db.Get()
	cfg.Settings = make(map[string]string)
	db.Set(cfg)
	return db.SaveConfig()
}

func (s *SettingService) getSetting(key string) (*settingRecord, error) {
	cfg := db.Get()
	if val, ok := cfg.Settings[key]; ok {
		return &settingRecord{Key: key, Value: val}, nil
	}
	return nil, common.NewErrorf("key <%v> not found", key)
}

type settingRecord struct {
	Key   string
	Value string
}

func (s *SettingService) getString(key string) (string, error) {
	setting, err := s.getSetting(key)
	if err != nil {
		value, ok := defaultValueMap[key]
		if !ok {
			return "", common.NewErrorf("key <%v> not in defaultValueMap", key)
		}
		return value, nil
	}
	return setting.Value, nil
}

func (s *SettingService) saveSetting(key string, value string) error {
	cfg := db.Get()
	if cfg.Settings == nil {
		cfg.Settings = make(map[string]string)
	}
	cfg.Settings[key] = value
	db.Set(cfg)
	return db.SaveConfig()
}

func (s *SettingService) setString(key string, value string) error {
	return s.saveSetting(key, value)
}

func (s *SettingService) getBool(key string) (bool, error) {
	str, err := s.getString(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(str)
}

func (s *SettingService) getInt(key string) (int, error) {
	str, err := s.getString(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (s *SettingService) setInt(key string, value int) error {
	return s.setString(key, strconv.Itoa(value))
}

func (s *SettingService) GetListen() (string, error) {
	return s.getString("webListen")
}

func (s *SettingService) GetWebDomain() (string, error) {
	return s.getString("webDomain")
}

func (s *SettingService) GetPort() (int, error) {
	return s.getInt("webPort")
}

func (s *SettingService) SetPort(port int) error {
	return s.setInt("webPort", port)
}

func (s *SettingService) GetCertFile() (string, error) {
	return s.getString("webCertFile")
}

func (s *SettingService) GetKeyFile() (string, error) {
	return s.getString("webKeyFile")
}

func (s *SettingService) GetWebPath() (string, error) {
	webPath, err := s.getString("webPath")
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(webPath, "/") {
		webPath = "/" + webPath
	}
	if !strings.HasSuffix(webPath, "/") {
		webPath += "/"
	}
	return webPath, nil
}

func (s *SettingService) SetWebPath(webPath string) error {
	if !strings.HasPrefix(webPath, "/") {
		webPath = "/" + webPath
	}
	if !strings.HasSuffix(webPath, "/") {
		webPath += "/"
	}
	return s.setString("webPath", webPath)
}

func (s *SettingService) GetSecret() ([]byte, error) {
	secret, err := s.getString("secret")
	if secret == defaultValueMap["secret"] {
		err := s.saveSetting("secret", secret)
		if err != nil {
			logger.Warning("save secret failed:", err)
		}
	}
	return []byte(secret), err
}

func (s *SettingService) GetSessionMaxAge() (int, error) {
	return s.getInt("sessionMaxAge")
}

func (s *SettingService) GetTrafficAge() (int, error) {
	return s.getInt("trafficAge")
}

// Node mode settings — read/write config.json
func (s *SettingService) GetNodeMode() (bool, error) {
	return config.Get().Node, nil
}

func (s *SettingService) SetNodeMode(enabled bool) error {
	cfg := config.Get()
	cfg.Node = enabled
	config.Set(cfg)
	return config.Save()
}

func (s *SettingService) GetXboardApiHost() (string, error) {
	return config.Get().Xboard.ApiHost, nil
}

func (s *SettingService) SetXboardApiHost(host string) error {
	cfg := config.Get()
	cfg.Xboard.ApiHost = host
	config.Set(cfg)
	return config.Save()
}

func (s *SettingService) GetXboardApiKey() (string, error) {
	return config.Get().Xboard.ApiKey, nil
}

func (s *SettingService) SetXboardApiKey(key string) error {
	cfg := config.Get()
	cfg.Xboard.ApiKey = key
	config.Set(cfg)
	return config.Save()
}

func (s *SettingService) GetNodeID() (int, error) {
	return config.Get().Xboard.NodeID, nil
}

func (s *SettingService) SetNodeID(id int) error {
	cfg := config.Get()
	cfg.Xboard.NodeID = id
	config.Set(cfg)
	return config.Save()
}

func (s *SettingService) GetNodeType() (string, error) {
	return config.Get().Xboard.NodeType, nil
}

func (s *SettingService) SetNodeType(nodeType string) error {
	cfg := config.Get()
	cfg.Xboard.NodeType = nodeType
	config.Set(cfg)
	return config.Save()
}

func (s *SettingService) GetTimeLocation() (*time.Location, error) {
	l, err := s.getString("timeLocation")
	if err != nil {
		return nil, err
	}
	if runtime.GOOS == "windows" {
		l = "Local"
	}
	location, err := time.LoadLocation(l)
	if err != nil {
		defaultLocation := defaultValueMap["timeLocation"]
		logger.Errorf("location <%v> not exist, using default location: %v", l, defaultLocation)
		return time.LoadLocation(defaultLocation)
	}
	return location, nil
}

// Subscription methods — no longer needed, stubs for compatibility
func (s *SettingService) GetSubListen() (string, error)           { return "", nil }
func (s *SettingService) GetSubPort() (int, error)                  { return 0, nil }
func (s *SettingService) SetSubPort(int) error                     { return nil }
func (s *SettingService) GetSubPath() (string, error)               { return "", nil }
func (s *SettingService) SetSubPath(string) error                   { return nil }
func (s *SettingService) GetSubDomain() (string, error)            { return "", nil }
func (s *SettingService) GetSubCertFile() (string, error)          { return "", nil }
func (s *SettingService) GetSubKeyFile() (string, error)           { return "", nil }
func (s *SettingService) GetSubUpdates() (int, error)              { return 0, nil }
func (s *SettingService) GetSubEncode() (bool, error)              { return false, nil }
func (s *SettingService) GetSubShowInfo() (bool, error)            { return false, nil }
func (s *SettingService) GetSubURI() (string, error)                { return "", nil }
func (s *SettingService) GetFinalSubURI(string) (string, error)    { return "", nil }
func (s *SettingService) GetSubJsonExt() (string, error)           { return "", nil }
func (s *SettingService) GetSubClashExt() (string, error)          { return "", nil }

func (s *SettingService) GetConfig() (string, error) {
	return s.getString("config")
}

func (s *SettingService) SetConfig(config string) error {
	return s.setString("config", config)
}

func (s *SettingService) SaveConfig(tx interface{}, config json.RawMessage) error {
	configs, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return s.saveSetting("config", string(configs))
}

func (s *SettingService) Save(tx interface{}, data json.RawMessage) error {
	var settings map[string]string
	err := json.Unmarshal(data, &settings)
	if err != nil {
		return err
	}
	for key, obj := range settings {
		if obj != "" && (key == "webCertFile" ||
			key == "webKeyFile" ||
			key == "subCertFile" ||
			key == "subKeyFile") {
			if _, err := os.Stat(obj); err != nil {
				return common.NewError(" -> ", obj, " is not exists")
			}
		}
		if key == "webPath" || key == "subPath" {
			if !strings.HasPrefix(obj, "/") {
				obj = "/" + obj
			}
			if !strings.HasSuffix(obj, "/") {
				obj += "/"
			}
		}
		err = s.saveSetting(key, obj)
		if err != nil {
			return err
		}
	}
	return nil
}
