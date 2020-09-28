// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"go.uber.org/zap"
)

var globalLogger = NewLogger()

// NewLogger create a new SugaredLogger
func NewLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction(zap.AddCallerSkip(1))
	return logger.Sugar()
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	globalLogger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	globalLogger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	globalLogger.Errorf(template, args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message.
func Fatalf(template string, args ...interface{}) {
	globalLogger.Fatalf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message.
func Panicf(template string, args ...interface{}) {
	globalLogger.Panicf(template, args...)
}
