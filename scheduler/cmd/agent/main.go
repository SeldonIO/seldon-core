/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"

	agent2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/drainservice"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/mlserver"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/triton"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/rclone"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s"
	"k8s.io/client-go/kubernetes"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/cmd/agent/cli"
)

const (
	// the max time to wait for a subservice ready after client start, i.e. during operation
	maxElapsedTimeReadySubServiceAfterStart = 30 * time.Second
	// the max time to wait for a subservice ready before client start, i.e. during startup
	maxElapsedTimeReadySubServiceBeforeStart = 15 * time.Minute // 15 mins is the default MaxElapsedTime
	// period for subservice ready "cron"
	periodReadySubService = 60 * time.Second
)

func makeDirs() (string, string, error) {
	modelRepositoryDir := filepath.Join(cli.AgentFolder, "models")
	rcloneRepositoryDir := filepath.Join(cli.AgentFolder, "rclone")
	err := os.MkdirAll(modelRepositoryDir, fs.ModePerm)
	if err != nil {
		return modelRepositoryDir, rcloneRepositoryDir, err
	}
	_ = os.Chmod(modelRepositoryDir, fs.ModePerm)
	err = os.MkdirAll(rcloneRepositoryDir, fs.ModePerm)
	if err != nil {
		return modelRepositoryDir, rcloneRepositoryDir, err
	}
	_ = os.Chmod(rcloneRepositoryDir, fs.ModePerm)
	return modelRepositoryDir, rcloneRepositoryDir, nil
}

func getRepositoryHandler(logger log.FieldLogger) repository.ModelRepositoryHandler {
	switch cli.ServerType {
	case "mlserver":
		logger.Infof("Creating MLServer repository handler")
		return mlserver.NewMLServerRepositoryHandler(logger)
	case "triton":
		logger.Infof("Creating Triton repository handler")
		return triton.NewTritonRepositoryHandler(logger)
	default:
		logger.Infof("Using default as no server type requested - creating MLServer repository handler")
		return mlserver.NewMLServerRepositoryHandler(logger)
	}
}

func createReplicaConfig() *agent2.ReplicaConfig {
	var rc *agent2.ReplicaConfig
	if cli.ReplicaConfigStr != "" {
		var err error
		rc, err = agent.ParseReplicaConfig(cli.ReplicaConfigStr)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse replica config %s", cli.ReplicaConfigStr)
		}
		log.Infof("Created replicaConfig from command line")
	} else {
		rc = &agent2.ReplicaConfig{
			InferenceSvc:         cli.InferenceSvcName,
			InferenceHttpPort:    int32(cli.InferenceHttpPort),
			InferenceGrpcPort:    int32(cli.InferenceGrpcPort),
			MemoryBytes:          cli.MemoryBytes64,
			Capabilities:         cli.Capabilities,
			OverCommitPercentage: uint32(cli.OverCommitPercentage),
		}
		log.Infof("Created replicaConfig from environment")
	}
	//Point to proxy always in replica config
	rc.InferenceHttpPort = int32(cli.ReverseProxyHttpPort)
	rc.InferenceGrpcPort = int32(cli.ReverseProxyGrpcPort)
	log.Infof("replicaConfig %+v", rc)
	return rc
}

func runningInsideK8s() bool {
	return cli.Namespace != ""
}

func makeTermSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
}

func main() {
	logger := log.New()

	cli.UpdateArgs()

	logIntLevel, err := log.ParseLevel(cli.LogLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", cli.LogLevel)
	}
	logger.Infof("Setting log level to %s", cli.LogLevel)
	logger.SetLevel(logIntLevel)

	// Make required folders
	//TODO handle via initContainer?
	modelRepositoryDir, rcloneRepositoryDir, err := makeDirs()
	if err != nil {
		logger.
			WithError(err).
			Fatalf("Failed to create required folders %s and %s", modelRepositoryDir, rcloneRepositoryDir)
	}
	log.Infof("Model repository dir %s, Rclone repository dir %s ", modelRepositoryDir, rcloneRepositoryDir)

	done := make(chan bool, 1)

	go makeTermSignalHandler(logger, done)

	var clientset kubernetes.Interface
	if runningInsideK8s() {
		clientset, err = k8s.CreateClientset()
		if err != nil { //TODO change to Error from Fatal?
			logger.WithError(err).Fatal("Failed to create kubernetes clientset")
		}
	}

	tracer, err := tracing.NewTraceProvider("seldon-agent", &cli.TracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	// Start Agent configuration handler
	agentConfigHandler, err := config.NewAgentConfigHandler(cli.ConfigPath, cli.Namespace, logger, clientset)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create agent config handler")
	}
	defer func() {
		_ = agentConfigHandler.Close()
		logger.Info("Closed agent handler")
	}()

	// Create Rclone client
	rcloneClient := rclone.NewRCloneClient(cli.RcloneHost, cli.RclonePort, rcloneRepositoryDir, logger, cli.Namespace)

	// Create Model Repository
	modelRepository := repository.NewModelRepository(
		logger,
		rcloneClient,
		modelRepositoryDir,
		getRepositoryHandler(logger),
		cli.EnvoyHost,
		cli.EnvoyPort,
	)

	// Create V2 Protocol Handler
	v2Client := agent.NewV2Client(cli.InferenceHost, cli.InferenceGrpcPort, logger, true)

	promMetrics, err := metrics.NewPrometheusModelMetrics(cli.ServerName, cli.ReplicaIdx, logger)
	if err != nil {
		logger.WithError(err).Fatal("Can't create prometheus metrics")
	}
	go func() {
		err := promMetrics.Start(cli.MetricsPort)
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.WithError(err).Fatal("Can't start metrics server")
		close(done)
	}()
	defer func() { _ = promMetrics.Stop() }()

	modelLagStatsWrapper := modelscaling.ModelScalingStatsWrapper{
		Stats:     modelscaling.NewModelReplicaLagsKeeper(),
		Operator:  interfaces.Gte,
		Threshold: uint(cli.ModelInferenceLagThreshold),
		Reset:     true,
		EventType: modelscaling.ScaleUpEvent,
	}
	modelLastUsedStatsWrapper := modelscaling.ModelScalingStatsWrapper{
		Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
		Operator:  interfaces.Gte,
		Threshold: uint(cli.ModelInactiveSecondsThreshold),
		Reset:     false,
		EventType: modelscaling.ScaleDownEvent,
	}
	modelScalingStatsCollector := modelscaling.NewDataPlaneStatsCollector(
		modelLagStatsWrapper.Stats,
		modelLastUsedStatsWrapper.Stats,
	)

	rpHTTP := agent.NewReverseHTTPProxy(
		logger,
		cli.InferenceHost,
		uint(cli.InferenceHttpPort),
		uint(cli.ReverseProxyHttpPort),
		promMetrics,
		modelScalingStatsCollector,
	)
	defer func() { _ = rpHTTP.Stop() }()

	rpGRPC := agent.NewReverseGRPCProxy(
		promMetrics,
		logger,
		cli.InferenceHost,
		uint(cli.InferenceGrpcPort),
		uint(cli.ReverseProxyGrpcPort),
		modelScalingStatsCollector,
	)
	defer func() { _ = rpGRPC.Stop() }()

	agentDebugService := agent.NewAgentDebug(logger, uint(cli.DebugGrpcPort))
	defer func() { _ = agentDebugService.Stop() }()

	modelScalingService := modelscaling.NewStatsAnalyserService(
		[]modelscaling.ModelScalingStatsWrapper{modelLagStatsWrapper, modelLastUsedStatsWrapper}, logger, uint(cli.ScalingStatsPeriodSeconds))
	defer func() { _ = modelScalingService.Stop() }()

	drainerService := drainservice.NewDrainerService(
		logger, uint(cli.DrainerServicePort))
	defer func() { _ = drainerService.Stop() }()

	// Create Agent
	client := agent.NewClient(
		agent.NewClientSettings(
			cli.ServerName,
			uint32(cli.ReplicaIdx),
			cli.SchedulerHost,
			cli.SchedulerPort,
			cli.SchedulerTlsPort,
			periodReadySubService,
			maxElapsedTimeReadySubServiceBeforeStart,
			maxElapsedTimeReadySubServiceAfterStart,
		),
		logger,
		modelRepository,
		v2Client,
		createReplicaConfig(),
		cli.Namespace,
		rpHTTP,
		rpGRPC,
		agentDebugService,
		modelScalingService,
		drainerService,
		promMetrics,
	)

	// Wait for required services to be ready
	err = client.WaitReadySubServices(true)
	if err != nil {
		logger.WithError(err).Fatal("Failed to wait for all agent dependent services to be ready")
		close(done)
	}

	// Now we are ready start config listener
	err = rcloneClient.StartConfigListener(agentConfigHandler)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialise rclone config listener")
		close(done)
	}

	// Start client grpc server
	go func() {
		err = client.Start()
		if err != nil {
			logger.WithError(err).Error("Failed to initialise client")
		}
		close(done)
	}()
	defer func() { client.Stop() }()

	// Wait for completion
	<-done
	logger.Warning("Agent shutting down")
}