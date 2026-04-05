package cronjob

import (
	"github.com/pupmme/sub/service"
)

type CheckCoreJob struct{}

func NewCheckCoreJob() *CheckCoreJob {
	return &CheckCoreJob{}
}

func (s *CheckCoreJob) Run() {
	c := service.GetCoreService()
	if c == nil {
		return
	}
	if !c.IsRunning() {
		_ = c.Restart()
	}
}
