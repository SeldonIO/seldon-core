/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package drainservice

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
)

const (
	terminateEndpoint = "/terminate"
	eventWindowMs     = 100
	eventsTarget      = 3 // e.g. the number of containers in a pod
)

type DrainerService struct {
	server      *http.Server
	port        uint
	logger      log.FieldLogger
	serverReady bool
	// mutex to guard changes to `serverReady`
	muServerReady sync.RWMutex
	triggered     bool
	// mutex to guard changes to `triggered`
	muTriggered sync.Mutex
	// wait group to block consumers of the DrainerService until a call to /terminate has occurred.
	// this is effectively triggering downstream logic in agent
	triggeredWg *sync.WaitGroup
	// wait group to block until the logic of draining models (rescheduling) has finished.
	// this is effectively including agent and scheduler related logic.
	// at this state we should be confident that this server replica (agent) can go down gracefully.
	drainingFinishedWg *sync.WaitGroup
	// we want to make sure that we get 3 terminate requests in a short period ot time otherwise we assume
	// it is not a pod terminate
	events uint32
}

func NewDrainerService(logger log.FieldLogger, port uint) *DrainerService {
	triggeredWg := sync.WaitGroup{}
	triggeredWg.Add(1)
	schedulerWg := sync.WaitGroup{}
	schedulerWg.Add(1)
	return &DrainerService{
		port:               port,
		logger:             logger.WithField("source", "DrainerService"),
		serverReady:        false,
		triggered:          false,
		drainingFinishedWg: &schedulerWg,
		triggeredWg:        &triggeredWg,
		events:             0,
	}
}

func (drainer *DrainerService) SetState(state interface{}) {
}

func (drainer *DrainerService) Start() error {
	rtr := mux.NewRouter()
	rtr.HandleFunc(terminateEndpoint, drainer.terminate).Methods("GET")

	drainer.server = &http.Server{
		Addr: ":" + strconv.Itoa(int(drainer.port)), Handler: rtr,
	}
	drainer.logger.Infof("Starting drainer HTTP server on port %d", drainer.port)
	go func() {
		drainer.muServerReady.Lock()
		drainer.serverReady = true
		drainer.muServerReady.Unlock()
		err := drainer.server.ListenAndServe()
		drainer.logger.WithError(err).Info("HTTP drainer service stopped")
		drainer.muServerReady.Lock()
		drainer.serverReady = false
		drainer.muServerReady.Unlock()
	}()
	return nil
}

func (drainer *DrainerService) Ready() bool {
	drainer.muServerReady.RLock()
	defer drainer.muServerReady.RUnlock()
	return drainer.serverReady
}

func (drainer *DrainerService) Stop() error {
	drainer.logger.Info("Start graceful shutdown")
	// Shutdown is graceful
	drainer.muServerReady.Lock()
	defer drainer.muServerReady.Unlock()
	var err error
	if drainer.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), util.ServerControlPlaneTimeout)
		defer cancel()
		err = drainer.server.Shutdown(ctx)
	}
	drainer.serverReady = false
	drainer.logger.Info("Finished graceful shutdown")
	return err
}

func (drainer *DrainerService) Name() string {
	return "Agent drainer service"
}

func (drainer *DrainerService) WaitOnTrigger() {
	drainer.triggeredWg.Wait()
}

func (drainer *DrainerService) SetSchedulerDone() {
	drainer.drainingFinishedWg.Done()
}

func (drainer *DrainerService) terminate(w http.ResponseWriter, _ *http.Request) {
	// this is the crux of this service:
	// once someone (e.g. kubelet) calls `\terminate` we trigger downstream logic to drain this particular agent/server
	// the drain logic is defined in pkg/agent/server.go `drainServerReplicaImpl`
	// the flow is:
	// 0. wait for at least 3 events to arrive in a short-period of time (to signal pod restart)
	// 1. call \terminate (this is atomic)
	// 2. agent (drainOnRequest) is unblocked
	// 3. agent sends an AgentDrain grpc message to scheduler and waits for a reply
	// 4. scheduler reschedules models to a different server and waits for them to be Available
	// 5. grpc message returns to agent
	// 6. agent unblocks logic here
	// 7. \terminate returns

	drainer.muTriggered.Lock()
	drainer.events++
	drainer.muTriggered.Unlock()
	time.Sleep(eventWindowMs * time.Millisecond)
	drainer.muTriggered.Lock()
	if drainer.events >= eventsTarget {
		if !drainer.triggered {
			drainer.triggered = true
			drainer.triggeredWg.Done()
		}
		drainer.drainingFinishedWg.Wait()
	} else {
		drainer.events = 0
	}
	fmt.Fprintf(w, "ok\n")
	drainer.muTriggered.Unlock()
}
