package cli

import "github.com/seldonio/seldon-core/operatorv2/pkg/cli"

// Flags
const (
	flagAddHeader           = "header"
	flagAuthority           = "authority"
	flagFile                = "file-path"
	flagInferenceHost       = "inference-host"
	flagInferenceIterations = "iterations"
	flagInferenceMode       = "inference-mode"
	flagKafkaBroker         = "kafka-broker"
	flagSchedulerHost       = "scheduler-host"
	flagShowHeaders         = "show-headers"
	flagShowRequest         = "show-request"
	flagShowResponse        = "show-response"
	flagStickySession       = "sticky-session"
	flagWaitCondition       = "wait"
)

// Env vars
const (
	envInfer     = "SELDON_INFER_HOST"
	envKafka     = "SELDON_KAFKA_BROKER"
	envScheduler = "SELDON_SCHEDULE_HOST"
)

// Defaults
const (
	defaultInferHost     = "0.0.0.0:9000"
	defaultKafkaHost     = "0.0.0.0:9092"
	defaultSchedulerHost = "0.0.0.0:9004"
)

// Help statements
const (
	helpAddHeader           = "add a header, e.g. key" + cli.HeaderSeparator + "value; use the flag multiple times to add more than one header"
	helpAuthority           = "authority (HTTP/2) or virtual host (HTTP/1)"
	helpFileInference       = "inference payload file"
	helpInferenceHost       = "seldon inference host"
	helpInferenceIterations = "how many times to run inference"
	helpInferenceMode       = "inference mode (rest or grpc)"
	helpSchedulerHost       = "seldon scheduler host"
	helpShowHeaders         = "show request and response headers"
	helpStickySession       = "use sticky session from last inference (only works with experiments)"
)
