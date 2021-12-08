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

	"github.com/seldonio/seldon-core/scheduler/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/filters"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/sorters"
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

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, logger)

	// Start xDS server
	go func() {
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, envoyPort)
	}()

	ss := store.NewMemoryStore(logger, store.NewLocalSchedulerStore())
	es := processor.NewIncrementalProcessor(cache, nodeID, logger, ss)
	sched := scheduler.NewSimpleScheduler(logger,
		ss,
		[]scheduler.ServerFilter{filters.SharingServerFilter{}},
		[]scheduler.ReplicaFilter{filters.RequirementsReplicaFilter{}, filters.AvailableMemoryFilter{}},
		[]sorters.ServerSorter{},
		[]sorters.ReplicaSorter{sorters.ModelAlreadyLoadedSorter{}})
	as := agent.NewAgentServer(logger, ss, es, sched)

	go as.ListenForSyncs() // Start agent syncs
	go es.ListenForSyncs() // Start envoy syncs

	s := server2.NewSchedulerServer(logger, ss, sched, as)
	go func() {
		err := s.StartGrpcServer(schedulerPort)
		if err != nil {
			log.WithError(err).Fatalf("Scheduler start server error")
		}
	}()

	err := as.StartGrpcServer(agentPort)
	if err != nil {
		log.Fatalf("Failed to start agent grpc server %s", err.Error())
	}

	as.StopAgentSync()
	es.StopEnvoySync()
}
