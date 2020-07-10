package log

import (
	"github.com/go-logr/logr"
)

// nullLogger is a no-op logger enabled by default
type nullLogger struct {
	logr.Logger
}

func (nullLogger) Info(_ string, _ ...interface{}) {}

func (nullLogger) Enabled() bool {
	return false
}

func (nullLogger) Error(_ error, _ string, _ ...interface{}) {}

func (n *nullLogger) V(_ int) logr.InfoLogger {
	return n
}

func (n *nullLogger) WithName(_ string) logr.Logger {
	return n
}

func (n *nullLogger) WithValues(_ ...interface{}) logr.Logger {
	return n
}
