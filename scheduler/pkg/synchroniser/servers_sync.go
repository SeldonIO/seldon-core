/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// TODO: explain synchroniser logic here
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
