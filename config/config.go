package config

import (
	"encoding/json"
	"os"
	"sync"

)

var (
	cfg     *SubConfig
	cfgMu   sync.RWMutex
	cfgPath = "/etc/sub/config.json"
)

type SubConfig struct {
	Log  LogConfig  `json:"log"`
	Web  WebConfig  `json:"web"`
	Node bool       `json:"node"`
	Xboard XboardConfig `json:"xboard"`
}

type LogConfig struct {
	Level  string `json:"level"`
	Output string `json:"output"`
}

type WebConfig struct {
	Port     int    `json:"port"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type XboardConfig struct {
	ApiHost                 string         `json:"apiHost"`
	ApiKey                  string         `json:"apiKey"`
	NodeID                  int            `json:"nodeId"`
	NodeType                string         `json:"nodeType"`
	Timeout                 int            `json:"timeout"`
	ListenIP                string         `json:"listenIP"`
	SendIP                  string         `json:"sendIP"`
	DeviceOnlineMinTraffic   int            `json:"deviceOnlineMinTraffic"`
	MinReportTraffic        int            `json:"minReportTraffic"`
	TCPFastOpen             bool           `json:"tcpFastOpen"`
	SniffEnabled            bool           `json:"sniffEnabled"`
	InboundConfig           InboundConfig  `json:"inboundConfig"`
	CertConfig              CertConfig     `json:"certConfig"`
}

type InboundConfig struct {
	ProtocolOptions ProtocolOptions `json:"protocol_options"`
}

type ProtocolOptions struct {
	HeartbeatInterval string `json:"heartbeat_interval"`
	IdleTimeout       string `json:"idle_timeout"`
}

type CertConfig struct {
	CertMode         string            `json:"certMode"`
	RejectUnknownSni bool             `json:"rejectUnknownSni"`
	CertDomain       string            `json:"certDomain"`
	CertFile         string            `json:"certFile"`
	KeyFile          string            `json:"keyFile"`
	Provider         string            `json:"provider"`
	DNSEnv           map[string]string `json:"dnsEnv"`
}

func SetPath(path string) {
	cfgPath = path
}

func Load() error {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &SubConfig{
				Log: LogConfig{Level: "info"},
				Web: WebConfig{Port: 2053, Username: "admin", Password: "password"},
				Node: false,
				Xboard: XboardConfig{
					ListenIP: "0.0.0.0",
					SendIP:   "0.0.0.0",
					CertConfig: CertConfig{
						CertFile: "/etc/sub/cert.crt",
						KeyFile:  "/etc/sub/cert.key",
						CertMode: "dns",
						Provider: "cloudflare",
						DNSEnv:   make(map[string]string),
					},
					InboundConfig: InboundConfig{
						ProtocolOptions: ProtocolOptions{
							HeartbeatInterval: "3s",
							IdleTimeout:       "300s",
						},
					},
				},
			}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &cfg)
}

func Save() error {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0644)
}

func Get() *SubConfig {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return cfg
}

func Set(c *SubConfig) {
	cfgMu.Lock()
	cfg = c
	cfgMu.Unlock()
}

func GetWebPort() int {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if cfg == nil || cfg.Web.Port == 0 {
		return 2053
	}
	return cfg.Web.Port
}

func GetWebCert() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if cfg == nil {
		return ""
	}
	return cfg.Web.Cert
}

func GetWebKey() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if cfg == nil {
		return ""
	}
	return cfg.Web.Key
}

func GetWebUsername() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if cfg == nil {
		return "admin"
	}
	return cfg.Web.Username
}

func GetWebPassword() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if cfg == nil {
		return "password"
	}
	return cfg.Web.Password
}

func GetVersion() string {
	return "1.0.0"
}

func IsDebug() bool {
	return false
}
