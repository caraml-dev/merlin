package queue

import (
	"sync"

	"github.com/gojek/merlin/log"
	"github.com/jinzhu/gorm"
)

type Producer interface {
	// EnqueueJob return error indicate that whether job is successfully queue or not
	EnqueueJob(job *Job) error
}

type Consumer interface {
	// RegisterJob will register function that will be run for specific job name
	RegisterJob(jobName string, jobFn JobFn)
	// Start will run consumer workers which will consume job from queue
	Start()
	// Stop will stop all consumer workers
	Stop()
}

type Dispatcher struct {
	db         *gorm.DB
	workers    []*worker
	jobFuncMap *sync.Map
	jobChan    chan *Job
}

type JobFn func(*Job) error

type Config struct {
	NumWorkers int
	Db         *gorm.DB
}

func NewDispatcher(cfg Config) *Dispatcher {
	workers := make([]*worker, 0, cfg.NumWorkers)
	jobChan := make(chan *Job)
	jobFuncMap := &sync.Map{}
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

func (d *Dispatcher) RegisterJob(jobName string, jobFn JobFn) {
	d.jobFuncMap.Store(jobName, jobFn)
	d.updateWorkersJobFunction()
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
