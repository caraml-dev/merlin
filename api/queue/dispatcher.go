package queue

import (
	"sync"

	"github.com/jinzhu/gorm"
)

type Dispatcher struct {
	sync.Mutex
	db         *gorm.DB
	workers    []*worker
	jobFuncMap map[string]JobFn
	jobChan    chan *Job
}

type JobFn func(*Job) error

type Config struct {
	NumWorkers int
	Db         *gorm.DB
}

type Job struct {
	ID   string
	Data interface{}
	Name string
}

func NewDispatcher(cfg Config) *Dispatcher {
	workers := make([]*worker, 0, cfg.NumWorkers)
	jobChan := make(chan *Job)
	jobFuncMap := make(map[string]JobFn)
	for i := 0; i < cfg.NumWorkers; i++ {
		worker := newWorker(cfg.Db, jobChan)
		workers = append(workers, worker)
	}
	return &Dispatcher{
		workers:    workers,
		db:         cfg.Db,
		jobFuncMap: jobFuncMap,
		jobChan:    jobChan,
	}
}

func (d *Dispatcher) EnqueueJob(job *Job) error {
	go func() {
		d.jobChan <- job
	}()

	return nil
}

func (d *Dispatcher) RegisterJob(jobName string, jobFn func(*Job) error) {
	d.Lock()
	d.jobFuncMap[jobName] = jobFn
	d.updateWorkersJobFunction()
	d.Unlock()
}

func (d *Dispatcher) StartWorkers() {
	for _, w := range d.workers {
		w.start()
	}
}

func (d *Dispatcher) updateWorkersJobFunction() {
	for _, w := range d.workers {
		w.updateWorkerJobFunction(d.jobFuncMap)
	}
}
