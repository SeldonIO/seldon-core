/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"flag"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

func makeArgs() {
	flag.StringVar(&agentHost, "agent-host", "0.0.0.0", "Agent hostname")
	flag.StringVar(&ServerName, flagServerName, "mlserver", "Server name")
	flag.UintVar(&ReplicaIdx, "server-idx", 0, "Server index")
	flag.StringVar(&SchedulerHost, flagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&SchedulerPort, flagSchedulerPlaintxtPort, defaultSchedulerPort, "Scheduler port")
	flag.IntVar(&SchedulerTlsPort, flagSchedulerTlsPort, defaultSchedulerTlsPort, "Scheduler mTLS port")
	flag.StringVar(&RcloneHost, "rclone-host", "0.0.0.0", "RClone host")
	flag.IntVar(&RclonePort, "rclone-port", defaultRclonePort, "RClone server port")
	flag.StringVar(&InferenceHost, "inference-host", "0.0.0.0", "Inference server host")
	flag.IntVar(&InferenceHttpPort, flagInferenceHttpPort, defaultInferenceHttpPort, "Inference server http port")
	flag.IntVar(&InferenceGrpcPort, flagInferenceGrpcPort, defaultInferenceGrpcPort, "Inference server grpc port")
	flag.IntVar(&ReverseProxyHttpPort, flagReverseProxyHttpPort, util.DefaultReverseProxyHTTPPort, "Reverse proxy http port")
	flag.IntVar(&ReverseProxyGrpcPort, flagReverseProxyGrpcPort, agent.ReverseGRPCProxyPort, "Reverse proxy grpc port")
	flag.IntVar(&DebugGrpcPort, flagDebugGrpcPort, agent.GRPCDebugServicePort, "Debug grpc port")
	flag.IntVar(&MetricsPort, flagMetricsPort, defaultMetricsPort, "Metrics port")
	flag.StringVar(&AgentFolder, "agent-folder", "/mnt/agent", "Model repository folder")
	flag.StringVar(&ReplicaConfigStr, flagReplicaConfig, "", "Replica Json Config")
	flag.StringVar(&Namespace, "namespace", "", "Namespace")
	flag.StringVar(&ConfigPath, "config-path", "/mnt/config", "Path to folder with configuration files. Will assume agent.yaml or agent.json in this folder")
	flag.StringVar(&ServerType, flagServerType, serverTypes[0], "Server type. Default mlserver")
	flag.IntVar(&memoryBytes, flagMemoryBytes, 1000000, "Memory available for server")
	flag.StringVar(&capabilitiesList, flagCapabilities, "sklearn,xgboost", "Server capabilities")
	flag.IntVar(&OverCommitPercentage, flagOverCommitPercentage, 0, "Overcommit memory percentage")
	flag.StringVar(&LogLevel, flagLogLevel, "debug", "Log level - examples: debug, info, error")
	flag.StringVar(&TracingConfigPath, flagTracingConfigPath, "", "Tracing config path")
	flag.StringVar(&EnvoyHost, flagEnvoyHost, defaultEnvoyHost, "Envoy host")
	flag.IntVar(&EnvoyPort, flagEnvoyPort, defaultEnvoyPort, "Envoy port")
	flag.IntVar(&DrainerServicePort, flagDrainerServicePort, defaultDrainerServicePort, "Drainer port")
	flag.IntVar(&ModelInferenceLagThreshold, flagModelInferenceLagThreshold, lagThresholdDefault, "Model inference lag threshold")
	flag.IntVar(&ModelInferenceDelayMSThreshold, flagModelInferenceDelayMSThreshold, delayMSThresholdDefault, "Model inference delay milliseconds threshold")
	flag.IntVar(&ModelInactiveSecondsThreshold, flagModelInactiveSecondsThreshold, lastUsedThresholdSecondsDefault, "Model inactive seconds threshold")
	flag.IntVar(&ScalingStatsPeriodSeconds, flagScalingStatsPeriodSeconds, statsPeriodSecondsDefault, "Scaling stats period seconds")
}

func parseFlags() {
	makeArgs()
	flag.Parse()

	parseMemoryBytes()
	parseCapabilities()
}

func parseMemoryBytes() {
	MemoryBytes64 = uint64(memoryBytes)
}

func parseCapabilities() {
	cs := strings.Split(capabilitiesList, ",")
	cs = trimStrings(cs)
	Capabilities = cs
	log.Infof("Server Capabilities %v", Capabilities)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
