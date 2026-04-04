package database

import (
	"encoding/json"
	"os"
	"path"
	"sync"

	"github.com/alireza0/s-ui/config"
	"github.com/alireza0/s-ui/db"
	"github.com/alireza0/s-ui/logger"
)

// cfgPath is the path to the JSON config file.
var cfgPath string

// initMu protects first-time initialization.
var initMu sync.Mutex

// cfgMu protects write operations.
var cfgMu sync.Mutex

// InitDB loads (or creates) the JSON config file and seeds defaults.
// This replaces the previous SQLite + GORM InitDB.
func InitDB(dbPath string) error {
	initMu.Lock()
	defer initMu.Unlock()

	cfgPath = dbPath

	// Ensure parent directory exists
	dir := path.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	err := db.Load(dbPath)
	if err != nil {
		return err
	}

	cfg := db.Get()
	if cfg == nil {
		cfg = &db.Config{Version: 1, Settings: make(map[string]string)}
		db.Set(cfg)
	}

	// Seed default admin user if no users exist
	if len(cfg.Users) == 0 {
		cfg.Users = append(cfg.Users, db.User{
			Id:       1,
			Username: "admin",
			Password: "admin",
		})
		db.Set(cfg)
		if err := db.Save(dbPath); err != nil {
			logger.Warning("failed to save default admin user:", err)
		}
	}

	// Seed default outbounds if none exist
	if len(cfg.Outbounds) == 0 {
		cfg.Outbounds = append(cfg.Outbounds, db.Outbound{
			Id:      1,
			Type:    "direct",
			Tag:     "direct",
			Options: json.RawMessage(`{}`),
		})
		db.Set(cfg)
		if err := db.Save(dbPath); err != nil {
			logger.Warning("failed to save default outbound:", err)
		}
	}

	return nil
}

// OpenDB exists for backward compatibility (previous SQLite path).
// No-op now that we use JSON.
func OpenDB(dbPath string) error {
	return nil
}

// GetDB returns nil. All data access goes through the db package.
func GetDB() interface{} {
	return nil
}

// IsNotFound always returns false in JSON mode.
// Callers that previously checked IsNotFound should adapt to nil checks.
func IsNotFound(err error) bool {
	return false
}

// SaveConfig persists the JSON config to disk.
// Exported so callers that do direct saves can trigger it.
func SaveConfig() error {
	if cfgPath == "" {
		cfgPath = config.GetDBPath()
	}
	return db.Save(cfgPath)
}

// WithTx is a no-op stub for API compatibility.
// Previous GORM transaction closures should be replaced with direct calls.
func WithTx(fn func() error) error {
	return fn()
}
