package logger

const (
	LoggerWorkerQueueSize = 100
	CloudEventsIdHeader   = "Ce-Id"
	CloudEventsTypeHeader = "Ce-type"
	CloudEventsTypeSource = "Ce-source"
)

func GetLoggerDefaultUrl(namespace string) string {
	return "http://default-broker." + namespace
}
