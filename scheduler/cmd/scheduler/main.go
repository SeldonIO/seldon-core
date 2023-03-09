//   Copyright Steve Sloka 2021
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/dataflow"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/cleaner"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/processor"
	server2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/server"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/server"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	log "github.com/sirupsen/logrus"
)

var (
	envoyPort               uint
	agentPort               uint
	agentMtlsPort           uint
	schedulerPort           uint
	schedulerMtlsPort       uint
	chainerPort             uint
	namespace               string
	pipelineGatewayHost     string
	pipelineGatewayHttpPort int
	pipelineGatewayGrpcPort int
	logLevel                string
	tracingConfigPath       string
	dbPath                  string
	nodeID                  string
	allowPlaintxt           bool //scheduler server
	autoscalingDisabled     bool
	kafkaConfigPath         string
)

func init() {
	rand.Seed(time.Now().UnixNano())

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
	flag.BoolVar(&autoscalingDisabled, "disable-autoscaling", false, "Disable autoscaling feature")
	flag.StringVar(
		&kafkaConfigPath,
		"kafka-config-path",
		"/mnt/config/kafka.json",
		"Path to kafka configuration file",
	)
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

func main() {
	logger := log.New()
	flag.Parse()
	logIntLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", logLevel)
	}
	logger.Infof("Setting log level to %s", logLevel)
	logger.SetLevel(logIntLevel)

	done := make(chan bool, 1)

	namespace = getNamespace()

	// Create event Hub
	eventHub, err := coordinator.NewEventHub(logger)
	if err != nil {
		log.WithError(err).Fatal("Unable to create event hub")
	}
	defer eventHub.Close()
	go makeSignalHandler(logger, done)

	// Start xDS server
	// Create a cache
	xdsCache := cache.NewSnapshotCache(false, cache.IDHash{}, logger)
	ctx := context.Background()
	srv := serverv3.NewServer(ctx, xdsCache, nil)
	xdsServer := server.NewXDSServer(srv, logger)
	err = xdsServer.StartXDSServer(envoyPort)
	if err != nil {
		log.WithError(err).Fatalf("Failed to start envoy xDS server")
	}

	tracer, err := tracing.NewTraceProvider("seldon-scheduler", &tracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	ss := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	ps := pipeline.NewPipelineStore(logger, eventHub, ss)
	es := experiment.NewExperimentServer(logger, eventHub, ss, ps)
	cleaner := cleaner.NewVersionCleaner(ss, logger)

	pipelineGatewayDetails := xdscache.PipelineGatewayDetails{
		Host:     pipelineGatewayHost,
		HttpPort: pipelineGatewayHttpPort,
		GrpcPort: pipelineGatewayGrpcPort,
	}

	_, err = processor.NewIncrementalProcessor(xdsCache, nodeID, logger, ss, es, ps, eventHub, &pipelineGatewayDetails, cleaner)
	if err != nil {
		log.WithError(err).Fatalf("Failed to create incremental processor")
	}

	sched := scheduler.NewSimpleScheduler(
		logger,
		ss,
		scheduler.DefaultSchedulerConfig(ss),
	)
	logger.Infof("Autoscaling service is set to %t", !autoscalingDisabled)
	as := agent.NewAgentServer(logger, ss, sched, eventHub, !autoscalingDisabled)

	dataFlowLoadBalancer := util.NewRingLoadBalancer(1)
	kafkaConfigMap, err := config.NewKafkaConfig(kafkaConfigPath)
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
		err := ps.InitialiseOrRestoreDB(dbPath)
		if err != nil {
			log.WithError(err).Fatalf("Failed to initialise pipeline db at %s", dbPath)
		}
		err = es.InitialiseOrRestoreDB(dbPath)
		if err != nil {
			log.WithError(err).Fatalf("Failed to initialise experiment db at %s", dbPath)
		}
	} else {
		log.Warn("Not running with scheduler local DB")
	}

	s := server2.NewSchedulerServer(logger, ss, es, ps, sched, eventHub)
	err = s.StartGrpcServers(allowPlaintxt, schedulerPort, schedulerMtlsPort)
	if err != nil {
		log.WithError(err).Fatalf("Scheduler start servers error")
	}

	err = as.StartGrpcServer(allowPlaintxt, agentPort, agentMtlsPort)
	if err != nil {
		log.Fatalf("Failed to start agent grpc server %s", err.Error())
	}

	// Wait for completion
	<-done

	log.Info("Shutting down services")
	s.StopSendModelEvents()
	s.StopSendServerEvents()
	s.StopSendExperimentEvents()
	s.StopSendPipelineEvents()
	cs.StopSendPipelineEvents()
	as.StopAgentStreams()
}
