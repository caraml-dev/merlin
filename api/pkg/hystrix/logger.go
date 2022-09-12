package hystrix

import "go.uber.org/zap"

// hystrxLogger implements https://github.com/afex/hystrix-go/blob/master/hystrix/logger.go
type hystrixLogger struct {
	zapLogger *zap.SugaredLogger
}

// newHystrixLogger create new instance of hystrixLogger backed by zap sugared logger.
func NewHystrixLogger(zapLogger *zap.Logger) *hystrixLogger {
	return &hystrixLogger{zapLogger: zapLogger.Sugar()}
}

// Printf will format and log the arguments as INFO log
func (l *hystrixLogger) Printf(format string, items ...interface{}) {
	l.zapLogger.Infof(format, items)
}
