package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pupmme/pupmsub/config"
	"github.com/pupmme/pupmsub/core"
	"github.com/pupmme/pupmsub/db"
	"github.com/pupmme/pupmsub/logger"
	"github.com/pupmme/pupmsub/network"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// XboardDaemon runs as a background goroutine, handling all xboard communication.
// It replaces xboard-node entirely — pupmsub manages itself as a xboard sub-node.
type XboardDaemon struct {
	client  *network.XboardClient
	syncSvc *XboardSync
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	syncMu      sync.Mutex
	lastSync    time.Time
	lastTraffic time.Time
	connected   bool
	connMu      sync.RWMutex // ★ protects connected

	// Per-user last seen traffic for delta computation
	lastTrafficSnap map[int64][2]int64
}

// NewXboardDaemon creates the daemon (does not start it).
func NewXboardDaemon() *XboardDaemon {
	client := network.NewXboardClient()
	return &XboardDaemon{
		client:           client,
		syncSvc:          NewXboardSync(client),
		lastTrafficSnap: make(map[int64][2]int64),
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
			d.setConnected(true)
			logger.Info("[xboard-daemon] handshake ok, xboard version: ", hs.Version)

			// Initialize traffic baseline from persisted DB before any report.
			// This prevents silent traffic loss if pupmsub restarted while users were active.
			d.syncMu.Lock()
			if cfg := db.Get(); cfg != nil {
				for _, c := range cfg.Clients {
					d.lastTrafficSnap[int64(c.Id)] = [2]int64{c.Up, c.Down}
				}
				logger.Info("[xboard-daemon] traffic baseline loaded for ", len(cfg.Clients), " users")
			}
			d.syncMu.Unlock()

			// Pass hs data directly instead of re-fetching via Sync(false)
			if err := d.syncSvc.SyncWithHandshake(hs); err != nil {
				logger.Error("[xboard-daemon] initial sync failed: ", err)
			}
			if c := GetCoreService(); c != nil {
				_ = c.Restart()
			}
			return
		}
		logger.Warning(fmt.Sprintf("[xboard-daemon] handshake attempt %d failed: %v", attempt, err))
		if attempt < 3 {
			time.Sleep(time.Duration(attempt*5) * time.Second)
		}
	}
	// ETag cache may be stale after prolonged failures — clear it so the next
	// syncLoop attempt does not get stuck on 304 Not Modified.
	d.client.ResetETags()
	logger.Error("[xboard-daemon] all handshake attempts failed — daemon will retry in background")
}

// syncLoop periodically pulls config and user updates from xboard.
func (d *XboardDaemon) syncLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if !d.isConnected() {
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
			if !d.isConnected() {
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
			if !d.isConnected() {
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
			if c := GetCoreService(); c != nil {
				_ = c.Restart()
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

// pushTraffic collects current traffic stats, computes delta, and sends to xboard.
func (d *XboardDaemon) pushTraffic() {
	box := core.GetCore()
	if box == nil || !box.IsRunning() {
		return
	}

	statsSvc := &StatsService{}
	if err := statsSvc.SaveStats(false); err != nil {
		logger.Debug("[xboard-daemon] save stats: ", err)
	}

	d.syncMu.Lock()
	cfg := db.Get()
	delta := make(map[int64][2]int64)
	for _, c := range cfg.Clients {
		last, ok := d.lastTrafficSnap[int64(c.Id)]
		if ok {
			delta[int64(c.Id)] = [2]int64{c.Up - last[0], c.Down - last[1]}
		} else {
			// Baseline was 0 — report zero delta to establish new baseline
			delta[int64(c.Id)] = [2]int64{0, 0}
		}
		d.lastTrafficSnap[int64(c.Id)] = [2]int64{c.Up, c.Down}
	}
	d.syncMu.Unlock()

	if len(delta) > 0 {
		if err := d.syncSvc.ReportTraffic(delta); err != nil {
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

	cpuPct, _ := cpu.Percent(0, false)
	var cpuVal float64
	if len(cpuPct) > 0 {
		cpuVal = cpuPct[0]
	}

	memInfo, _ := mem.VirtualMemory()
	memTotal, memUsed := uint64(0), uint64(0)
	if memInfo != nil {
		memTotal = memInfo.Total
		memUsed = memInfo.Used
	}

	swapTotal, swapUsed := uint64(0), uint64(0)
	// disk usage of root partition
	parts, _ := disk.Partitions(false)
	var diskTotal, diskUsed uint64
	for _, p := range parts {
		if p.Mountpoint == "/" || p.Mountpoint == "" {
			if usage, err := disk.Usage(p.Mountpoint); err == nil {
				diskTotal = usage.Total
				diskUsed = usage.Used
			}
			break
		}
	}

	if err := d.client.PushStatus(cpuVal,
		[2]uint64{memTotal, memUsed},
		[2]uint64{swapTotal, swapUsed},
		[2]uint64{diskTotal, diskUsed},
	); err != nil {
		logger.Debug("[xboard-daemon] push status: ", err)
	}
}

// IsConnected returns the current connection status.
func (d *XboardDaemon) IsConnected() bool {
	return d.isConnected()
}

// TriggerSync forces an immediate sync (called from API or CLI).
func (d *XboardDaemon) TriggerSync() error {
	if !d.isConnected() {
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
		"connected":  d.isConnected(),
		"lastSync":   d.lastSync.Format(time.RFC3339),
		"lastReport": d.lastTraffic.Format(time.RFC3339),
		"apiHost":    cfg.Xboard.ApiHost,
		"nodeId":     cfg.Xboard.NodeID,
		"nodeType":   cfg.Xboard.NodeType,
	}
}

// isConnected returns d.connected with read lock.
func (d *XboardDaemon) isConnected() bool {
	d.connMu.RLock()
	defer d.connMu.RUnlock()
	return d.connected
}

// setConnected sets d.connected with write lock.
func (d *XboardDaemon) setConnected(v bool) {
	d.connMu.Lock()
	defer d.connMu.Unlock()
	d.connected = v
}
