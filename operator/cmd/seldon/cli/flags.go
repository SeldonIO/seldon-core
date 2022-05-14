package cli

const (
	fileFlag                = "file-path"
	schedulerHostFlag       = "scheduler-host"
	modelNameFlag           = "model-name"
	experimentFlag          = "experiment-name"
	showResponseFlag        = "show-response"
	showRequestFlag         = "show-request"
	waitConditionFlag       = "wait"
	inferenceHostFlag       = "inference-host"
	inferenceModeFlag       = "inference-mode"
	inferenceIterationsFlag = "iterations"

	EnvScheduler = "SELDON_SCHEDULE_HOST"
	EnvInfer     = "SELDON_INFER_HOST"

	DefaultScheduleHost = "0.0.0.0:9004"
	DefaultInferHost    = "0.0.0.0:9000"
)
