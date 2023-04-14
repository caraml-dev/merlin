package logger

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Worker struct {
	id                         string
	workerQueue                *BatchQueue
	minBatchSize, maxBatchSize int

	quitChan chan bool

	logger   *zap.SugaredLogger
	logSinks []LogSink
}

type WorkerConfig struct {
	MinBatchSize int
	MaxBatchSize int
}

// NewWorker creates a worker that listen to work in workerQueue
func NewWorker(workerQueue *BatchQueue, workerConfig *WorkerConfig, logger *zap.SugaredLogger, logSinks ...LogSink) *Worker {
	return &Worker{
		id:           uuid.New().String(),
		workerQueue:  workerQueue,
		minBatchSize: workerConfig.MinBatchSize,
		maxBatchSize: workerConfig.MaxBatchSize,
		logger:       logger,
		logSinks:     logSinks,
	}
}

// Start start worker and listen to work submitted to the workerQueue
//
// Call Stop to stop the worker from performing further works
func (w *Worker) Start() {
	go func() {
		w.logger.Infof("Starting worker: %s\n", w.id)
		defer func() {
			w.logger.Infof("worker %s is stopped, exiting... \n", w.id)
		}()

		for {

			select {
			case <-w.quitChan:
				return

			default:
				msgs := w.workerQueue.GetMinMax(w.minBatchSize, w.maxBatchSize)
				if len(msgs) == 0 {
					w.logger.Infof("Worker Queue has been closed")
					msgs = w.workerQueue.GetAll()
					if len(msgs) < 1 {
						w.logger.Infof("no more queued message, quiting..")
						return
					}
					w.logger.Infof("processing remaining %d messages", len(msgs))
					for _, msg := range msgs {
						logEntry, ok := msg.(*LogEntry)
						if !ok {
							w.logger.Errorf("invalid type %v", logEntry)
						}
					}
				}

				logEntries := make([]*LogEntry, 0)
				for _, msg := range msgs {
					logEntry, ok := msg.(*LogEntry)
					if !ok {
						w.logger.Errorf("invalid type %v", logEntry)
					}
					logEntries = append(logEntries, logEntry)
				}

				err := w.Send(logEntries)
				if err != nil {
					w.logger.Errorf("error processing log entry: %v", err)
				}
			}
		}
	}()
}

// Submit submit a log entry to worker to be sent.
//
// Note that the log entry might be buffered before being sent to the log sink.
func (w *Worker) Submit(logEntry *LogEntry) error {
	return w.workerQueue.Put(logEntry)
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its on-going work.
func (w *Worker) Stop() {
	go func() {
		w.quitChan <- true
	}()
}

func (w *Worker) Send(rawLogEntries []*LogEntry) error {
	for _, logSink := range w.logSinks {
		err := logSink.Sink(rawLogEntries)
		if err != nil {
			return err
		}
	}
	return nil
}
