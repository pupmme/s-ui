package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var (
	cfg *Config
	mu  sync.RWMutex
)

// Load reads the JSON config file from disk.
// If the file does not exist, a new empty Config is initialized.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &Config{Version: 1, Settings: make(map[string]string)}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &cfg)
}

var dbPath = "/etc/sub/singbox.json"

// SaveConfig persists the JSON config to disk (no args = use default path).
func SaveConfig() error {
	return Save(dbPath)
}

// Save writes the current Config to disk as indented JSON.
func Save(path string) error {
	mu.RLock()
	defer mu.RUnlock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Get returns the current Config (thread-safe read).
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}

// Set replaces the current Config (thread-safe write).
func Set(c *Config) {
	mu.Lock()
	defer mu.Unlock()
	cfg = c
}

// GetSettings returns the settings map (convenience accessor).
func GetSettings() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	if cfg.Settings == nil {
		return make(map[string]string)
	}
	return cfg.Settings
}

// GetUsers returns the users slice.
func GetUsers() []User {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Users
}

// GetInbounds returns the inbounds slice.
func GetInbounds() []Inbound {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Inbounds
}

// GetOutbounds returns the outbounds slice.
func GetOutbounds() []Outbound {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Outbounds
}

// GetServices returns the services slice.
func GetServices() []Service {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Services
}

// GetEndpoints returns the endpoints slice.
func GetEndpoints() []Endpoint {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Endpoints
}

// GetClients returns the clients slice.
func GetClients() []Client {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Clients
}

// GetStats returns the stats slice.
func GetStats() []Stat {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Stats
}

// GetChanges returns the changes slice.
func GetChanges() []Change {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Changes
}

// GetTLS returns the tls slice.
func GetTLS() []TLS {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.TLS
}
