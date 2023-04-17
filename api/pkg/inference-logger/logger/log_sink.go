package logger

type LogSink interface {
	Sink(rawLogEntries []*LogEntry) error
}
