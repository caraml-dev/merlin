package logger

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/merlin/pkg/inference-logger/pkg/mocks"
)

const (
	projectName  = "my-project"
	modelName    = "my-model"
	modelVersion = "1"
	serviceName  = "my-service"
	topicName    = "my-topic"

	minBatchSize  = 10
	maxBatchSize  = 32
	workQueueSize = 100
)

var workerConfig = &WorkerConfig{
	MinBatchSize: minBatchSize,
	MaxBatchSize: maxBatchSize,
}
var l = zap.NewNop()
var logger = l.Sugar()

func TestSubmit(t *testing.T) {
	mockNewRelicLogsClient := &mocks.NewRelicLogsClient{}
	mockNewRelicLogsClient.On("CreateLogEntry", mock.Anything).Return(nil)

	queue := NewBatchQueue(workQueueSize)
	worker := NewWorker(queue, workerConfig, logger, NewNewRelicSink(logger, mockNewRelicLogsClient, serviceName, projectName, modelName, modelVersion), NewConsoleSink(logger))
	worker.Start()

	logEntries := make([]*LogEntry, 0)
	for i := 0; i < minBatchSize; i++ {
		logEntry := &LogEntry{
			RequestId:      fmt.Sprint(i),
			EventTimestamp: timestamppb.Now(),
			RequestPayload: &RequestPayload{
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte(`{"instances" : [[1,2,3,4]]}`),
			},
			ResponsePayload: &ResponsePayload{
				StatusCode: 200,
				Body:       []byte(`{"predictions": [2]}`),
			},
		}
		err := worker.Submit(logEntry)
		assert.NoError(t, err)
		logEntries = append(logEntries, logEntry)
	}
	assert.Equal(t, len(logEntries), minBatchSize)

	// check that the work queue has been cleared
	time.Sleep(time.Millisecond * 5)
	assert.Equal(t, 0, queue.Len())

	// check that call to NewRelic client has been made
	mockNewRelicLogsClient.AssertCalled(t, "CreateLogEntry", mock.Anything)
	mockNewRelicLogsClient.AssertExpectations(t)
	worker.Stop()

}

func TestSubmitNoRequest(t *testing.T) {
	mockNewRelicLogsClient := &mocks.NewRelicLogsClient{}
	mockNewRelicLogsClient.On("CreateLogEntry", mock.Anything).Return(nil)

	queue := NewBatchQueue(workQueueSize)
	worker := NewWorker(queue, workerConfig, logger, NewNewRelicSink(logger, mockNewRelicLogsClient, serviceName, projectName, modelName, modelVersion), NewConsoleSink(logger))
	worker.Start()

	logEntries := make([]*LogEntry, 0)
	for i := 0; i < minBatchSize; i++ {
		logEntry := &LogEntry{
			RequestId:      fmt.Sprint(i),
			EventTimestamp: timestamppb.Now(),
			RequestPayload: nil,
			ResponsePayload: &ResponsePayload{
				StatusCode: 200,
				Body:       []byte(`{"predictions": [2]}`),
			},
		}
		err := worker.Submit(logEntry)
		assert.NoError(t, err)
		logEntries = append(logEntries, logEntry)
	}
	assert.Equal(t, len(logEntries), minBatchSize)

	// check that the work queue has been cleared
	time.Sleep(time.Millisecond * 5)
	assert.Equal(t, 0, queue.Len())

	// check that call to NewRelic client has been made
	mockNewRelicLogsClient.AssertCalled(t, "CreateLogEntry", mock.Anything)
	mockNewRelicLogsClient.AssertExpectations(t)
	worker.Stop()
}

func TestSubmitNoRequestResponse(t *testing.T) {
	mockNewRelicLogsClient := &mocks.NewRelicLogsClient{}
	mockNewRelicLogsClient.On("CreateLogEntry", mock.Anything).Return(nil)

	queue := NewBatchQueue(workQueueSize)
	worker := NewWorker(queue, workerConfig, logger, NewNewRelicSink(logger, mockNewRelicLogsClient, serviceName, projectName, modelName, modelVersion), NewConsoleSink(logger))
	worker.Start()

	logEntries := make([]*LogEntry, 0)
	for i := 0; i < minBatchSize; i++ {
		logEntry := &LogEntry{
			RequestId:       fmt.Sprint(i),
			EventTimestamp:  timestamppb.Now(),
			RequestPayload:  nil,
			ResponsePayload: nil,
		}
		err := worker.Submit(logEntry)
		assert.NoError(t, err)
		logEntries = append(logEntries, logEntry)
	}
	assert.Equal(t, len(logEntries), minBatchSize)

	// check that the work queue has been cleared
	time.Sleep(time.Millisecond * 5)
	assert.Equal(t, 0, queue.Len())

	// check that call to NewRelic client has been made
	mockNewRelicLogsClient.AssertCalled(t, "CreateLogEntry", mock.Anything)
	mockNewRelicLogsClient.AssertExpectations(t)
	worker.Stop()
}

func TestSubmitBuffered(t *testing.T) {
	mockNewRelicLogsClient := &mocks.NewRelicLogsClient{}
	mockNewRelicLogsClient.On("CreateLogEntry", mock.Anything).Return(nil)

	queue := NewBatchQueue(workQueueSize)
	worker := NewWorker(queue, workerConfig, logger, NewNewRelicSink(logger, mockNewRelicLogsClient, serviceName, projectName, modelName, modelVersion), NewConsoleSink(logger))
	worker.Start()

	// Send log as many as MinBatchSize - 1, so that it will be buffered
	for i := 0; i < minBatchSize-1; i++ {
		err := worker.Submit(&LogEntry{
			RequestId:      fmt.Sprint(i),
			EventTimestamp: timestamppb.Now(),
			RequestPayload: &RequestPayload{
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte(`{"instances" : [[1,2,3,4]]}`),
			},
			ResponsePayload: &ResponsePayload{
				StatusCode: 200,
				Body:       []byte(`{"predictions": [2]}`),
			},
		})

		if err != nil {
			t.Error(err)
		}
	}

	// check that the queue size is as expected and there is no call to NewRelic client
	assert.Equal(t, minBatchSize-1, queue.Len())
	mockNewRelicLogsClient.AssertNotCalled(t, "CreateLogEntry", mock.Anything)

	// submit another log entry which will trigger sending log to NewRelic client
	err := worker.Submit(&LogEntry{
		RequestId:      "X",
		EventTimestamp: timestamppb.Now(),
		RequestPayload: &RequestPayload{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"instances" : [[1,2,3,4]]}`),
		},
		ResponsePayload: &ResponsePayload{
			StatusCode: 200,
			Body:       []byte(`{"predictions": [2]}`),
		},
	})
	assert.NoError(t, err)

	// check that the work queue has been cleared and a call to NewRelic client has been made
	time.Sleep(time.Millisecond * 5)
	assert.Equal(t, 0, queue.Len())

	// check that call to NewRelic client has been made
	mockNewRelicLogsClient.AssertCalled(t, "CreateLogEntry", mock.Anything)
	mockNewRelicLogsClient.AssertExpectations(t)

	worker.Stop()
}

func BenchmarkSend10(b *testing.B) {
	benchmarkSend(b, 10)
}

func BenchmarkSend100(b *testing.B) {
	benchmarkSend(b, 100)
}

func BenchmarkSend1000(b *testing.B) {
	benchmarkSend(b, 1000)
}

func benchmarkSend(b *testing.B, n int) {
	b.StopTimer()
	b.ReportAllocs()

	mockNewRelicLogsClient := &mocks.NewRelicLogsClient{}
	mockNewRelicLogsClient.On("CreateLogEntry", mock.Anything).Return(nil)

	queue := NewBatchQueue(workQueueSize)
	worker := NewWorker(queue, workerConfig, logger, NewNewRelicSink(logger, mockNewRelicLogsClient, serviceName, projectName, modelName, modelVersion), NewConsoleSink(logger))
	worker.Start()

	logEntries := make([]*LogEntry, 0)
	for i := 0; i < n; i++ {
		logEntry := &LogEntry{
			RequestId:      fmt.Sprint(i),
			EventTimestamp: timestamppb.Now(),
			RequestPayload: &RequestPayload{
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte(`{"instances" : [[1,2,3,4]]}`),
			},
			ResponsePayload: &ResponsePayload{
				StatusCode: 200,
				Body:       []byte(`{"predictions": [2]}`),
			},
		}

		logEntries = append(logEntries, logEntry)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := worker.Send(logEntries)
		assert.NoError(b, err)
	}
	b.StopTimer()

	worker.Stop()
}
