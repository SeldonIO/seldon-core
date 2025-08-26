/*
Copyright (c) 2025 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package readyservice

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	readyEndpoint = "/ready"
)

type ReadyService struct {
	server      *http.Server
	port        uint
	logger      log.FieldLogger
	serverReady bool
	// mutex to guard changes to `serverReady`
	muServerReady   sync.RWMutex
	muTarget        sync.RWMutex
	readinessTarget interfaces.ServiceWithReadinessCheck
}

func NewReadyService(logger log.FieldLogger, port uint) *ReadyService {
	return &ReadyService{
		port:        port,
		logger:      logger.WithField("source", "ReadinessService"),
		serverReady: false,
	}
}

func (ready *ReadyService) SetState(state any) {
	ready.muTarget.Lock()
	defer ready.muTarget.Unlock()
	if state != nil {
		ready.readinessTarget = state.(interfaces.ServiceWithReadinessCheck)
	} else {
		ready.readinessTarget = nil
	}
}

func (ready *ReadyService) Start() error {
	rtr := mux.NewRouter()
	rtr.HandleFunc(readyEndpoint, ready.handleReadinessCheck).Methods("GET")

	ready.server = &http.Server{
		Addr: ":" + strconv.Itoa(int(ready.port)), Handler: rtr,
	}
	ready.logger.Infof("Starting HTTP server for readiness checks on port %d", ready.port)
	go func() {
		ready.muServerReady.Lock()
		ready.serverReady = true
		ready.muServerReady.Unlock()
		err := ready.server.ListenAndServe()
		ready.logger.WithError(err).Info("HTTP readiness service stopped")
		ready.muServerReady.Lock()
		ready.serverReady = false
		ready.muServerReady.Unlock()
	}()
	return nil
}

func (ready *ReadyService) Ready() bool {
	ready.muServerReady.RLock()
	defer ready.muServerReady.RUnlock()
	return ready.serverReady
}

func (ready *ReadyService) Stop() error {
	ready.logger.Info("Start graceful shutdown")
	// Shutdown is graceful
	ready.muServerReady.Lock()
	defer ready.muServerReady.Unlock()
	var err error
	if ready.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), util.ServerControlPlaneTimeout)
		defer cancel()
		err = ready.server.Shutdown(ctx)
	}
	ready.serverReady = false
	ready.logger.Info("Finished graceful shutdown")
	return err
}

func (ready *ReadyService) Name() string {
	return "Agent readiness service"
}

func (ready *ReadyService) GetType() interfaces.SubServiceType {
	return interfaces.CriticalControlPlaneService
}

func (ready *ReadyService) GetHTTPHandler() *http.Handler {
	if ready.server != nil {
		return &ready.server.Handler
	}
	return nil
}

func (ready *ReadyService) handleReadinessCheck(w http.ResponseWriter, _ *http.Request) {
	ready.muTarget.RLock()
	defer ready.muTarget.RUnlock()

	am := ready.readinessTarget

	if am == nil {
		ready.logger.Warn("Agent readiness HTTP handler failed, target not set")
		http.Error(w, "No service monitored for readiness by this endpoint", http.StatusNotFound)
		return
	}

	// Check if the agent and its sub-services are ready
	if am.Ready() {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("The agent and its dependent sub-services are ready"))
		if err != nil {
			ready.logger.WithError(err).Error("Failed to write response for readiness check")
		}
		return
	}

	ready.logger.Warn("Agent readiness failed, agent service manager is not ready")
	http.Error(w, "The agent is not ready", http.StatusServiceUnavailable)
}
