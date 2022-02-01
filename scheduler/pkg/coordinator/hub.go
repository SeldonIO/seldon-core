package coordinator

import (
	"fmt"
	"sync"
)

type ModelEventMsg struct {
	ModelName    string
	ModelVersion uint32
}

func (m ModelEventMsg) String() string {
	return fmt.Sprintf("%s:%d", m.ModelName, m.ModelVersion)
}

type ModelEventHub struct {
	mu        sync.RWMutex
	closed    bool
	listeners []chan<- ModelEventMsg
}

func (h *ModelEventHub) AddListener(c chan<- ModelEventMsg) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners = append(h.listeners, c)
}

func (h *ModelEventHub) TriggerModelEvent(event ModelEventMsg) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	for _, listener := range h.listeners {
		listener <- event
	}
}

func (h *ModelEventHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, listener := range h.listeners {
		close(listener)
	}
	h.closed = true
}
