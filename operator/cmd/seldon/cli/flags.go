package cli

const (
	fileFlag                = "file-path"
	schedulerHostFlag       = "scheduler-host"
	showResponseFlag        = "show-response"
	showRequestFlag         = "show-request"
	waitConditionFlag       = "wait"
	inferenceHostFlag       = "inference-host"
	inferenceModeFlag       = "inference-mode"
	inferenceIterationsFlag = "iterations"
	kafkaBrokerFlag         = "kafka-broker"

	// Defaults

	EnvScheduler = "SELDON_SCHEDULE_HOST"
	EnvInfer     = "SELDON_INFER_HOST"
	EnvKafka     = "SELDON_KAFKA_BROKER"

	DefaultScheduleHost = "0.0.0.0:9004"
	DefaultInferHost    = "0.0.0.0:9000"
	DefaultKafkaHost    = "0.0.0.0:9092"
)
