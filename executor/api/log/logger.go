package log

import (
	"github.com/go-logr/logr"
)

var Log Logger

func SetLogger(l logr.Logger) {
}

func ZapLogger(development bool) logr.Logger {
	return Logger{}
}

type Logger struct {
}

func (l Logger) Enabled() bool {
	return true
}

func (l Logger) Info(msg string, keysAndValues ...interface{}) {}

func (l Logger) Error(err error, msg string, keysAndValues ...interface{}) {}

func (l Logger) V(level int) logr.InfoLogger {
	return Logger{}
}

func (l Logger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return Logger{}
}

func (l Logger) WithName(name string) logr.Logger {
	return Logger{}
}
