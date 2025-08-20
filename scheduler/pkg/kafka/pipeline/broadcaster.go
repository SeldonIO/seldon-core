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
	mu    sync.Mutex
	ready chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{}
}

func (b *Broadcaster) Subscribe() <-chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ready == nil {
		b.ready = make(chan struct{})
	}
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
