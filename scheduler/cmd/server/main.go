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
	"k8s.io/client-go/util/homedir"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/server"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	log "github.com/sirupsen/logrus"
)

var (
	l log.FieldLogger

	configFilename string
	envoyPort      uint
	schedulerPort       uint
	mode           string
	kubeconfig     string
	namespace      string

	nodeID string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// The envoyPort that this xDS server listens on
	flag.UintVar(&envoyPort, "envoy-port", 9002, "xDS management server port")

	// The envoyPort that this xDS server listens on
	flag.UintVar(&schedulerPort, "scheduler-port", 9004, "scheduler server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
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
	log.SetLevel(log.DebugLevel)
	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, logger)

	// Start xDS server
	go func() {
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, envoyPort)
	}()

	s := scheduler.NewScheduler(cache, nodeID, logger)
	err := s.StartGrpcServer(schedulerPort)
	if err != nil {
		log.WithError(err).Fatalf("Scheduler start server error")
	}



}
