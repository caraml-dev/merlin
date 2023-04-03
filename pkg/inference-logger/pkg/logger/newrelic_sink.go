package logger

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	mlogs "github.com/caraml-dev/merlin/pkg/log"
)

const (
	// Hostname
	Hostname = "merlin-log-collector"
)

type NewRelicLogsClient interface {
	CreateLogEntry(logEntry interface{}) error
}

type NewRelicSink struct {
	logger    *zap.SugaredLogger
	logClient NewRelicLogsClient

	serviceName string

	projectName  string
	modelName    string
	modelVersion string
}

func NewNewRelicSink(
	logger *zap.SugaredLogger,
	logClient NewRelicLogsClient,
	serviceName string,
	projectName string,
	modelName string,
	modelVersion string,
) LogSink {
	return &NewRelicSink{
		logger:      logger,
		logClient:   logClient,
		serviceName: serviceName,

		projectName:  projectName,
		modelName:    modelName,
		modelVersion: modelVersion,
	}
}

type NewRelicLogEntry struct {
	Attributes map[string]string `json:"attributes"`
	Message    string            `json:"message"`
}

func (n *NewRelicSink) Sink(rawLogEntries []*LogEntry) error {
	// Format log inputs
	inferenceLogs, err := n.newLogMessages(rawLogEntries)
	if err != nil {
		return err
	}

	formattedLogs, err := n.buildNewRelicLogEntry(inferenceLogs)
	if err != nil {
		return err
	}

	for _, msg := range formattedLogs {
		err = n.logClient.CreateLogEntry(msg)
		if err != nil {
			n.logger.Fatal("error creating log entry: ", err)
			return err
		}
	}

	return nil
}

func (n *NewRelicSink) buildNewRelicLogEntry(merlinLogs []*mlogs.InferenceLogMessage) ([]NewRelicLogEntry, error) {
	var messages []NewRelicLogEntry
	for _, merlinLog := range merlinLogs {
		// Format Attributes
		formattedAttributes := map[string]string{
			"hostname":       Hostname,
			"requestId":      merlinLog.RequestId,
			"eventTimestamp": merlinLog.EventTimestamp.String(),
			"projectName":    merlinLog.ProjectName,
			"modelName":      merlinLog.ModelName,
			"modelVersion":   merlinLog.ModelVersion,
			"serviceName":    n.serviceName,
		}
		// Format Message
		bytes, err := protojson.Marshal(merlinLog)
		if err != nil {
			return nil, err
		}

		newRelicEntry := NewRelicLogEntry{
			Attributes: formattedAttributes,
			Message:    string(bytes),
		}
		messages = append(messages, newRelicEntry)
	}

	return messages, nil
}

func (n *NewRelicSink) newLogMessages(logEntries []*LogEntry) ([]*mlogs.InferenceLogMessage, error) {
	messages := make([]*mlogs.InferenceLogMessage, 0)
	for _, logEntry := range logEntries {

		request := &mlogs.Request{}
		if logEntry.RequestPayload != nil {
			request = &mlogs.Request{
				Header: logEntry.RequestPayload.Headers,
				Body:   string(logEntry.RequestPayload.Body),
			}
		}

		response := &mlogs.Response{}
		if logEntry.ResponsePayload != nil {
			response = &mlogs.Response{
				StatusCode: int32(logEntry.ResponsePayload.StatusCode),
				Body:       string(logEntry.ResponsePayload.Body),
			}
		}

		logMessage := &mlogs.InferenceLogMessage{
			RequestId:      logEntry.RequestId,
			EventTimestamp: logEntry.EventTimestamp,
			ProjectName:    n.projectName,
			ModelName:      n.modelName,
			ModelVersion:   n.modelVersion,
			Request:        request,
			Response:       response,
		}

		messages = append(messages, logMessage)
	}

	return messages, nil
}
