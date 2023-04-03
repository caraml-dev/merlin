package logger

import (
	"go.uber.org/zap"
)

type ConsoleSink struct {
	logger *zap.SugaredLogger
}

func NewConsoleSink(logger *zap.SugaredLogger) LogSink {
	return &ConsoleSink{logger: logger}
}

func (c *ConsoleSink) Sink(entries []*LogEntry) error {
	for _, message := range entries {
		kv := make([]interface{}, 0)
		if message.RequestPayload != nil {
			kv = append(kv, "request", string(message.RequestPayload.Body))
		}

		if message.ResponsePayload != nil {
			kv = append(kv, "response", string(message.ResponsePayload.Body))
			kv = append(kv, "statusCode", message.ResponsePayload.StatusCode)
		}

		if len(kv) == 0 {
			continue
		}

		c.logger.Infow("inference_log", kv...)
	}

	return nil
}
