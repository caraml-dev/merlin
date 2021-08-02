package cronjob

import "github.com/robfig/cron"

// CronJob wraps robfig.Cron.
type CronJob struct {
	c *cron.Cron
}

// New returns an initialized CronJob.
func New() (*CronJob, error) {
	return &CronJob{
		c: cron.New(),
	}, nil
}

// AddFunc adds a func to the CronJob to be run on the given schedule.
func (c *CronJob) AddFunc(spec string, cmd func()) error {
	return c.c.AddFunc(spec, cmd)
}

// Start the cron scheduler in its own go-routine.
func (c *CronJob) Start() {
	c.c.Start()
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
func (c *CronJob) Stop() {
	c.c.Stop()
}
