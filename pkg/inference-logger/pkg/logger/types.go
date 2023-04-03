package logger

import "github.com/golang/protobuf/ptypes/timestamp"

type LogEntry struct {
	RequestId      string
	EventTimestamp *timestamp.Timestamp

	RequestPayload  *RequestPayload
	ResponsePayload *ResponsePayload
}

type RequestPayload struct {
	Headers map[string]string
	Body    []byte
}

type ResponsePayload struct {
	StatusCode int
	Body       []byte
}

type LogMode string

const (
	LogModeAll          LogMode = "all"
	LogModeRequestOnly  LogMode = "request"
	LogModeResponseOnly LogMode = "response"
)

type LoggerSinkKind = string

const (
	Kafka    LoggerSinkKind = "kafka"
	NewRelic LoggerSinkKind = "newrelic"
	Console  LoggerSinkKind = "console"
)

var LoggerSinkKinds = []LoggerSinkKind{Kafka, NewRelic, Console}
