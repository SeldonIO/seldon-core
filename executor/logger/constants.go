package logger

import "os"

const (
	LoggerWorkerQueueSize = 100
	CloudEventsIdHeader   = "Ce-Id"
	CloudEventsTypeHeader = "Ce-type"
	CloudEventsTypeSource = "Ce-source"
)

// Variable to cache the value of ENV default request logger
var defaultRequestLoggerEndpointPrefix string = ""

func GetLoggerDefaultUrl(namespace string) string {
	if defaultRequestLoggerEndpointPrefix == "" {
		if value, ok := os.LookupEnv("REQUEST_LOGGER_DEFAULT_ENDPOINT_PREFIX"); ok && value != "" {
			defaultRequestLoggerEndpointPrefix = value
		} else {
			defaultRequestLoggerEndpointPrefix = "http://default-broker."
		}
	}
	return defaultRequestLoggerEndpointPrefix + namespace
}
