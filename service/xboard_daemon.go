package service

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/pupmme/sub/config"
	"github.com/pupmme/sub/core"
	"github.com/pupmme/sub/db"
	"github.com/pupmme/sub/logger"
	"github.com/pupmme/sub/network"
)

// XboardDaemon runs as a background goroutine, handling all xboard communication.
// It replaces xboard-node entirely — s-ui manages itself as a xboard sub-node.
type XboardDaemon struct {
	client   *network.XboardClient
	syncSvc  *XboardSync
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	syncMu      sync.Mutex
	lastSync    time.Time
	lastTraffic time.Time
	connected   bool
}

// NewXboardDaemon creates the daemon (does not start it).
func NewXboardDaemon() *XboardDaemon {
	return &XboardDaemon{
		client:  network.NewXboardClient(),
		syncSvc: NewXboardSync(),
	}
}

// Start launches the daemon goroutines.
func (d *XboardDaemon) Start() {
	cfg := config.Get()
	if !cfg.Node {
		logger.Debug("[xboard-daemon] not in node mode, skipped")
		return
	}
	if cfg.Xboard.ApiHost == "" {
		logger.Warning("[xboard-daemon] no apiHost configured, skipped")
		return
	}

	d.ctx, d.cancel = context.WithCancel(context.Background())

	logger.Info("[xboard-daemon] starting, target: ", cfg.Xboard.ApiHost)

	// Initial handshake + sync
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.doHandshake()
	}()

	// Config sync loop (every 60s)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.syncLoop()
	}()

	// Traffic report loop (every 30s)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.reportLoop()
	}()

	// Status report loop (every 60s)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.statusLoop()
	}()

	logger.Info("[xboard-daemon] started")
}

// Stop gracefully shuts down the daemon.
func (d *XboardDaemon) Stop() {
	if d.cancel == nil {
		return
	}
	logger.Info("[xboard-daemon] stopping...")
	d.cancel()
	d.wg.Wait()
	logger.Info("[xboard-daemon] stopped")
}

// doHandshake connects to xboard and performs initial sync.
func (d *XboardDaemon) doHandshake() {
	for attempt := 1; attempt <= 3; attempt++ {
		hs, err := d.client.Handshake()
		if err == nil {
			d.connected = true
			logger.Info("[xboard-daemon] handshake ok, xboard version: ", hs.Version)
			if err := d.syncSvc.DoFullSync(); err != nil {
				logger.Error("[xboard-daemon] initial sync failed: ", err)
			}
			// Reload core with new config
			if c := core.GetCore(); c != nil {
				GetCoreService().Restart()
			}
			return
		}
		logger.Warning(fmt.Sprintf("[xboard-daemon] handshake attempt %d failed: %v", attempt, err))
		if attempt < 3 {
			time.Sleep(time.Duration(attempt*5) * time.Second)
		}
	}
	logger.Error("[xboard-daemon] all handshake attempts failed — daemon will retry in background")
}

// syncLoop periodically pulls config and user updates from xboard.
func (d *XboardDaemon) syncLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Retry handshake on disconnect
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if !d.connected {
				d.doHandshake()
				continue
			}
			d.doSync()
		}
	}
}

// reportLoop collects traffic from StatsService and pushes to xboard.
func (d *XboardDaemon) reportLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if !d.connected {
				continue
			}
			d.pushTraffic()
		}
	}
}

// statusLoop pushes system status to xboard.
func (d *XboardDaemon) statusLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if !d.connected {
				continue
			}
			d.pushStatus()
		}
	}
}

// doSync performs a single config+user sync.
func (d *XboardDaemon) doSync() {
	d.syncMu.Lock()
	defer d.syncMu.Unlock()

	cfg := config.Get()
	if !cfg.Node {
		return
	}

	// Get config updates
	nodeCfg, err := d.client.GetConfig()
	if err != nil {
		logger.Debug("[xboard-daemon] get config: ", err)
		return
	}
	if nodeCfg != nil {
		if err := d.syncSvc.applyInboundConfig(nodeCfg); err != nil {
			logger.Error("[xboard-daemon] apply inbound config: ", err)
		} else {
			logger.Debug("[xboard-daemon] inbound config updated, reloading core")
			if c := core.GetCore(); c != nil {
				GetCoreService().Restart()
			}
		}
	}

	// Get user updates
	users, err := d.client.GetUsers()
	if err != nil {
		logger.Debug("[xboard-daemon] get users: ", err)
		return
	}
	if users != nil {
		if err := d.syncSvc.applyUsers(users); err != nil {
			logger.Error("[xboard-daemon] apply users: ", err)
		}
	}

	d.lastSync = time.Now()
}

// pushTraffic collects current traffic stats and sends to xboard.
func (d *XboardDaemon) pushTraffic() {
	box := core.GetCore()
	if box == nil || !box.IsRunning() {
		return
	}

	statsSvc := &StatsService{}
	if err := statsSvc.SaveStats(false); err != nil {
		logger.Debug("[xboard-daemon] save stats: ", err)
		return
	}

	// Collect traffic from db clients (updated by SaveStats)
	cfg := db.Get()
	traffic := make(map[int64][2]int64)
	for _, c := range cfg.Clients {
		traffic[int64(c.Id)] = [2]int64{c.Up, c.Down}
	}

	if len(traffic) > 0 {
		if err := d.syncSvc.ReportTraffic(traffic); err != nil {
			logger.Debug("[xboard-daemon] push traffic: ", err)
		}
	}

	d.lastTraffic = time.Now()
}

// pushStatus collects and sends system status to xboard.
func (d *XboardDaemon) pushStatus() {
	box := core.GetCore()
	if box == nil || !box.IsRunning() {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	cpu := 0.0 // CPU measurement would need per-interval sampling
	mem := [2]uint64{m.Alloc, m.Sys}
	swap := [2]uint64{0, 0}
	disk := [2]uint64{0, 0}

	if err := d.client.PushStatus(cpu, mem, swap, disk); err != nil {
		logger.Debug("[xboard-daemon] push status: ", err)
	}
}

// IsConnected returns the current connection status.
func (d *XboardDaemon) IsConnected() bool {
	return d.connected
}

// TriggerSync forces an immediate sync (called from API or CLI).
func (d *XboardDaemon) TriggerSync() error {
	if !d.connected {
		return fmt.Errorf("not connected to xboard")
	}
	d.doSync()
	return nil
}

// GetStatus returns daemon status for the API.
func (d *XboardDaemon) GetStatus() map[string]interface{} {
	d.syncMu.Lock()
	defer d.syncMu.Unlock()
	cfg := config.Get()
	return map[string]interface{}{
		"enabled":    cfg.Node,
		"connected":  d.connected,
		"lastSync":   d.lastSync.Format(time.RFC3339),
		"lastReport":  d.lastTraffic.Format(time.RFC3339),
		"apiHost":    cfg.Xboard.ApiHost,
		"nodeId":     cfg.Xboard.NodeID,
		"nodeType":   cfg.Xboard.NodeType,
	}
}

// marshalInbounds is a helper.
func marshalInboundsJSON(ids []uint) []byte {
	data, _ := json.Marshal(ids)
	return data
}
