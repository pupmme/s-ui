	package service

import (
	"github.com/pupmme/pupmsub/core"
	"github.com/pupmme/pupmsub/db"
	"sort"
	"time"

)

type onlines struct {
	Inbound  []string `json:"inbound,omitempty"`
	User     []string `json:"user,omitempty"`
	Outbound []string `json:"outbound,omitempty"`
}

var onlineResources = &onlines{}

type StatsService struct{}

func (s *StatsService) SaveStats(enableTraffic bool) error {
	if core.GetCore() == nil || !core.GetCore().IsRunning() {
		return nil
	}
	box := core.GetCore().GetInstance()
	if box == nil {
		return nil
	}
	st := box.StatsTracker()
	if st == nil {
		return nil
	}
	stats := st.GetStats()

	// Reset onlines
	onlineResources.Inbound = nil
	onlineResources.Outbound = nil
	onlineResources.User = nil

	if len(*stats) == 0 {
		return nil
	}

	// Update client traffic from stats
	cfg := db.Get()
	for _, stat := range *stats {
		if stat.Resource == "user" {
			for i := range cfg.Clients {
				if cfg.Clients[i].Name == stat.Tag {
					if stat.Direction {
						cfg.Clients[i].Up += stat.Traffic
					} else {
						cfg.Clients[i].Down += stat.Traffic
					}
					break
				}
			}
		}
		if stat.Direction {
			switch stat.Resource {
			case "inbound":
				onlineResources.Inbound = append(onlineResources.Inbound, stat.Tag)
			case "outbound":
				onlineResources.Outbound = append(onlineResources.Outbound, stat.Tag)
			case "user":
				onlineResources.User = append(onlineResources.User, stat.Tag)
			}
		}
	}
	db.Set(cfg)

	if !enableTraffic {
		return nil
	}

	// Append new stats records (fetch cfg once, then batch)
	if enableTraffic && len(*stats) > 0 {
		cfg := db.Get()
		for _, stat := range *stats {
			cfg.Stats = append(cfg.Stats, db.Stat{
				DateTime:  stat.DateTime,
				Resource:  stat.Resource,
				Tag:       stat.Tag,
				Direction: stat.Direction,
				Traffic:   stat.Traffic,
			})
		}
		db.Set(cfg)
		return db.SaveConfig()
	}
	return nil
}

func (s *StatsService) GetStats(resource string, tag string, limit int) ([]db.Stat, error) {
	cfg := db.Get()
	currentTime := time.Now().Unix()
	timeDiff := currentTime - (int64(limit) * 3600)

	var result []db.Stat
	resources := []string{resource}
	if resource == "endpoint" {
		resources = []string{"inbound", "outbound"}
	}
	for _, stat := range cfg.Stats {
		if stat.DateTime > timeDiff {
			for _, r := range resources {
				if stat.Resource == r && stat.Tag == tag {
					result = append(result, db.Stat{
						DateTime:  stat.DateTime,
						Resource:  stat.Resource,
						Tag:       stat.Tag,
						Direction: stat.Direction,
						Traffic:   stat.Traffic,
					})
					break
				}
			}
		}
	}

	result = s.downsampleStats(result, 60)
	return result, nil
}

// downsampleStats reduces stats to maxRows rows.
func (s *StatsService) downsampleStats(stats []db.Stat, maxRows int) []db.Stat {
	if len(stats) <= maxRows {
		return stats
	}
	numBuckets := int(maxRows / 2)
	sort.Slice(stats, func(i, j int) bool { return stats[i].DateTime < stats[j].DateTime })
	timeMin, timeMax := stats[0].DateTime, stats[len(stats)-1].DateTime
	bucketSpan := (timeMax - timeMin) / int64(numBuckets)
	if bucketSpan == 0 {
		bucketSpan = 1
	}
	downsampled := make([]db.Stat, 0, maxRows)
	for i := 0; i < numBuckets; i++ {
		bucketStart := timeMin + int64(i)*bucketSpan
		bucketEnd := timeMin + int64(i+1)*bucketSpan
		if i == numBuckets-1 {
			bucketEnd = timeMax + 1
		}
		for _, dir := range []bool{false, true} {
			var sum int64
			var count int
			for _, r := range stats {
				if r.DateTime >= bucketStart && r.DateTime < bucketEnd && r.Direction == dir {
					sum += r.Traffic
					count++
				}
			}
			avg := int64(0)
			if count > 0 {
				avg = sum / int64(count)
			}
			downsampled = append(downsampled, db.Stat{
				DateTime:  bucketStart,
				Resource:  stats[0].Resource,
				Tag:       stats[0].Tag,
				Direction: dir,
				Traffic:   avg,
			})
		}
	}
	return downsampled
}

func (s *StatsService) GetOnlines() (onlines, error) {
	return *onlineResources, nil
}

func (s *StatsService) DelOldStats(days int) error {
	oldTime := time.Now().AddDate(0, 0, -days).Unix()
	cfg := db.Get()
	newStats := make([]db.Stat, 0, len(cfg.Stats))
	for _, stat := range cfg.Stats {
		if stat.DateTime >= oldTime {
			newStats = append(newStats, stat)
		}
	}
	cfg.Stats = newStats
	db.Set(cfg)
	return db.SaveConfig()
}
