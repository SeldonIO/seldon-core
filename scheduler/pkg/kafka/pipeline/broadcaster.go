/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import "sync"

type Broadcaster struct {
	mu    sync.RWMutex
	ready chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{}
}

func (b *Broadcaster) Subscribe() <-chan struct{} {
	b.mu.RLock()
	if b.ready != nil {
		defer b.mu.RUnlock()
		return b.ready
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()
	// we must check again if ready has been set as multiple goroutines could have been waiting to acquire lock
	if b.ready != nil {
		return b.ready
	}
	b.ready = make(chan struct{})
	return b.ready
}

func (b *Broadcaster) Broadcast() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ready != nil {
		close(b.ready)
		b.ready = nil
	}
}
