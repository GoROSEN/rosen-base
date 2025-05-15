package cronjob

import (
	"github.com/google/martian/log"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	c *cron.Cron
}

var gScheduler Scheduler

func GetScheduler() *Scheduler {
	if gScheduler.c == nil {
		gScheduler.c = cron.New()
	}
	return &gScheduler
}

func (s *Scheduler) Start() {
	s.c.Start()
	log.Infof("cron job started")
}

func (s *Scheduler) Stop() {
	s.c.Stop()
	log.Infof("cron job stopped")
}

func (s *Scheduler) AddFunc(spec string, cmd func()) error {
	_, err := s.c.AddFunc(spec, cmd)
	return err
}
