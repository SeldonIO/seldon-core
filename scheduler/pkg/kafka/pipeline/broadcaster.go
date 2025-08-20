package pipeline

import "sync"

type Broadcaster struct {
	mu        sync.RWMutex
	listeners []chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		listeners: make([]chan struct{}, 0),
	}
}

func (b *Broadcaster) Subscribe() <-chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan struct{})
	b.listeners = append(b.listeners, ch)
	return ch
}

func (b *Broadcaster) HasListeners() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.listeners) > 0
}

func (b *Broadcaster) Broadcast() {
	b.mu.Lock()
	defer b.mu.Lock()

	for _, ch := range b.listeners {
		close(ch)
	}
	b.listeners = b.listeners[:0]
}
