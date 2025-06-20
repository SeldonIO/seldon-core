/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// This package is responsible to synchronise starting up the different components of the "scheduler".
// In particular, it is responsible for making sure that the time between the scheduler starts and while
// the different model servers connect that the data plane (inferences) is not affected.
package synchroniser

import (
	"sync"
	"sync/atomic"
	"time"
)

type Synchroniser interface {
	// mainly for testing, this api should mean little in production as the synchroniser should be
	// rely on the other methods to determine if it is ready.
	IsTriggered() bool
	IsReady() bool
	WaitReady()
	Signals(uint)
}

type SimpleSynchroniser struct {
	isReady   atomic.Bool
	wg        sync.WaitGroup
	timeout   time.Duration
	triggered atomic.Bool
}

func NewSimpleSynchroniser(timeout time.Duration) *SimpleSynchroniser {
	s := &SimpleSynchroniser{
		isReady: atomic.Bool{},
		wg:      sync.WaitGroup{},
		timeout: timeout,
	}
	s.isReady.Store(false)
	s.triggered.Store(false)
	s.wg.Add(1)
	time.AfterFunc(s.timeout, s.done)
	return s
}

func (s *SimpleSynchroniser) IsTriggered() bool {
	return s.triggered.Load()
}

func (s *SimpleSynchroniser) IsReady() bool {
	return s.isReady.Load()
}

func (s *SimpleSynchroniser) Signals(_ uint) {
	s.triggered.Store(true)
}

func (s *SimpleSynchroniser) WaitReady() {
	if !s.IsReady() {
		s.wg.Wait()
	}
}

func (s *SimpleSynchroniser) done() {
	s.isReady.Store(true)
	s.wg.Done()
}
