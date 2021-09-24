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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/seldonio/seldon-core/scheduler/internal/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/internal/envoy/server"
	"github.com/seldonio/seldon-core/scheduler/internal/envoy/watcher"
	log "github.com/sirupsen/logrus"
)

var (
	l log.FieldLogger

	configFilename string
	port           uint
	basePort       uint
	mode           string
	kubeconfig     string
	namespace      string

	nodeID string
)

func init() {
	l = log.New()
	log.SetLevel(log.DebugLevel)

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 9002, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	// Define the directory to watch for Envoy configuration files
	flag.StringVar(&configFilename, "config", "", "full path to directory to watch for files")

	// Namespace - used for configmap watcher
	flag.StringVar(&namespace, "namespace", "seldon-mesh", "namespace we are running in")

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
	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Create a processor
	proc := processor.NewSeldonProcessor(
		cache, nodeID, log.WithField("context", "processor"))

	// Notify channel for file system events
	notifyCh := make(chan watcher.NotifyMessage)

	if configFilename != "" {
		log.Info("Starting from local config file")
		// Create initial snapshot from file
		yamlFile, err := ioutil.ReadFile(configFilename)
		if err != nil {
			panic("Failed to load file")
		}
		proc.ProcessFile(watcher.NotifyMessage{
			Operation: watcher.Create,
			Contents:  yamlFile,
		})

		go func() {
			// WatchFile for file changes
			watcher.WatchFile(configFilename, notifyCh)
		}()
	} else {
		log.Info("Starting from configmap")
		var clientCfg *rest.Config
		if _, err := os.Stat(kubeconfig); err == nil {
			// use the current context in kubeconfig
			log.Info("Using user kubeconfig")
			clientCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				panic(err.Error())
			}
		} else {
			log.Info("Using in cluster kubeconfig")
			// Get things set up for watching - we need a valid k8s client
			clientCfg, err = rest.InClusterConfig()
			if err != nil {
				panic("Unable to get our client configuration")
			}
		}
		clientset, err := kubernetes.NewForConfig(clientCfg)
		if err != nil {
			panic("Unable to create our clientset")
		}
		go func() {
			// WatchFile for file changes
			watcher.WatchConfigmap(clientset, getNamespace(), notifyCh)
		}()
	}


	go func() {
		// Run the xDS server
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, port)
	}()

	for {
		select {
		case msg := <-notifyCh:
			proc.ProcessFile(msg)
		}
	}
}
