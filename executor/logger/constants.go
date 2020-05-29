package logger

const (
	LoggerWorkerQueueSize = 100
	CloudEventsIdHeader   = "Ce-Id"
	CloudEventsTypeHeader = "Ce-type"
	CloudEventsTypeSource = "Ce-source"
)

// Variables to cache the value of ENV default request logger
var defaultRequestLogger string = ""

func GetLoggerDefaultUrl() string {
	if defaultRequestLogger == "" {
		defaultRequestLogger = GetEnv("REQUEST_LOGGER_DEFAULT_ENDPOINT", "http://default-broker")
	}
	return defaultRequestLogger
}
