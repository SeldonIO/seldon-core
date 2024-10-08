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
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
)

func TestAgentCliArgsDefault(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                                                    string
		args                                                    []string
		envs                                                    []string
		expectedAgentHost                                       string
		expectedServerName                                      string
		expectedReplicaIdx                                      uint
		expectedSchedulerHost                                   string
		expectedSchedulerPort                                   int
		expectedSchedulerTlsPort                                int
		expectedRcloneHost                                      string
		expectedRclonePort                                      int
		expectedInferenceHost                                   string
		expectedInferenceHttpPort                               int
		expectedInferenceGrpcPort                               int
		expectedReverseProxyHttpPort                            int
		expectedReverseProxyGrpcPort                            int
		expectedDebugGrpcPort                                   int
		expectedMetricsPort                                     int
		expectedAgentFolder                                     string
		expectedReplicaConfigStr                                string
		expectedNamespace                                       string
		expectedConfigPath                                      string
		expectedLogLevel                                        string
		expectedServerType                                      string
		expectedMemoryRequest                                   uint64
		expectedCapabilities                                    []string
		expectedOverCommitPercentage                            int
		expectedEnvoyHost                                       string
		expectedEnvoyPort                                       int
		expectedDrainerPort                                     int
		expectedModelInferenceLagThreshold                      int
		expectedModelInactiveSecondsThreshold                   int
		expectedScalingStatsPeriodSeconds                       int
		expectedMaxElapsedTimeReadySubServiceAfterStartSeconds  int
		expectedMaxElapsedTimeReadySubServiceBeforeStartMinutes int
		expectedPeriodReadySubServiceSeconds                    int
		expectedMaxLoadElapsedTimeMinute                        int
		expectedMaxUnloadElapsedTimeMinute                      int
		expectedMaxLoadRetryCount                               int
		expectedMaxUnloadRetryCount                             int
		expectedUnloadGraceSeconds                              int
	}
	tests := []test{
		{
			name:                                  "default args",
			args:                                  []string{},
			envs:                                  []string{},
			expectedAgentHost:                     "0.0.0.0",
			expectedServerName:                    "mlserver",
			expectedReplicaIdx:                    0,
			expectedSchedulerHost:                 "0.0.0.0",
			expectedSchedulerPort:                 defaultSchedulerPort,
			expectedSchedulerTlsPort:              defaultSchedulerTlsPort,
			expectedRcloneHost:                    "0.0.0.0",
			expectedRclonePort:                    defaultRclonePort,
			expectedInferenceHost:                 "0.0.0.0",
			expectedInferenceHttpPort:             defaultInferenceHttpPort,
			expectedInferenceGrpcPort:             defaultInferenceGrpcPort,
			expectedReverseProxyHttpPort:          9999,
			expectedReverseProxyGrpcPort:          9998,
			expectedDebugGrpcPort:                 agent.GRPCDebugServicePort,
			expectedMetricsPort:                   defaultMetricsPort,
			expectedAgentFolder:                   "/mnt/agent",
			expectedReplicaConfigStr:              "",
			expectedNamespace:                     "",
			expectedConfigPath:                    "/mnt/config",
			expectedLogLevel:                      "debug",
			expectedServerType:                    "mlserver",
			expectedMemoryRequest:                 1000000,
			expectedCapabilities:                  []string{"sklearn", "xgboost"},
			expectedOverCommitPercentage:          0,
			expectedEnvoyHost:                     defaultEnvoyHost,
			expectedEnvoyPort:                     defaultEnvoyPort,
			expectedDrainerPort:                   defaultDrainerServicePort,
			expectedModelInferenceLagThreshold:    lagThresholdDefault,
			expectedModelInactiveSecondsThreshold: lastUsedThresholdSecondsDefault,
			expectedScalingStatsPeriodSeconds:     statsPeriodSecondsDefault,
			expectedMaxElapsedTimeReadySubServiceAfterStartSeconds:  defaultMaxElapsedTimeReadySubServiceAfterStartSeconds,
			expectedMaxElapsedTimeReadySubServiceBeforeStartMinutes: defaultMaxElapsedTimeReadySubServiceBeforeStartMinutes,
			expectedPeriodReadySubServiceSeconds:                    defaultPeriodReadySubServiceSeconds,
			expectedMaxLoadElapsedTimeMinute:                        defaultMaxLoadElapsedTimeMinute,
			expectedMaxUnloadElapsedTimeMinute:                      defaultMaxUnloadElapsedTimeMinute,
			expectedMaxLoadRetryCount:                               defaultMaxLoadRetryCount,
			expectedMaxUnloadRetryCount:                             defaultMaxUnloadRetryCount,
			expectedUnloadGraceSeconds:                              defautUnloadGraceSeconds,
		},
		{
			name: "good args",
			args: []string{
				"--agent-host=1.1.1.1",
				"--server-name=triton",
				"--server-idx=1",
				"--scheduler-host=10.10.10.10",
				"--scheduler-port=10",
				"--scheduler-tls-port=20",
				"--rclone-host=11.11.11.11",
				"--rclone-port=11",
				"--inference-host=12.12.12.12",
				"--inference-http-port=12",
				"--inference-grpc-port=122",
				"--reverse-proxy-http-port=13",
				"--reverse-proxy-grpc-port=133",
				"--debug-grpc-port=14",
				"--metrics-port=15",
				"--agent-folder=/tmp",
				"--replica-config=config",
				"--config-path=/config",
				"--log-level=info",
				"--namespace=namespace",
				"--server-type=triton",
				"--memory-bytes=300",
				"--capabilities=a,b",
				"--over-commit-percentage=10",
				"--envoy-host=2.2.2.2",
				"--envoy-port=2000",
				"--drainer-port=2001",
				"--model-inference-lag-threshold=20",
				"--model-inactive-seconds-threshold=30",
				"--scaling-stats-period-seconds=40",
				"--max-elapsed-time-ready-sub-service-after-start-seconds=30",
				"--max-elapsed-time-ready-sub-service-before-start-minutes=15",
				"--period-ready-sub-service-seconds=60",
				"--max-load-elapsed-time-minutes=120",
				"--max-unload-elapsed-time-minutes=15",
				"--max-load-retry-count=5",
				"--max-unload-retry-count=1",
				"--unload-grace-period-seconds=10",
			},
			envs:                                  []string{},
			expectedAgentHost:                     "1.1.1.1",
			expectedServerName:                    "triton",
			expectedReplicaIdx:                    1,
			expectedSchedulerHost:                 "10.10.10.10",
			expectedSchedulerPort:                 10,
			expectedSchedulerTlsPort:              20,
			expectedRcloneHost:                    "11.11.11.11",
			expectedRclonePort:                    11,
			expectedInferenceHost:                 "12.12.12.12",
			expectedInferenceHttpPort:             12,
			expectedInferenceGrpcPort:             122,
			expectedReverseProxyHttpPort:          13,
			expectedReverseProxyGrpcPort:          133,
			expectedDebugGrpcPort:                 14,
			expectedMetricsPort:                   15,
			expectedAgentFolder:                   "/tmp",
			expectedReplicaConfigStr:              "config",
			expectedNamespace:                     "namespace",
			expectedConfigPath:                    "/config",
			expectedLogLevel:                      "info",
			expectedServerType:                    "triton",
			expectedMemoryRequest:                 300,
			expectedCapabilities:                  []string{"a", "b"},
			expectedOverCommitPercentage:          10,
			expectedEnvoyHost:                     "2.2.2.2",
			expectedEnvoyPort:                     2000,
			expectedDrainerPort:                   2001,
			expectedModelInferenceLagThreshold:    20,
			expectedModelInactiveSecondsThreshold: 30,
			expectedScalingStatsPeriodSeconds:     40,
			expectedMaxElapsedTimeReadySubServiceAfterStartSeconds:  30,
			expectedMaxElapsedTimeReadySubServiceBeforeStartMinutes: 15,
			expectedPeriodReadySubServiceSeconds:                    60,
			expectedMaxLoadElapsedTimeMinute:                        120,
			expectedMaxUnloadElapsedTimeMinute:                      15,
			expectedMaxLoadRetryCount:                               5,
			expectedMaxUnloadRetryCount:                             1,
			expectedUnloadGraceSeconds:                              10,
		},
		{
			name: "good envs",
			args: []string{},
			envs: []string{
				"SELDON_SERVER_TYPE=mlserver",
				"SELDON_SERVER_HTTP_PORT=10",
				"SELDON_SERVER_GRPC_PORT=20",
				"SELDON_REVERSE_PROXY_HTTP_PORT=11",
				"SELDON_REVERSE_PROXY_GRPC_PORT=21",
				"SELDON_DEBUG_GRPC_PORT=30",
				"SELDON_METRICS_PORT=40",
				"SELDON_SCHEDULER_HOST=10.10.10.10",
				"SELDON_SCHEDULER_PORT=100",
				"SELDON_SCHEDULER_TLS_PORT=111",
				"SELDON_REPLICA_CONFIG=config",
				"SELDON_LOG_LEVEL=info",
				"MEMORY_REQUEST=400",
				"SELDON_SERVER_CAPABILITIES=c,d",
				"SELDON_OVERCOMMIT_PERCENTAGE=30",
				"SELDON_ENVOY_HOST=3.3.3.3",
				"SELDON_ENVOY_PORT=3000",
				"SELDON_DRAINER_PORT=3001",
				"SELDON_MODEL_INFERENCE_LAG_THRESHOLD=50",
				"SELDON_MODEL_INACTIVE_SECONDS_THRESHOLD=60",
				"SELDON_SCALING_STATS_PERIOD_SECONDS=70",
				"SELDON_MAX_TIME_READY_SUB_SERVICE_AFTER_START_SECONDS=30",
				"SELDON_MAX_ELAPSED_TIME_READY_SUB_SERVICE_BEFORE_START_MINUTES=15",
				"SELDON_PERIOD_READY_SUB_SERVICE_SECONDS=60",
				"SELDON_MAX_LOAD_ELAPSED_TIME_MINUTES=120",
				"SELDON_MAX_UNLOAD_ELAPSED_TIME_MINUTES=15",
				"SELDON_MAX_LOAD_RETRY_COUNT=5",
				"SELDON_MAX_UNLOAD_RETRY_COUNT=1",
				"SELDON_UNLOAD_GRACE_PERIOD_SECONDS=5",
			},
			expectedAgentHost:                     "0.0.0.0",
			expectedServerName:                    "mlserver",
			expectedReplicaIdx:                    0,
			expectedSchedulerHost:                 "10.10.10.10",
			expectedSchedulerPort:                 100,
			expectedSchedulerTlsPort:              111,
			expectedRcloneHost:                    "0.0.0.0",
			expectedRclonePort:                    defaultRclonePort,
			expectedInferenceHost:                 "0.0.0.0",
			expectedInferenceHttpPort:             10,
			expectedInferenceGrpcPort:             20,
			expectedReverseProxyHttpPort:          11,
			expectedReverseProxyGrpcPort:          21,
			expectedDebugGrpcPort:                 30,
			expectedMetricsPort:                   40,
			expectedAgentFolder:                   "/mnt/agent",
			expectedReplicaConfigStr:              "config",
			expectedNamespace:                     "",
			expectedConfigPath:                    "/mnt/config",
			expectedLogLevel:                      "info",
			expectedServerType:                    "mlserver",
			expectedMemoryRequest:                 400,
			expectedCapabilities:                  []string{"c", "d"},
			expectedOverCommitPercentage:          30,
			expectedEnvoyHost:                     "3.3.3.3",
			expectedEnvoyPort:                     3000,
			expectedDrainerPort:                   3001,
			expectedModelInferenceLagThreshold:    50,
			expectedModelInactiveSecondsThreshold: 60,
			expectedScalingStatsPeriodSeconds:     70,
			expectedMaxElapsedTimeReadySubServiceAfterStartSeconds:  30,
			expectedMaxElapsedTimeReadySubServiceBeforeStartMinutes: 15,
			expectedPeriodReadySubServiceSeconds:                    60,
			expectedMaxLoadElapsedTimeMinute:                        120,
			expectedMaxUnloadElapsedTimeMinute:                      15,
			expectedMaxLoadRetryCount:                               5,
			expectedMaxUnloadRetryCount:                             1,
			expectedUnloadGraceSeconds:                              5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldArgs := os.Args
			oldEnvs := os.Environ()

			// set
			os.Args = []string{"cmd"}
			os.Args = append(os.Args, test.args...)
			for _, e := range test.envs {
				pair := strings.SplitN(e, "=", 2)
				os.Setenv(pair[0], pair[1])
			}

			UpdateArgs()

			g.Expect(agentHost).To(Equal(test.expectedAgentHost))
			g.Expect(ServerName).To(Equal(test.expectedServerName))
			g.Expect(ReplicaIdx).To(Equal(test.expectedReplicaIdx))
			g.Expect(SchedulerHost).To(Equal(test.expectedSchedulerHost))
			g.Expect(SchedulerPort).To(Equal(test.expectedSchedulerPort))
			g.Expect(SchedulerTlsPort).To(Equal(test.expectedSchedulerTlsPort))
			g.Expect(RcloneHost).To(Equal(test.expectedRcloneHost))
			g.Expect(RclonePort).To(Equal(test.expectedRclonePort))
			g.Expect(InferenceHost).To(Equal(test.expectedInferenceHost))
			g.Expect(InferenceHttpPort).To(Equal(test.expectedInferenceHttpPort))
			g.Expect(InferenceGrpcPort).To(Equal(test.expectedInferenceGrpcPort))
			g.Expect(ReverseProxyHttpPort).To(Equal(test.expectedReverseProxyHttpPort))
			g.Expect(ReverseProxyGrpcPort).To(Equal(test.expectedReverseProxyGrpcPort))
			g.Expect(DebugGrpcPort).To(Equal(test.expectedDebugGrpcPort))
			g.Expect(MetricsPort).To(Equal(test.expectedMetricsPort))
			g.Expect(AgentFolder).To(Equal(test.expectedAgentFolder))
			g.Expect(ReplicaConfigStr).To(Equal(test.expectedReplicaConfigStr))
			g.Expect(Namespace).To(Equal(test.expectedNamespace))
			g.Expect(ConfigPath).To(Equal(test.expectedConfigPath))
			g.Expect(LogLevel).To(Equal(test.expectedLogLevel))
			g.Expect(ServerType).To(Equal(test.expectedServerType))
			g.Expect(MemoryBytes64).To(Equal(test.expectedMemoryRequest))
			g.Expect(Capabilities).To(Equal(test.expectedCapabilities))
			g.Expect(OverCommitPercentage).To(Equal(test.expectedOverCommitPercentage))
			g.Expect(EnvoyHost).To(Equal(test.expectedEnvoyHost))
			g.Expect(EnvoyPort).To(Equal(test.expectedEnvoyPort))
			g.Expect(DrainerServicePort).To(Equal(test.expectedDrainerPort))
			g.Expect(ModelInferenceLagThreshold).To(Equal(test.expectedModelInferenceLagThreshold))
			g.Expect(ModelInactiveSecondsThreshold).To(Equal(test.expectedModelInactiveSecondsThreshold))
			g.Expect(ScalingStatsPeriodSeconds).To(Equal(test.expectedScalingStatsPeriodSeconds))
			g.Expect(MaxElapsedTimeReadySubServiceAfterStartSeconds).To(Equal(test.expectedMaxElapsedTimeReadySubServiceAfterStartSeconds))
			g.Expect(MaxElapsedTimeReadySubServiceBeforeStartMinutes).To(Equal(test.expectedMaxElapsedTimeReadySubServiceBeforeStartMinutes))
			g.Expect(PeriodReadySubServiceSeconds).To(Equal(test.expectedPeriodReadySubServiceSeconds))
			g.Expect(MaxLoadElapsedTimeMinute).To(Equal(test.expectedMaxLoadElapsedTimeMinute))
			g.Expect(MaxUnloadElapsedTimeMinute).To(Equal(test.expectedMaxUnloadElapsedTimeMinute))
			g.Expect(MaxLoadRetryCount).To(Equal(test.expectedMaxLoadRetryCount))
			g.Expect(MaxUnloadRetryCount).To(Equal(test.expectedMaxUnloadRetryCount))
			g.Expect(UnloadGraceSeconds).To(Equal(test.expectedUnloadGraceSeconds))
			// reset
			flag.CommandLine = flag.NewFlagSet("cmd", flag.ExitOnError)
			os.Clearenv()
			for _, e := range oldEnvs {
				pair := strings.SplitN(e, "=", 2)
				os.Setenv(pair[0], pair[1])
			}
			os.Args = oldArgs
		})
	}
}
