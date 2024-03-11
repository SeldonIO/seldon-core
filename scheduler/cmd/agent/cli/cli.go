/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	envServerHttpPort                 = "SELDON_SERVER_HTTP_PORT"
	envServerGrpcPort                 = "SELDON_SERVER_GRPC_PORT"
	envReverseProxyHttpPort           = "SELDON_REVERSE_PROXY_HTTP_PORT"
	envReverseProxyGrpcPort           = "SELDON_REVERSE_PROXY_GRPC_PORT"
	envDebugGrpcPort                  = "SELDON_DEBUG_GRPC_PORT"
	envMetricsPort                    = "SELDON_METRICS_PORT"
	envPodName                        = "POD_NAME"
	envSchedulerHost                  = "SELDON_SCHEDULER_HOST"
	envSchedulerPort                  = "SELDON_SCHEDULER_PORT"
	envSchedulerTlsPort               = "SELDON_SCHEDULER_TLS_PORT"
	envReplicaConfig                  = "SELDON_REPLICA_CONFIG"
	envLogLevel                       = "SELDON_LOG_LEVEL"
	envServerType                     = "SELDON_SERVER_TYPE"
	envMemoryRequest                  = "MEMORY_REQUEST"
	envCapabilities                   = "SELDON_SERVER_CAPABILITIES"
	envOverCommitPercentage           = "SELDON_OVERCOMMIT_PERCENTAGE"
	envEnvoyHost                      = "SELDON_ENVOY_HOST"
	envEnvoyPort                      = "SELDON_ENVOY_PORT"
	envDrainerServicePort             = "SELDON_DRAINER_PORT"
	envModelInferenceLagThreshold     = "SELDON_MODEL_INFERENCE_LAG_THRESHOLD"
	envModelInferenceDelayMSThreshold = "SELDON_MODEL_INFERENCE_DELAY_MS_THRESHOLD"
	envModelInactiveSecondsThreshold  = "SELDON_MODEL_INACTIVE_SECONDS_THRESHOLD"
	envScalingStatsPeriodSeconds      = "SELDON_SCALING_STATS_PERIOD_SECONDS"

	flagSchedulerHost                  = "scheduler-host"
	flagSchedulerPlaintxtPort          = "scheduler-port"
	flagSchedulerTlsPort               = "scheduler-tls-port"
	flagServerName                     = "server-name"
	flagServerIdx                      = "server-idx"
	flagInferenceHttpPort              = "inference-http-port"
	flagInferenceGrpcPort              = "inference-grpc-port"
	flagReverseProxyHttpPort           = "reverse-proxy-http-port"
	flagReverseProxyGrpcPort           = "reverse-proxy-grpc-port"
	flagDebugGrpcPort                  = "debug-grpc-port"
	flagMetricsPort                    = "metrics-port"
	flagReplicaConfig                  = "replica-config"
	flagLogLevel                       = "log-level"
	flagServerType                     = "server-type"
	flagMemoryBytes                    = "memory-bytes"
	flagCapabilities                   = "capabilities"
	flagOverCommitPercentage           = "over-commit-percentage"
	flagTracingConfigPath              = "tracing-config-path"
	flagEnvoyHost                      = "envoy-host"
	flagEnvoyPort                      = "envoy-port"
	flagDrainerServicePort             = "drainer-port"
	flagModelInferenceLagThreshold     = "model-inference-lag-threshold"
	flagModelInferenceDelayMSThreshold = "model-inference-delay-ms-threshold"
	flagModelInactiveSecondsThreshold  = "model-inactive-seconds-threshold"
	flagScalingStatsPeriodSeconds      = "scaling-stats-period-seconds"
)

const (
	defaultInferenceHttpPort        = 8080
	defaultInferenceGrpcPort        = 9500
	defaultRclonePort               = 5572
	defaultSchedulerPort            = 9005
	defaultSchedulerTlsPort         = 9055
	defaultMetricsPort              = 9006
	defaultEnvoyHost                = "0.0.0.0"
	defaultEnvoyPort                = 9000
	defaultDrainerServicePort       = 9007
	statsPeriodSecondsDefault       = 5
	lagThresholdDefault             = 30
	delayMSThresholdDefault         = 1000
	lastUsedThresholdSecondsDefault = 30
)

var (
	agentHost                      string
	ServerName                     string
	ReplicaIdx                     uint
	SchedulerHost                  string
	SchedulerPort                  int
	SchedulerTlsPort               int
	RcloneHost                     string
	RclonePort                     int
	InferenceHost                  string
	InferenceHttpPort              int
	InferenceGrpcPort              int
	ReverseProxyHttpPort           int
	ReverseProxyGrpcPort           int
	DebugGrpcPort                  int
	MetricsPort                    int
	AgentFolder                    string
	Namespace                      string
	ReplicaConfigStr               string
	InferenceSvcName               string
	ConfigPath                     string
	LogLevel                       string
	ServerType                     string
	memoryBytes                    int
	MemoryBytes64                  uint64
	capabilitiesList               string
	Capabilities                   []string
	OverCommitPercentage           int
	serverTypes                    = [...]string{"mlserver", "triton"}
	TracingConfigPath              string
	EnvoyHost                      string
	EnvoyPort                      int
	DrainerServicePort             int
	ModelInferenceLagThreshold     int
	ModelInferenceDelayMSThreshold int
	ModelInactiveSecondsThreshold  int
	ScalingStatsPeriodSeconds      int
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func UpdateArgs() {
	parseFlags()
	updateFlagsFromEnv()
	setInferenceSvcName()
	updateNamespace()
}

func updateFlagsFromEnv() {
	maybeUpdateOverCommitPercentage()
	maybeUpdateCapabilities()
	maybeUpdateMemoryRequest()
	maybeUpdateInferenceHttpPort()
	maybeUpdateInferenceGrpcPort()
	maybeUpdateReverseProxyHttpPort()
	maybeUpdateReverseProxyGrpcPort()
	maybeUpdateDebugGrpcPort()
	maybeUpdateSchedulerHost()
	maybeUpdateSchedulerPort()
	maybeUpdateSchedulerTlsPort()
	maybeUpdateMetricsPort()
	maybeUpdateServerNameAndIndex()
	maybeUpdateReplicaConfig()
	maybeUpdateLogLevel()
	maybeUpdateServerType()
	maybeUpdateEnvoyHost()
	maybeUpdateEnvoyPort()
	maybeUpdateDrainerPort()
	maybeUpdateModelInferenceLagThreshold()
	maybeUpdateModelInferenceDelayMSThreshold()
	maybeUpdateModelInactiveSecondsThreshold()
	maybeUpdateScalingStatsPeriodSeconds()
}

func maybeUpdateModelInferenceLagThreshold() {
	maybeUpdateFromIntEnv(
		flagModelInferenceLagThreshold,
		envModelInferenceLagThreshold,
		&ModelInferenceLagThreshold,
		"inference lag threshold",
	)
}

func maybeUpdateModelInferenceDelayMSThreshold() {
	maybeUpdateFromIntEnv(
		flagModelInferenceDelayMSThreshold,
		envModelInferenceDelayMSThreshold,
		&ModelInferenceDelayMSThreshold,
		"inference delay threshold in milliseconds",
	)
}

func maybeUpdateModelInactiveSecondsThreshold() {
	maybeUpdateFromIntEnv(
		flagModelInactiveSecondsThreshold,
		envModelInactiveSecondsThreshold,
		&ModelInactiveSecondsThreshold,
		"inactive model seconds threshold",
	)
}

func maybeUpdateScalingStatsPeriodSeconds() {
	maybeUpdateFromIntEnv(
		flagScalingStatsPeriodSeconds,
		envScalingStatsPeriodSeconds,
		&ScalingStatsPeriodSeconds,
		"scaling stats period seconds",
	)
}

func maybeUpdateFromIntEnv(flag string, env string, param *int, tag string) {
	if isFlagPassed(flag) {
		return
	}

	valueFromEnv, found, parsed := getEnvUint(env)
	if !found {
		return
	}
	if !parsed {
		log.Fatalf(
			"Failed to parse %s for %s",
			env, tag)
	}

	log.Infof(
		"Setting %s from env %s with value %d",
		tag,
		env,
		int(valueFromEnv),
	)
	*param = int(valueFromEnv)
}

func maybeUpdateOverCommitPercentage() {
	if isFlagPassed(flagOverCommitPercentage) {
		return
	}

	overCommitPercentageFromEnv, found, parsed := getEnvUint(envOverCommitPercentage)
	if !found {
		return
	}
	if !parsed {
		log.Fatalf(
			"Failed to parse %s for overcommit percentage",
			envOverCommitPercentage)
	}

	log.Infof(
		"Setting overcommit percentage from env %s with value %d",
		envOverCommitPercentage,
		uint32(overCommitPercentageFromEnv),
	)
	OverCommitPercentage = int(overCommitPercentageFromEnv)
}

func maybeUpdateCapabilities() {
	if isFlagPassed(flagCapabilities) {
		return
	}

	capabilitiesFromEnv, found := getEnvString(envCapabilities)
	if !found {
		return
	}

	cs := strings.Split(capabilitiesFromEnv, ",")
	cs = trimStrings(cs)

	log.Infof("Setting capabilities from env %s with value %v", envCapabilities, cs)
	Capabilities = cs
}

func maybeUpdateMemoryRequest() {
	if isFlagPassed(flagMemoryBytes) {
		return
	}

	envMemoryBytes, found, parsed := getEnvUint(envMemoryRequest)
	if !found {
		return
	}
	if !parsed {
		// TODO - don't print value as it'll always be default for type?
		log.Fatalf("Failed to parse %s with value %d", envMemoryRequest, envMemoryBytes)
	}

	log.Infof("Setting memory request from env %s with value %d", envMemoryRequest, envMemoryBytes)
	MemoryBytes64 = uint64(envMemoryBytes)
}

func maybeUpdatePort(flagName string, envName string, port *int) {
	if isFlagPassed(flagName) {
		return
	}

	envPort, found, parsed := getEnvInt(envName)
	if !found {
		return
	}
	if !parsed {
		log.Fatalf("Failed to parse %s with value %d", envName, envPort)
	}

	log.Infof("Setting %s from %s to %d", flagName, envName, envPort)
	*port = envPort
}

func maybeUpdateEnvoyHost() {
	if isFlagPassed(flagEnvoyHost) {
		return
	}

	envoyHostFromEnv, found := getEnvString(envEnvoyHost)
	if !found {
		return
	}

	log.Infof("Setting %s from %s to %s", flagEnvoyHost, envEnvoyHost, envoyHostFromEnv)
	EnvoyHost = envoyHostFromEnv
}

func maybeUpdateEnvoyPort() {
	maybeUpdatePort(flagEnvoyPort, envEnvoyPort, &EnvoyPort)
}

func maybeUpdateDrainerPort() {
	maybeUpdatePort(flagDrainerServicePort, envDrainerServicePort, &DrainerServicePort)
}

func maybeUpdateInferenceHttpPort() {
	maybeUpdatePort(flagInferenceHttpPort, envServerHttpPort, &InferenceHttpPort)
}

func maybeUpdateInferenceGrpcPort() {
	maybeUpdatePort(flagInferenceGrpcPort, envServerGrpcPort, &InferenceGrpcPort)
}

func maybeUpdateReverseProxyHttpPort() {
	maybeUpdatePort(flagReverseProxyHttpPort, envReverseProxyHttpPort, &ReverseProxyHttpPort)
}

func maybeUpdateReverseProxyGrpcPort() {
	maybeUpdatePort(flagReverseProxyGrpcPort, envReverseProxyGrpcPort, &ReverseProxyGrpcPort)
}

func maybeUpdateDebugGrpcPort() {
	maybeUpdatePort(flagDebugGrpcPort, envDebugGrpcPort, &DebugGrpcPort)
}

func maybeUpdateSchedulerPort() {
	maybeUpdatePort(flagSchedulerPlaintxtPort, envSchedulerPort, &SchedulerPort)
}

func maybeUpdateSchedulerTlsPort() {
	maybeUpdatePort(flagSchedulerTlsPort, envSchedulerTlsPort, &SchedulerTlsPort)
}

func maybeUpdateMetricsPort() {
	maybeUpdatePort(flagMetricsPort, envMetricsPort, &MetricsPort)
}

func maybeUpdateSchedulerHost() {
	if isFlagPassed(flagSchedulerHost) {
		return
	}

	schedulerHostFromEnv, found := getEnvString(envSchedulerHost)
	if !found {
		return
	}

	log.Infof("Setting %s from %s to %s", flagSchedulerHost, envSchedulerHost, schedulerHostFromEnv)
	SchedulerHost = schedulerHostFromEnv
}

func maybeUpdateServerNameAndIndex() {
	if isFlagPassed(flagServerName) && isFlagPassed(flagServerIdx) {
		log.Warnf(
			"Using passed in values for server name and server index. Server name %s server index %d",
			ServerName,
			ReplicaIdx,
		)
		return
	}

	setServerNameAndIdxFromPodName()
}

func setServerNameAndIdxFromPodName() {
	log.Infof("Trying to set server name and replica index from pod name")

	podName := os.Getenv(envPodName)
	if podName != "" {
		lastDashIdx := strings.LastIndex(podName, "-")
		if lastDashIdx == -1 {
			log.Infof("Can't decypher pod name to find last dash and index. %s", podName)
		} else {
			serverIdxStr := podName[lastDashIdx+1:]
			var err error
			serverIdx, err := strconv.Atoi(serverIdxStr)
			if err != nil {
				log.
					WithError(err).
					Fatalf("Failed to parse to integer %s with value %s", envPodName, serverIdxStr)
			} else {
				ReplicaIdx = uint(serverIdx)
				ServerName = podName[0:lastDashIdx]

				log.Infof(
					"Got server name and index from %s with value %s. Server name:%s Replica Idx:%d",
					envPodName,
					podName,
					ServerName,
					ReplicaIdx,
				)
			}
		}
	}
}

func maybeUpdateReplicaConfig() {
	if isFlagPassed(flagReplicaConfig) {
		return
	}

	envConfig, found := getEnvString(envReplicaConfig)
	if !found {
		log.Warnf("No value set for %s", flagReplicaConfig)
		return
	}

	log.Infof("Setting %s from %s to %s", flagReplicaConfig, envReplicaConfig, envConfig)
	ReplicaConfigStr = envConfig
}

func maybeUpdateLogLevel() {
	if isFlagPassed(flagLogLevel) {
		return
	}

	envLevel, found := getEnvString(envLogLevel)
	if !found {
		return
	}

	log.Infof("Setting %s from %s to %s", flagLogLevel, envLogLevel, envLevel)
	LogLevel = envLevel
}

func maybeUpdateServerType() {
	if isFlagPassed(flagServerType) {
		return
	}

	envType, found := getEnvString(envServerType)
	if !found {
		log.Warnf("No value set for %s", flagServerType)
		return
	}

	log.Infof("Setting %s from %s to %s", flagServerType, envServerType, envType)
	ServerType = envType
}

func updateNamespace() {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Warn("Using namespace from command line argument")
	} else {
		ns := string(nsBytes)
		log.Infof("Setting namespace from k8s file to %s", ns)
		Namespace = ns
	}
}

func setInferenceSvcName() {
	podName := os.Getenv(envPodName)
	if podName != "" {
		InferenceSvcName = podName
	} else {
		InferenceSvcName = agentHost
	}
	log.Infof("Setting inference svc name to %s", InferenceSvcName)
}
