package cronjob

import (
	"github.com/pupmme/sub/core"
)

type CheckCoreJob struct{}

func NewCheckCoreJob() *CheckCoreJob {
	return &CheckCoreJob{}
}

func (s *CheckCoreJob) Run() {
	c := core.GetCore()
	if c != nil && !c.IsRunning() {
		_ = c.Start(nil)
	}
}
