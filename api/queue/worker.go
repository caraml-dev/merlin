package queue

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type worker struct {
	quitChan   chan bool
	jobFuncMap map[string]JobFn
	jobChan    chan *Job
	db         *gorm.DB
}

func newWorker(db *gorm.DB, jobChan chan *Job) *worker {
	return &worker{db: db, jobChan: jobChan}
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
				err := w.processJob(job)
				if err != nil {

				}
			}
		}
	}()
}

func (w *worker) processJob(job *Job) error {
	// Lock database row
	// ctx := context.Background()
	// w.db.BeginTx(ctx, &sql.TxOptions{})
	// w.db.Exec("SELECT * FROM jobs WHERE id = ?")
	jobFn, ok := w.jobFuncMap[job.Name]
	if !ok {
		return fmt.Errorf("no function for job %s", job.Name)
	}
	if err := jobFn(job); err != nil {
		return err
	}

	return nil
}

func (w *worker) Stop() {
	go func() {
		w.quitChan <- true
	}()
}
