/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// This file includes the ServerBasedSynchroniser struct and its methods.
// The ServerBasedSynchroniser struct is responsible for synchronising the starting up of the different components of the "scheduler".
// It ensures that the time between the scheduler starting and the different model servers connecting does not affect the data plane (inferences).
// In general terms, the synchroniser waits for all servers to connect before proceeding with processing events, especially those that are related to the servers connecting (i.e model scheduling).
// Otherwise if we dont wait for all servers to connect, we may get 404s when the scheduler tries to schedule models on servers that have not connected yet.
// The main trick is that the "controller" will send a number of expected servers to connect based on its etcd state, then the synchroniser will wait for all servers to connect before proceeding.
// The synchroniser will also wait for a timeout to be reached before proceeding if not all servers connect in time.
// the synchroniser subsribes to the server event handler and listens for SERVER_REPLICA_CONNECTED events, which are triggered when agents connect to the scheduler.
// The struct implements the Synchroniser interface, which includes the IsReady, WaitReady, and Signals methods.
// The struct also includes the handleServerEvents and doneFn methods.

// The ServerBasedSynchroniser struct is defined as follows:
// - It has the following fields:
//   - isReady: an atomic boolean value that indicates whether the synchroniser is ready.
//   - numEvents: an unsigned integer value that represents the number of events seen so far (connected servers).
//   - maxEvents: an unsigned integer value that represents the maximum number of events (expected servers).
//   - signalWg: a sync.WaitGroup value that is used to wait for the signal before processing events (controller to connect).
//   - eventHub: a pointer to the EventHub struct that is used to handle events.
//   - logger: a log.FieldLogger value that is used for logging.
//   - connectedServers: a map of strings to empty structs that stores the names of connected servers.
//   - connectedServersMu: a sync.Mutex value that is used to protect access to the connectedServers map.
//   - timeout: a time.Duration value that represents the timeout duration.
//   - doneWg: a sync.WaitGroup value that is used to wait for the timeout to be reached or all servers to connect.
//   - triggered: an atomic boolean value that indicates whether the synchroniser has been triggered.

package synchroniser

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

const (
	pendingSyncsQueueSize  int = 1000
	serverEventHandlerName     = "incremental.processor.servers"
)

type ServerBasedSynchroniser struct {
	isReady            atomic.Bool
	numEvents          uint64
	maxEvents          uint64
	signalWg           sync.WaitGroup
	eventHub           *coordinator.EventHub
	logger             log.FieldLogger
	connectedServers   map[string]struct{}
	connectedServersMu sync.Mutex
	timeout            time.Duration
	doneWg             sync.WaitGroup
	triggered          atomic.Bool
}

func NewServerBasedSynchroniser(eventHub *coordinator.EventHub, logger log.FieldLogger, timeout time.Duration) *ServerBasedSynchroniser {
	s := &ServerBasedSynchroniser{
		isReady:            atomic.Bool{},
		numEvents:          0,
		maxEvents:          0,
		eventHub:           eventHub,
		logger:             logger.WithField("source", "ServerBasedSynchroniser"),
		connectedServers:   make(map[string]struct{}),
		connectedServersMu: sync.Mutex{},
		timeout:            timeout,
		triggered:          atomic.Bool{},
	}
	s.isReady.Store(false)
	s.triggered.Store(false)
	s.signalWg.Add(1) // wait fist for signal before processing events
	s.doneWg.Add(1)   // we wait for the timeout to be reached or all servers to connect

	time.AfterFunc(s.timeout, s.doneFn)

	eventHub.RegisterServerEventHandler(
		serverEventHandlerName,
		pendingSyncsQueueSize,
		logger,
		s.handleServerEvents,
	)

	return s
}

func (s *ServerBasedSynchroniser) IsReady() bool {
	return s.isReady.Load()
}

func (s *ServerBasedSynchroniser) WaitReady() {
	if !s.isReady.Load() {
		s.doneWg.Wait()
	}
}

func (s *ServerBasedSynchroniser) Signals(numSignals uint) {
	if !s.isReady.Load() {
		swapped := s.triggered.CompareAndSwap(false, true) // make sure we run only once
		if swapped {
			atomic.AddUint64(&s.maxEvents, uint64(numSignals))
			s.signalWg.Done()
		}
	}
}

func (s *ServerBasedSynchroniser) doneFn() {
	if s.isReady.CompareAndSwap(false, true) {
		s.doneWg.Done()
		s.logger.Debugf("Synchroniser is ready")
	}
}

func (s *ServerBasedSynchroniser) handleServerEvents(event coordinator.ServerEventMsg) {
	logger := s.logger.WithField("func", "handleServerEvents")
	logger.Debugf("Received sync for server %s", event.String())

	// we do not want to block the event handler while waiting for the signal to be fired as it may cause a deadlock with
	// other events handlers.
	// we also do not care about order of events, so we can safely spawn a go routine to handle the signal
	go func() {
		s.connectedServersMu.Lock()
		defer s.connectedServersMu.Unlock()

		s.signalWg.Wait()

		if event.UpdateContext == coordinator.SERVER_REPLICA_CONNECTED && !s.IsReady() {
			serverNameWithIdx := fmt.Sprintf("%s-%d", event.ServerName, event.ServerIdx)

			if _, ok := s.connectedServers[serverNameWithIdx]; !ok {
				s.connectedServers[serverNameWithIdx] = struct{}{}
				if len(s.connectedServers) == int(s.maxEvents) {
					s.doneFn()
					s.logger.Infof("All (num: %d) servers connected", s.maxEvents)
				}
			}
		}
	}()
}
