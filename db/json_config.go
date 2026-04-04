package db

import "encoding/json"

// Config is the root JSON configuration structure.
// All data that was previously stored in SQLite is here.
type Config struct {
	Version int `json:"version"`

	// Users holds panel admin accounts (username/password login)
	Users []User `json:"users"`

	// Settings holds key-value panel settings
	Settings map[string]string `json:"settings"`

	// TLS holds TLS certificate definitions
	TLS []TLS `json:"tls"`

	// Inbounds holds sing-box inbound configurations
	Inbounds []Inbound `json:"inbounds"`

	// Outbounds holds sing-box outbound configurations
	Outbounds []Outbound `json:"outbounds"`

	// Services holds sing-box service configurations
	Services []Service `json:"services"`

	// Endpoints holds warp/endpoint configurations
	Endpoints []Endpoint `json:"endpoints"`

	// Clients holds subscription client records
	Clients []Client `json:"clients"`

	// Stats holds traffic statistics records
	Stats []Stat `json:"stats"`

	// Changes holds audit log entries
	Changes []Change `json:"changes"`
}

// User mirrors database/model/User for panel admin login.
type User struct {
	Id         uint   `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	LastLogins string `json:"lastLogins"`
}

// TLS mirrors database/model/Tls.
type TLS struct {
	Id     uint            `json:"id"`
	Name   string          `json:"name"`
	Server json.RawMessage `json:"server"`
	Client json.RawMessage `json:"client"`
}

// Inbound mirrors database/model/Inbound.
type Inbound struct {
	Id     uint            `json:"id"`
	Type   string          `json:"type"`
	Tag    string          `json:"tag"`
	TlsId  uint            `json:"tls_id"`
	Tls    *TLS            `json:"tls,omitempty"`
	Addrs  json.RawMessage `json:"addrs"`
	OutJson json.RawMessage `json:"out_json"`
	Options json.RawMessage `json:"options"`
}

// Outbound mirrors database/model/Outbound.
type Outbound struct {
	Id      uint            `json:"id"`
	Type    string          `json:"type"`
	Tag     string          `json:"tag"`
	Options json.RawMessage `json:"options"`
}

// Service mirrors database/model/Service.
type Service struct {
	Id      uint            `json:"id"`
	Type    string          `json:"type"`
	Tag     string          `json:"tag"`
	TlsId   uint            `json:"tls_id"`
	Tls     *TLS            `json:"tls,omitempty"`
	Options json.RawMessage `json:"options"`
}

// Endpoint mirrors database/model/Endpoint.
type Endpoint struct {
	Id      uint            `json:"id"`
	Type    string          `json:"type"`
	Tag     string          `json:"tag"`
	Options json.RawMessage `json:"options"`
	Ext     json.RawMessage `json:"ext"`
}

// Client mirrors database/model/Client.
type Client struct {
	Id        uint            `json:"id"`
	Enable    bool            `json:"enable"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	Inbounds  json.RawMessage `json:"inbounds"`
	Links     json.RawMessage `json:"links"`
	Volume    int64           `json:"volume"`
	Expiry    int64           `json:"expiry"`
	Down      int64           `json:"down"`
	Up        int64           `json:"up"`
	Desc      string          `json:"desc"`
	Group     string          `json:"group"`
	DelayStart bool           `json:"delayStart"`
	AutoReset  bool           `json:"autoReset"`
	ResetDays  int            `json:"resetDays"`
	NextReset  int64          `json:"nextReset"`
	TotalUp    int64          `json:"totalUp"`
	TotalDown  int64          `json:"totalDown"`
}

// Stat mirrors database/model/Stats.
type Stat struct {
	Id        uint64 `json:"id"`
	DateTime  int64  `json:"dateTime"`
	Resource  string `json:"resource"`
	Tag       string `json:"tag"`
	Direction bool   `json:"direction"`
	Traffic   int64  `json:"traffic"`
}

// Change mirrors database/model/Changes.
type Change struct {
	Id       uint64          `json:"id"`
	DateTime int64           `json:"dateTime"`
	Actor    string          `json:"actor"`
	Key      string          `json:"key"`
	Action   string          `json:"action"`
	Obj      json.RawMessage `json:"obj"`
}

// Tokens mirrors database/model/Tokens (kept for signature compatibility).
type Tokens struct {
	Id     uint   `json:"id"`
	Desc   string `json:"desc"`
	Token  string `json:"token"`
	Expiry int64  `json:"expiry"`
	UserId uint   `json:"userId"`
}
