package queue

import (
	"time"

	"github.com/gojek/merlin/log"
	"github.com/jinzhu/gorm"
)

var (
	findJobQuery           = "SELECT * FROM jobs where id = ? AND completed = ? FOR UPDATE SKIP LOCKED"
	allIncompleteJobsQuery = "SELECT * FROM jobs where completed = ? FOR UPDATE SKIP LOCKED"
)

type worker struct {
	quitChan   chan bool
	jobFuncMap map[string]JobFn
	jobChan    chan *Job
	db         *gorm.DB
}

func newWorker(db *gorm.DB, jobChan chan *Job) *worker {
	quitChan := make(chan bool)
	return &worker{db: db, jobChan: jobChan, quitChan: quitChan}
}

func (w *worker) updateWorkerJobFunction(jobFn map[string]JobFn) {
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
		log.Warnf("Job records not found with id: %d", job.ID)
		return
	}

	if err != nil {
		tx.Rollback()
		return
	}

	jobFn, ok := w.jobFuncMap[job.Name]
	if !ok {
		log.Warnf("There is no function be run for job %s", job.Name)
		tx.Rollback()
		return
	}
	if err := jobFn(job); err != nil {
		log.Errorf("Job execution is failed, with id:%d and error: %v", job.ID, err)
		tx.Rollback()
		return
	}

	refreshedJob.UpdatedAt = time.Now()
	refreshedJob.Completed = true
	if err := tx.Save(refreshedJob).Error; err != nil {
		log.Errorf("Failed to save job %d with error: %v", refreshedJob.ID, err)
		tx.Rollback()
		return
	}

	tx.Commit()
}

func (w *worker) stop() {
	go func() {
		w.quitChan <- true
	}()
}
