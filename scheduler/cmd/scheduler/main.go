/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/dataflow"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/cleaner"
	schedulerServer "github.com/seldonio/seldon-core/scheduler/v2/pkg/server"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

var (
	envoyPort                    uint
	agentPort                    uint
	agentMtlsPort                uint
	schedulerPort                uint
	schedulerMtlsPort            uint
	chainerPort                  uint
	namespace                    string
	pipelineGatewayHost          string
	pipelineGatewayHttpPort      int
	pipelineGatewayGrpcPort      int
	logLevel                     string
	tracingConfigPath            string
	dbPath                       string
	nodeID                       string
	allowPlaintxt                bool // scheduler server
	autoscalingModelDisabled     bool
	kafkaConfigPath              string
	schedulerReadyTimeoutSeconds uint
	deletedResourceTTLSeconds    uint
	serverPackingEnabled         bool
	serverPackingPercentage      float64
	accessLogPath                string
	enableAccessLog              bool
	includeSuccessfulRequests    bool
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

	// Whether to enable autoscaling, default is true
	flag.BoolVar(&autoscalingModelDisabled, "disable-model-autoscaling", false, "Disable native model autoscaling feature")

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
	logger := log.New()
	parseFlags()
	logIntLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", logLevel)
	}
	logger.Infof("Setting log level to %s", logLevel)
	logger.SetLevel(logIntLevel)

	logger.Debugf("Scheduler ready timeout is set to %d seconds", schedulerReadyTimeoutSeconds)
	logger.Debugf("Server packing is set to %t", serverPackingEnabled)
	logger.Debugf("Server packing percentage is set to %f", serverPackingPercentage)
	logger.Infof("Autoscaling service is set to %t", !autoscalingModelDisabled)

	done := make(chan bool, 1)

	namespace = getNamespace()

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
	dataFlowLoadBalancer := util.NewRingLoadBalancer(1)
	kafkaConfigMap, err := kafka_config.NewKafkaConfig(kafkaConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load Kafka config")
	}
	cs, err := dataflow.NewChainerServer(logger, eventHub, ps, namespace, dataFlowLoadBalancer, kafkaConfigMap)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start data engine chainer server")
	}
	go func() {
		err := cs.StartGrpcServer(chainerPort)
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

	// scheduler <-> controller grpc
	s := schedulerServer.NewSchedulerServer(
		logger, ss, es, ps, sched, eventHub, sync,
		schedulerServer.SchedulerServerConfig{
			PackThreshold: serverPackingPercentage, // note that if threshold is 0, packing is disabled
		})
	err = s.StartGrpcServers(allowPlaintxt, schedulerPort, schedulerMtlsPort)
	if err != nil {
		log.WithError(err).Fatalf("Failed to start server gRPC servers")
	}

	// scheduler <-> agent  grpc
	as := agent.NewAgentServer(logger, ss, sched, eventHub, !autoscalingModelDisabled)
	err = as.StartGrpcServer(allowPlaintxt, agentPort, agentMtlsPort)
	if err != nil {
		log.WithError(err).Fatalf("Failed to start agent gRPC server")
	}

	// wait for model servers to be ready
	sync.WaitReady()

	// extra wait to allow routes state to get created
	time.Sleep(xDSWaitTimeout)

	// create the processor separately, so it receives all updates
	xdsServer := processor.NewXDSServer(incrementalProcessor, logger)
	err = xdsServer.StartXDSServer(envoyPort)
	if err != nil {
		log.WithError(err).Fatalf("Failed to start envoy xDS server")
	}

	log.Info("Scheduler is ready")

	// Wait for completion
	<-done

	log.Info("Shutting down services")

	s.StopSendModelEvents()
	s.StopSendServerEvents()
	s.StopSendExperimentEvents()
	s.StopSendPipelineEvents()
	s.StopSendControlPlaneEvents()
	cs.StopSendPipelineEvents()
	as.StopAgentStreams()

	log.Info("Shutdown services")
}
