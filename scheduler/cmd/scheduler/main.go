/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/health"
	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"
	health_probe "github.com/seldonio/seldon-core/scheduler/v2/pkg/health-probe"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/dataflow"
	scaling_config "github.com/seldonio/seldon-core/scheduler/v2/pkg/scaling/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/cleaner"
	schedulerServer "github.com/seldonio/seldon-core/scheduler/v2/pkg/server"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	"github.com/seldonio/seldon-core/scheduler/v2/version"
)

var (
	displayVersion                   bool
	envoyPort                        uint
	agentPort                        uint
	agentMtlsPort                    uint
	schedulerPort                    uint
	schedulerMtlsPort                uint
	chainerPort                      uint
	healthProbePort                  uint
	namespace                        string
	pipelineGatewayHost              string
	pipelineGatewayHttpPort          int
	pipelineGatewayGrpcPort          int
	logLevel                         string
	tracingConfigPath                string
	dbPath                           string
	nodeID                           string
	allowPlaintxt                    bool // scheduler server
	autoscalingModelEnabled          bool
	autoscalingServerEnabled         bool
	kafkaConfigPath                  string
	scalingConfigPath                string
	schedulerReadyTimeoutSeconds     uint
	deletedResourceTTLSeconds        uint
	serverPackingEnabled             bool
	serverPackingPercentage          float64
	accessLogPath                    string
	enableAccessLog                  bool
	includeSuccessfulRequests        bool
	enablePprof                      bool
	pprofPort                        int
	pprofMutexRate                   int
	pprofBlockRate                   int
	retryFailedCreatingPipelinesTick time.Duration
	retryFailedDeletePipelinesTick   time.Duration
	maxRetryFailedPipelines          uint
)

const (
	xDSWaitTimeout = time.Duration(10 * time.Second)

	// percentage of time we try to pack server replicas, i.e. number of server replicas is greater than `MaxNumReplicaHostedModels`
	// this is to be a bit more conservative and not pack all the time as it can lead to
	// increased latency in the case of MMS
	// in the future we should have more metrics to decide whether packing can lead
	// to better performance
	allowPackingPercentageDefault = 0.25
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	flag.BoolVar(&displayVersion, "version", false, "display version and exit")

	// The envoyPort that this xDS server listens on
	flag.UintVar(&envoyPort, "envoy-port", 9002, "xDS management server port")

	// The scheduler port to listen for schedule commands
	flag.UintVar(&schedulerPort, "scheduler-port", 9004, "scheduler server port")

	// The scheduler port to listen for schedule commands using mtls
	flag.UintVar(&schedulerMtlsPort, "scheduler-mtls-port", 9044, "scheduler mtls server port")

	// The agent port to listen for agent subscriptions
	flag.UintVar(&agentPort, "agent-port", 9005, "agent server port")

	// The agent port to listen for schedule commands using mtls
	flag.UintVar(&agentMtlsPort, "agent-mtls-port", 9055, "agent mtls server port")

	// The dataflow port to listen for data flow agents
	flag.UintVar(&chainerPort, "dataflow-port", 9008, "dataflow server port")

	// The port the k8s HTTP health probe listens on
	flag.UintVar(&healthProbePort, "health-probe-port", 9999, "http health probe server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	// Kubernetes namespace
	flag.StringVar(&namespace, "namespace", "", "Namespace")

	flag.StringVar(&pipelineGatewayHost, "pipeline-gateway-host", "0.0.0.0", "Pipeline gateway server host")
	flag.IntVar(&pipelineGatewayHttpPort, "pipeline-gateway-http-port", 9010, "Pipeline gateway server http port")
	flag.IntVar(&pipelineGatewayGrpcPort, "pipeline-gateway-grpc-port", 9011, "Pipeline gateway server grpc port")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level - examples: debug, info, error")
	flag.StringVar(&tracingConfigPath, "tracing-config-path", "", "Tracing config path")
	flag.StringVar(&dbPath, "db-path", "", "State Db path")

	// Allow plaintext servers
	flag.BoolVar(&allowPlaintxt, "allow-plaintxt", true, "Allow plain text scheduler server")

	// Autoscaling
	// Scaling config path
	flag.StringVar(
		&scalingConfigPath,
		"scaling-config-path",
		"/mnt/config/scaling.json",
		"Path to scaling configuration file",
	)
	flag.BoolVar(&autoscalingModelEnabled, "enable-model-autoscaling", false, "Enable native model autoscaling feature")
	flag.BoolVar(&autoscalingServerEnabled, "enable-server-autoscaling", true, "Enable native server autoscaling feature")

	// Kafka config path
	flag.StringVar(
		&kafkaConfigPath,
		"kafka-config-path",
		"/mnt/config/kafka.json",
		"Path to kafka configuration file",
	)

	// Timeout for scheduler to be ready
	flag.UintVar(&schedulerReadyTimeoutSeconds, "scheduler-ready-timeout-seconds", 300, "Timeout for scheduler to be ready")

	// This TTL is set in badger DB
	flag.UintVar(&deletedResourceTTLSeconds, "deleted-resource-ttl-seconds", 86400, "TTL for deleted experiments and pipelines (in seconds)")

	// Server packing
	flag.BoolVar(&serverPackingEnabled, "server-packing-enabled", false, "Enable server packing")
	flag.Float64Var(&serverPackingPercentage, "server-packing-percentage", allowPackingPercentageDefault, "Percentage of time we try to pack server replicas")

	// Envoy access log config
	flag.StringVar(&accessLogPath, "envoy-accesslog-path", "/tmp/envoy-accesslog.txt", "Envoy access log path")
	flag.BoolVar(&enableAccessLog, "enable-envoy-accesslog", true, "Enable Envoy access log")
	flag.BoolVar(&includeSuccessfulRequests, "include-successful-requests-envoy-accesslog", false, "Include successful requests in Envoy access log")
	flag.BoolVar(&enablePprof, "enable-pprof", false, "Enables pprof on localhost - do not use in production, will affect performance")
	flag.IntVar(&pprofPort, "pprof-port", 6060, "pprof HTTP server port")
	flag.IntVar(&pprofBlockRate, "pprof-block-rate", 0, "pprof block rate")
	flag.IntVar(&pprofMutexRate, "pprof-mutex-rate", 0, "pprof mutex rate")

	// frequency to retry creating/deleting pipelines which failed to create/delete
	flag.DurationVar(&retryFailedCreatingPipelinesTick, "retry-creating-failed-pipelines-tick", time.Minute, "tick interval for re-attempting to create pipelines which failed to create")
	flag.DurationVar(&retryFailedDeletePipelinesTick, "retry-deleting-failed-pipelines-tick", time.Minute, "tick interval for re-attempting to delete pipelines which failed to terminate")
	flag.UintVar(&maxRetryFailedPipelines, "max-retry-failed-pipelines", 10, "max number of retry attempts to create/terminate pipelines which failed to create/terminate")
}

func getNamespace() string {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Warn("Using namespace from command line argument")
		return namespace
	}
	ns := string(nsBytes)
	log.Info("Namespace is ", ns)
	return ns
}

func makeSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
}

func parseFlags() {
	flag.Parse()
	if !serverPackingEnabled {
		// zero packing percentage == server packing is disabled
		serverPackingPercentage = 0
	}
}

func main() {
	parseFlags()
	logger := log.New()

	if displayVersion {
		logger.Infof("Version %s", version.Tag)
		os.Exit(0)
	}

	logIntLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", logLevel)
	}
	logger.Infof("Version %s", version.Tag)
	logger.Infof("Setting log level to %s", logLevel)
	logger.SetLevel(logIntLevel)

	logger.Debugf("Scheduler ready timeout is set to %d seconds", schedulerReadyTimeoutSeconds)
	logger.Debugf("Server packing is set to %t", serverPackingEnabled)
	logger.Debugf("Server packing percentage is set to %f", serverPackingPercentage)
	logger.Infof("Autoscaling (native) service is set to Model: %t and Server: %t", autoscalingModelEnabled, autoscalingServerEnabled)
	done := make(chan bool, 1)

	namespace = getNamespace()

	tlsOptions, err := tls.CreateControlPlaneTLSOptions(
		tls.Prefix(tls.EnvSecurityPrefixControlPlaneServer),
		tls.ValidationPrefix(tls.EnvSecurityPrefixControlPlaneClient))
	if err != nil {
		logger.WithError(err).Fatal("Failed to create TLS Options")
	}

	probesConfig := gRPCHealthProbes{}
	probesConfig.probes = append(probesConfig.probes, probe{
		port:       int(chainerPort),
		plaintText: true,
	})
	if allowPlaintxt {
		probesConfig.probes = append(probesConfig.probes,
			probe{port: int(agentPort), plaintText: true},
			probe{port: int(schedulerPort), plaintText: true})
	}
	if tlsOptions.Cert != nil {
		probesConfig.probes = append(probesConfig.probes,
			probe{port: int(agentMtlsPort), plaintText: false},
			probe{port: int(schedulerMtlsPort), plaintText: false})
	}

	httpServer, err := initHealthProbe(tlsOptions, logger, probesConfig, int(healthProbePort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to start health server")
	}

	if enablePprof {
		logger.Info("Starting pprof server")
		startPprofServer(pprofPort, pprofBlockRate, pprofMutexRate, logger)
	}

	logger.WithField("port", healthProbePort).Info("Started HTTP health server")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.WithError(err).Warn("Failed to shutdown health server")
		}
		cancel()
	}()

	// Create event Hub
	eventHub, err := coordinator.NewEventHub(logger)
	if err != nil {
		log.WithError(err).Fatal("Unable to create event hub")
	}
	defer eventHub.Close()
	go makeSignalHandler(logger, done)

	tracer, err := tracing.NewTraceProvider("seldon-scheduler", &tracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	// Create stores
	ss := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	ps := pipeline.NewPipelineStore(logger, eventHub, ss)
	es := experiment.NewExperimentServer(logger, eventHub, ss, ps)
	cleaner := cleaner.NewVersionCleaner(ss, logger)

	pipelineGatewayDetails := xdscache.PipelineGatewayDetails{
		Host:     pipelineGatewayHost,
		HttpPort: pipelineGatewayHttpPort,
		GrpcPort: pipelineGatewayGrpcPort,
	}

	// Create envoy incremental processor
	incrementalProcessor, err := processor.NewIncrementalProcessor(
		nodeID, logger, ss, es, ps, eventHub, &pipelineGatewayDetails, cleaner, &xdscache.EnvoyConfig{
			AccessLogPath: accessLogPath, EnableAccessLog: enableAccessLog, IncludeSuccessfulRequests: includeSuccessfulRequests})
	if err != nil {
		log.WithError(err).Fatalf("Failed to create incremental processor")
	}

	// scheduler <-> dataflow grpc
	kafkaConfigMap, err := kafka_config.NewKafkaConfig(kafkaConfigPath, logLevel)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load Kafka config")
	}
	scalingConfigHdl, err := scaling_config.NewScalingConfigHandler(scalingConfigPath, namespace, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to load Scaling config from %s", scalingConfigPath)
	}
	defer func() {
		_ = scalingConfigHdl.Close()
		logger.Info("Closed scheduler scaling config watcher")
	}()

	maxShardCountMultiplier := scalingConfigHdl.GetConfiguration().Pipelines.MaxShardCountMultiplier

	dataFlowLoadBalancer := util.NewRingLoadBalancer(maxShardCountMultiplier)
	log.Info("Using ring load balancer for data flow with numPartitions: ", maxShardCountMultiplier)

	cs, err := dataflow.NewChainerServer(logger, eventHub, ps, namespace, dataFlowLoadBalancer, kafkaConfigMap, scalingConfigHdl)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start data engine chainer server")
	}
	defer cs.Stop()

	ctx, stopPipelinePollers := context.WithCancel(context.Background())
	defer stopPipelinePollers()
	go func() {
		err := cs.StartGrpcServer(ctx, retryFailedCreatingPipelinesTick, retryFailedDeletePipelinesTick, maxRetryFailedPipelines, chainerPort)
		if err != nil {
			log.WithError(err).Fatalf("Chainer server start error")
		}
	}()

	// Load pipelines and experiments from DB
	// Do here after other services created so eventHub events will be handled on pipeline/experiment load
	// If we start earlier events will be sent but not received by services that start listening "late" to eventHub
	if dbPath != "" {
		err := ps.InitialiseOrRestoreDB(dbPath, deletedResourceTTLSeconds)
		if err != nil {
			log.WithError(err).Fatalf("Failed to initialise pipeline db at %s", dbPath)
		}
		err = es.InitialiseOrRestoreDB(dbPath, deletedResourceTTLSeconds)
		if err != nil {
			log.WithError(err).Fatalf("Failed to initialise experiment db at %s", dbPath)
		}
	} else {
		log.Warn("Not running with scheduler local DB")
	}

	// Setup synchroniser
	var sync synchroniser.Synchroniser

	if namespace == "" {
		// running outside k8s
		sync = synchroniser.NewSimpleSynchroniser(time.Duration(schedulerReadyTimeoutSeconds) * time.Second)
	} else {
		sync = synchroniser.NewServerBasedSynchroniser(eventHub, logger, time.Duration(schedulerReadyTimeoutSeconds)*time.Second)
	}

	// scheduler scheduling models service
	sched := scheduler.NewSimpleScheduler(
		logger,
		ss,
		scheduler.DefaultSchedulerConfig(ss),
		sync,
		eventHub,
	)

	// scheduler <-> controller and {pipeline,model-gw} grpc
	modelGwLoadBalancer := util.NewRingLoadBalancer(maxShardCountMultiplier)
	pipelineGWLoadBalancer := util.NewRingLoadBalancer(maxShardCountMultiplier)
	s := schedulerServer.NewSchedulerServer(
		logger, ss, es, ps, sched, eventHub, sync,
		schedulerServer.SchedulerServerConfig{
			PackThreshold:            serverPackingPercentage, // note that if threshold is 0, packing is disabled
			AutoScalingServerEnabled: autoscalingServerEnabled,
		},
		namespace,
		kafkaConfigMap.ConsumerGroupIdPrefix,
		modelGwLoadBalancer,
		pipelineGWLoadBalancer,
		scalingConfigHdl,
		*tlsOptions,
	)
	defer s.Stop()

	err = s.StartGrpcServers(ctx, allowPlaintxt, schedulerPort, schedulerMtlsPort, retryFailedCreatingPipelinesTick,
		retryFailedDeletePipelinesTick, maxRetryFailedPipelines)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start server gRPC servers")
	}

	// scheduler <-> agent  grpc
	// auto-scaling has been disabled for time being, until issue is fixed for scaling models back up
	autoscalingModelEnabled = false
	as := agent.NewAgentServer(logger, ss, sched, eventHub, autoscalingModelEnabled, *tlsOptions)
	err = as.StartGrpcServer(allowPlaintxt, agentPort, agentMtlsPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start agent gRPC server")
	}

	// wait for model servers to be ready
	sync.WaitReady()
	logger.Info("Inference servers ready")

	// extra wait to allow routes state to get created
	time.Sleep(xDSWaitTimeout)

	// create the processor separately, so it receives all updates
	xdsServer := processor.NewXDSServer(incrementalProcessor, logger)
	err = xdsServer.StartXDSServer(envoyPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start envoy xDS server")
	}

	log.Info("Scheduler is ready")

	// Wait for completion
	<-done

	log.Info("Shutting down services")

	s.StopSendModelEvents()
	s.StopSendServerEvents()
	s.StopSendExperimentEvents()
	s.StopSendPipelineEvents()
	stopPipelinePollers()
	s.StopSendControlPlaneEvents()
	as.StopAgentStreams()

	log.Info("All services have shut down cleanly")
}

func startPprofServer(port int, blockRate, mutexRate int, log *log.Logger) {
	if blockRate > 0 {
		log.Warn("Block rate > 0 - performance will be affected")
	}
	if mutexRate > 0 {
		log.Warn("Mutex rate > 0 - performance will be affected")
	}
	runtime.SetBlockProfileRate(blockRate)
	runtime.SetMutexProfileFraction(mutexRate)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Error("Failed running pprof server")
		}
	}()
}

type probe struct {
	port       int
	plaintText bool
}

type gRPCHealthProbes struct {
	probes []probe
}

func initHealthProbe(tlsOptions *tls.TLSOptions, log *log.Logger, config gRPCHealthProbes, healthSrvPort int) (*health_probe.HTTPServer, error) {
	manager := health_probe.NewManager()
	for _, probe := range config.probes {
		var cert *tls.CertificateStore
		if !probe.plaintText {
			cert = tlsOptions.Cert
		}
		if err := createGRPCHealthProbe(cert, probe.port, manager); err != nil {
			return nil, fmt.Errorf("failed to create health probe: %w", err)
		}
	}

	server := health_probe.NewHTTPServer(healthSrvPort, manager, log)
	go func() {
		if err := server.Start(); err != nil {
			log.WithError(err).Fatal("Failed to running health probe HTTP server")
		}
	}()

	return server, nil
}

func createGRPCHealthProbe(cert *tls.CertificateStore, port int, manager health_probe.Manager) error {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.DefaultConfig,
	}), grpc.WithKeepaliveParams(util.GetClientKeepAliveParameters()))

	if cert != nil {
		opts = append(opts, grpc.WithTransportCredentials(cert.CreateClientTransportCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// note this will not attempt connection handshake until req is sent
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port), opts...)
	if err != nil {
		return fmt.Errorf("error creating gRPC connection: %v", err)
	}
	gRPCClient := health.NewHealthCheckServiceClient(conn)
	manager.AddCheck(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		resp, err := gRPCClient.HealthCheck(ctx, &health.HealthCheckRequest{})
		if err != nil {
			return fmt.Errorf("gRPC health check error: %v", err)
		}
		if !resp.Ok {
			return fmt.Errorf("non-health gRPC response")
		}
		return nil
	}, health_probe.ProbeReadiness, health_probe.ProbeLiveness)

	return nil
}
