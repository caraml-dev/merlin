package queue

import (
	"sync"
	"time"

	"github.com/caraml-dev/merlin/log"
	"github.com/jinzhu/gorm"
)

var (
	findJobQuery           = "SELECT * FROM jobs where id = ? AND completed = ? FOR UPDATE SKIP LOCKED"
	allIncompleteJobsQuery = "SELECT * FROM jobs where completed = ? FOR UPDATE SKIP LOCKED"
	delayOfRetry           = 5 * time.Second
)

type worker struct {
	quitChan   chan bool
	jobFuncMap *sync.Map
	jobChan    chan *Job
	db         *gorm.DB
}

func newWorker(db *gorm.DB, jobChan chan *Job) *worker {
	quitChan := make(chan bool)
	return &worker{db: db, jobChan: jobChan, quitChan: quitChan}
}

func (w *worker) updateWorkerJobFunction(jobFn *sync.Map) {
	w.jobFuncMap = jobFn
}

func (w *worker) start() {
	go func() {
		for {
			select {
			case <-w.quitChan:
				return
			case job := <-w.jobChan:
				go w.processJob(job)
			}
		}
	}()
}

func (w *worker) processJob(job *Job) {
	var refreshedJob Job
	tx := w.db.Begin()
	err := tx.Raw(findJobQuery, job.ID, false).Scan(&refreshedJob).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		log.Warnf("Job with ID:%d is still running in other process or already ran successfully", job.ID)
		return
	}

	if err != nil {
		tx.Rollback()
		w.requeueJob(job)
		return
	}

	fn, ok := w.jobFuncMap.Load(job.Name)
	if !ok {
		log.Warnf("There is no function be run for job %s", job.Name)
		tx.Rollback()
		return
	}

	jobFn, ok := fn.(JobFn)
	if !ok {
		log.Warnf("Registered function is not correct")
		tx.Rollback()
		return
	}
	if err := jobFn(job); err != nil {
		log.Errorf("Job execution is failed, with id:%d and error: %v", job.ID, err)
		switch err.(type) {
		case RetryableError:
			w.requeueJob(job)
			return
		}
	}

	refreshedJob.UpdatedAt = time.Now()
	refreshedJob.Completed = true
	if err := tx.Save(refreshedJob).Error; err != nil {
		log.Errorf("Failed to save job %d with error: %v", refreshedJob.ID, err)
		tx.Rollback()

		w.requeueJob(job)
		return
	}

	tx.Commit()
}

func (w *worker) requeueJob(job *Job) {
	time.Sleep(delayOfRetry) // sleep for specific amount of time before retry
	go func() {
		w.jobChan <- job
	}()
}

func (w *worker) stop() {
	go func() {
		w.quitChan <- true
	}()
}
