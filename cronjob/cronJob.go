package cronjob

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	cron *cron.Cron
}

func NewCronJob() *CronJob {
	return &CronJob{}
}

func (c *CronJob) Start(loc *time.Location, trafficAge int) error {
	c.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())
	c.cron.Start()

	go func() {
		c.cron.AddJob("@every 5s", NewCheckCoreJob())
	}()

	return nil
}

func (c *CronJob) Stop() {
	if c.cron != nil {
		c.cron.Stop()
	}
}
