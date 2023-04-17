package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type Dispatcher struct {
	workerQueue *BatchQueue
	workers     []*Worker
}

func NewDispatcher(nworkers int,
	queueSize int,
	workerConfig *WorkerConfig,
	logger *zap.SugaredLogger,
	logSinks ...LogSink) *Dispatcher {

	workerQueue := NewBatchQueue(queueSize)

	workers := make([]*Worker, 0)
	for i := 0; i < nworkers; i++ {
		worker := NewWorker(workerQueue, workerConfig, logger, logSinks...)
		workers = append(workers, worker)
	}

	return &Dispatcher{
		workerQueue: workerQueue,
		workers:     workers,
	}
}

func (d *Dispatcher) Start() {
	for _, worker := range d.workers {
		worker.Start()
	}
}

func (d *Dispatcher) Submit(logEntry *LogEntry) error {
	return d.workerQueue.Put(logEntry)
}

func (d *Dispatcher) Stop() {
	err := d.workerQueue.Close()
	if err != nil {
		fmt.Println(err)
	}
}
