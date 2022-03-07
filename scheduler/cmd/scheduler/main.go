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
	envoyPort     uint
	agentPort     uint
	schedulerPort uint
	namespace     string

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

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

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
	logger.SetLevel(log.DebugLevel)
	flag.Parse()

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

	ss := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	es := experiment.NewExperimentServer(logger, eventHub)
	_ = processor.NewIncrementalProcessor(cache, nodeID, logger, ss, es, eventHub)
	sched := scheduler.NewSimpleScheduler(
		logger,
		ss,
		scheduler.DefaultSchedulerConfig(),
	)
	as := agent.NewAgentServer(logger, ss, sched, eventHub)

	s := server2.NewSchedulerServer(logger, ss, es, sched, eventHub)
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
}
