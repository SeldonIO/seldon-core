/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import "github.com/seldonio/seldon-core/operator/v2/pkg/cli"

// Flags
const (
	flagAddHeader           = "header"
	flagAuthority           = "authority"
	flagFile                = "file-path"
	flagInferenceHost       = "inference-host"
	flagInferenceIterations = "iterations"
	flagInferenceSecs       = "seconds"
	flagInferenceMode       = "inference-mode"
	flagKafkaBroker         = "kafka-broker"
	flagSchedulerHost       = "scheduler-host"
	flagShowHeaders         = "show-headers"
	flagShowRequest         = "show-request"
	flagShowResponse        = "show-response"
	flagStickySession       = "sticky-session"
	flagWaitCondition       = "wait"
	flagTimeout             = "timeout-secs"
	flagVerbose             = "verbose"
	flagKafkaConfigPath     = "kafka-config-path"
	flagForceControlPlane   = "force"
)

// Env vars
const (
	envInfer           = "SELDON_INFER_HOST"
	envKafka           = "SELDON_KAFKA_BROKER"
	envScheduler       = "SELDON_SCHEDULE_HOST"
	envKafkaConfigPath = "SELDON_KAFKA_CONFIG_PATH"
	// Note that we use `POD_NAMESPACE` instead of `SELDON_NAMESPACE` to be consistent with
	// the environment variable used by other components
	envNamespace         = "POD_NAMESPACE"
	envForceControlPlane = "SELDON_FORCE_CONTROL_PLANE"
)

// Defaults
const (
	defaultInferHost         = "0.0.0.0:9000"
	defaultKafkaHost         = "0.0.0.0:9092"
	defaultSchedulerHost     = "0.0.0.0:9004"
	defaultForceControlPlane = false
)

// Help statements
const (
	helpAddHeader                = "add a header, e.g. key" + cli.HeaderSeparator + "value; use the flag multiple times to add more than one header"
	helpAuthority                = "authority (HTTP/2) or virtual host (HTTP/1)"
	helpFileInference            = "inference payload file"
	helpInferenceHost            = "seldon inference host"
	helpInferenceIterations      = "how many times to run inference"
	helpInferenceSecs            = "number of secs to run inference"
	helpInferenceMode            = "inference mode (rest or grpc)"
	helpSchedulerHost            = "seldon scheduler host"
	helpShowHeaders              = "show request and response headers"
	helpStickySession            = "use sticky session from last inference (only works with experiments)"
	helpForceControlPlane        = "force control plane mode (load model, etc.), default is false"
	helpForceControlPlaneWarning = "This command is likely to cause inconsistencies, enable with care."
)
