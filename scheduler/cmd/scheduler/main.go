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
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/util"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/dataflow"
	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/cleaner"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/processor"
	server2 "github.com/seldonio/seldon-core/scheduler/pkg/server"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/server"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	log "github.com/sirupsen/logrus"
)

var (
	envoyPort               uint
	agentPort               uint
	schedulerPort           uint
	chainerPort             uint
	namespace               string
	pipelineGatewayHost     string
	pipelineGatewayHttpPort int
	pipelineGatewayGrpcPort int
	logLevel                string
	pipelineDbPath          string

	nodeID string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// The envoyPort that this xDS server listens on
	flag.UintVar(&envoyPort, "envoy-port", 9002, "xDS management server port")

	// The scheduler port to listen for schedule commands
	flag.UintVar(&schedulerPort, "scheduler-port", 9004, "scheduler server port")

	// The agent port to listen for agent subscriptions
	flag.UintVar(&agentPort, "agent-port", 9005, "agent server port")

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
	flag.StringVar(&pipelineDbPath, "pipeline-db-path", "", "Pipeline state Db")
}

func getNamespace() string {
	nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Warn("Using namespace from command line argument")
		return namespace
	}
	ns := string(nsBytes)
	log.Info("Namespace is ", ns)
	return ns
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

	namespace = getNamespace()

	// Create event Hub
	eventHub, err := coordinator.NewEventHub(logger)
	if err != nil {
		log.WithError(err).Fatal("Unable to create event hub")
	}
	defer eventHub.Close()

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, logger)

	// Start xDS server
	go func() {
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, envoyPort)
	}()

	ps := pipeline.NewPipelineStore(logger, eventHub)
	if pipelineDbPath != "" {
		logger.Infof("Opening db at %s", pipelineDbPath)
		err := ps.InitialiseDB(pipelineDbPath)
		if err != nil {
			log.WithError(err).Fatalf("Failed to initialise pipeline db at %s", pipelineDbPath)
		}
	} else {
		log.Warn("Not running with scheduler DB")
	}

	ss := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	es := experiment.NewExperimentServer(logger, eventHub, ss)
	pipelineGatewayDetails := xdscache.PipelineGatewayDetails{
		Host:     pipelineGatewayHost,
		HttpPort: pipelineGatewayHttpPort,
		GrpcPort: pipelineGatewayGrpcPort,
	}
	_ = processor.NewIncrementalProcessor(cache, nodeID, logger, ss, es, ps, eventHub, &pipelineGatewayDetails)
	sched := scheduler.NewSimpleScheduler(
		logger,
		ss,
		scheduler.DefaultSchedulerConfig(),
	)
	as := agent.NewAgentServer(logger, ss, sched, eventHub)

	dataFlowLoadBalancer := util.NewRingLoadBalancer(1)
	cs := dataflow.NewChainerServer(logger, eventHub, ps, namespace, dataFlowLoadBalancer)
	go func() {
		err := cs.StartGrpcServer(chainerPort)
		if err != nil {
			log.WithError(err).Fatalf("Chainer server start error")
		}
	}()

	s := server2.NewSchedulerServer(logger, ss, es, ps, sched, eventHub)
	go func() {
		err := s.StartGrpcServer(schedulerPort)
		if err != nil {
			log.WithError(err).Fatalf("Scheduler start server error")
		}
	}()

	_ = cleaner.NewVersionCleaner(ss, logger, eventHub)

	// TODO - it's subtle (and thus fragile) to use the fact that this method
	// is blocking to await shutdown.
	// We should instead use a done channel (as elsewhere) and defer stops/shutdowns
	// OR use a wait-group as defers runs sequentially.
	err = as.StartGrpcServer(agentPort)
	if err != nil {
		log.Fatalf("Failed to start agent grpc server %s", err.Error())
	}

	s.StopSendModelEvents()
	s.StopSendServerEvents()
	s.StopSendExperimentEvents()
	s.StopSendPipelineEvents()
	cs.StopSendPipelineEvents()
}
