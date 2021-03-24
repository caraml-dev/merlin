package queue

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gojek/merlin/log"
	"github.com/jinzhu/gorm"
)

type Producer interface {
	EnqueueJob(job *Job) error
}

type Consumer interface {
	RegisterJob(jobName string, jobFn func(*Job) error)
	Start()
	Stop()
}

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

type Status string

type Argument map[string]interface{}

type Job struct {
	ID        int64     `json:"id" gorm:"primary_key;"`
	Arguments Argument  `json:"arguments"`
	Completed bool      `json:"completed"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a Argument) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Argument) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
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
	var savedJob Job
	if err := d.db.Save(job).Scan(&savedJob).Error; err != nil {
		log.Errorf("Failed to save job %d with error: %v", job.ID, err)
		return err
	}
	go func() {
		d.jobChan <- &savedJob
	}()

	return nil
}

func (d *Dispatcher) RegisterJob(jobName string, jobFn func(*Job) error) {
	d.Lock()
	d.jobFuncMap[jobName] = jobFn
	d.updateWorkersJobFunction()
	d.Unlock()
}

func (d *Dispatcher) Start() {
	// When start workers make sure all incomplete jobs which is not currently running is rerun
	d.reQueueIncompleteJobs()
	for _, w := range d.workers {
		w.start()
	}
}

func (d *Dispatcher) Stop() {
	for _, w := range d.workers {
		w.stop()
	}
}

func (d *Dispatcher) reQueueIncompleteJobs() {
	var jobs []Job

	err := d.db.Raw(allIncompleteJobsQuery, false).Scan(&jobs).Error
	if err != nil {
		return
	}
	for _, job := range jobs {
		d.EnqueueJob(&job)
	}
}

func (d *Dispatcher) updateWorkersJobFunction() {
	for _, w := range d.workers {
		w.updateWorkerJobFunction(d.jobFuncMap)
	}
}
