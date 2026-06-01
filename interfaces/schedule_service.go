package interfaces

import (
	"time"

	"github.com/robfig/cron/v3"
)

type ScheduleService interface {
	WithConfigPath
	Module
	GetLocation() (loc *time.Location)
	SetLocation(loc *time.Location)
	GetDelay() (delay bool)
	SetDelay(delay bool)
	GetSkip() (skip bool)
	SetSkip(skip bool)
	GetUpdateInterval() (interval time.Duration)
	SetUpdateInterval(interval time.Duration)
	Enable(s Schedule, args ...any) (err error)
	Disable(s Schedule, args ...any) (err error)
	Update()
	GetCron() (c *cron.Cron)
}
