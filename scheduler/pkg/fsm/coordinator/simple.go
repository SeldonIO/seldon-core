/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package coordinator

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/events"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/orchestrator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/orchestrator/handlers"
)

// EventPriority defines the priority of events (lower number = higher priority)
type EventPriority int

const (
	PriorityHigh   EventPriority = 0 // Server updates, critical failures
	PriorityNormal EventPriority = 1 // Model loads, pipeline updates
	PriorityLow    EventPriority = 2 // Metrics, status updates
)

// PrioritizedEvent wraps an event with priority and timestamp
type PrioritizedEvent struct {
	Event     events.Event
	Priority  EventPriority
	Timestamp time.Time
	index     int // Used by heap
}

// EventPriorityQueue implements heap.Interface for priority-based event processing
type EventPriorityQueue []*PrioritizedEvent

func (pq EventPriorityQueue) Len() int { return len(pq) }

func (pq EventPriorityQueue) Less(i, j int) bool {
	// Lower priority number = higher priority
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority < pq[j].Priority
	}
	// Same priority: FIFO (earlier timestamp first)
	return pq[i].Timestamp.Before(pq[j].Timestamp)
}

func (pq EventPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *EventPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PrioritizedEvent)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *EventPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// EventSubscriber receives output events from the coordinator
type EventSubscriber struct {
	ID      string
	Channel chan events.Output
	Filter  func(events.Output) bool // Optional filter to only receive certain events
}

// SimpleCoordinator is a single-threaded event processor with priority queue
type SimpleCoordinator struct {
	fsm *orchestrator.Fsm

	// Event queue
	eventQueue EventPriorityQueue
	queueMu    sync.Mutex
	queueCond  *sync.Cond

	// Event fan-out
	subscribers map[string]*EventSubscriber
	subMu       sync.RWMutex

	// Control
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
	runningMu sync.RWMutex
	wg        sync.WaitGroup

	// Stats
	stats CoordinatorStatistics
}

// CoordinatorStatistics tracks coordinator metrics
type CoordinatorStatistics struct {
	mu                 sync.RWMutex
	EventsProcessed    int64
	EventsQueued       int64
	CurrentQueueDepth  int
	Subscribers        int
	LastProcessedEvent time.Time
}

// NewSimpleCoordinator creates a single-threaded coordinator
func NewSimpleCoordinator(fsm *orchestrator.Fsm) *SimpleCoordinator {
	c := &SimpleCoordinator{
		fsm:         fsm,
		eventQueue:  make(EventPriorityQueue, 0),
		subscribers: make(map[string]*EventSubscriber),
	}
	c.queueCond = sync.NewCond(&c.queueMu)

	heap.Init(&c.eventQueue)

	return c
}

// Start begins processing events in a single background goroutine
func (c *SimpleCoordinator) Start(ctx context.Context) error {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.running {
		return fmt.Errorf("coordinator already running")
	}

	// Initialize FSM (replay uncommitted events)
	if err := c.fsm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start FSM: %w", err)
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.running = true

	// Start single event processing loop
	c.wg.Add(1)
	go c.processLoop()

	return nil
}

// Stop gracefully shuts down the coordinator
func (c *SimpleCoordinator) Stop() error {
	c.runningMu.Lock()
	if !c.running {
		c.runningMu.Unlock()
		return fmt.Errorf("coordinator not running")
	}
	c.runningMu.Unlock()

	// Signal shutdown
	c.cancel()

	// Wake up processing loop if it's waiting
	c.queueCond.Signal()

	// Wait for processing to finish
	c.wg.Wait()

	// Close all subscriber channels
	c.subMu.Lock()
	for _, sub := range c.subscribers {
		close(sub.Channel)
	}
	c.subscribers = make(map[string]*EventSubscriber)
	c.subMu.Unlock()

	c.runningMu.Lock()
	c.running = false
	c.runningMu.Unlock()

	return nil
}

// SubmitEvent adds an event to the priority queue (non-blocking)
func (c *SimpleCoordinator) SubmitEvent(event events.Event, priority EventPriority) error {
	c.runningMu.RLock()
	if !c.running {
		c.runningMu.RUnlock()
		return fmt.Errorf("coordinator not running")
	}
	c.runningMu.RUnlock()

	prioritizedEvent := &PrioritizedEvent{
		Event:     event,
		Priority:  priority,
		Timestamp: time.Now(),
	}

	c.queueMu.Lock()
	heap.Push(&c.eventQueue, prioritizedEvent)
	c.updateStats()
	c.queueMu.Unlock()

	// Wake up processing loop
	c.queueCond.Signal()

	return nil
}

// SubmitEventSync adds an event and waits for it to be processed (blocking)
// Returns the output events generated
func (c *SimpleCoordinator) SubmitEventSync(ctx context.Context, event events.Event, priority EventPriority) ([]events.Output, error) {
	// Create a temporary subscriber to capture output
	resultChan := make(chan events.Output, 10)
	subID := fmt.Sprintf("sync-%d", time.Now().UnixNano())

	// Subscribe before submitting
	sub := &EventSubscriber{
		ID:      subID,
		Channel: resultChan,
	}

	c.subMu.Lock()
	c.subscribers[subID] = sub
	c.subMu.Unlock()

	// Submit event
	if err := c.SubmitEvent(event, priority); err != nil {
		c.Unsubscribe(subID)
		return nil, err
	}

	// Wait for results (with timeout)
	var outputEvents []events.Output
	timeout := time.After(5 * time.Second)

collectLoop:
	for {
		select {
		case outEvent, ok := <-resultChan:
			if !ok {
				break collectLoop
			}
			outputEvents = append(outputEvents, outEvent)

			// Check if this is the last event for our input event
			// (This is a simplification - in production you'd correlate events)
			// For now, we'll just collect for a short window
			if len(outputEvents) > 0 {
				time.Sleep(10 * time.Millisecond) // Small window to collect related events
				break collectLoop
			}

		case <-timeout:
			c.Unsubscribe(subID)
			return nil, fmt.Errorf("timeout waiting for event processing")

		case <-ctx.Done():
			c.Unsubscribe(subID)
			return nil, ctx.Err()
		}
	}

	c.Unsubscribe(subID)
	return outputEvents, nil
}

// Subscribe registers a channel to receive output events
func (c *SimpleCoordinator) Subscribe(id string, bufferSize int, filter func(events.Output) bool) <-chan events.Output {
	ch := make(chan events.Output, bufferSize)

	sub := &EventSubscriber{
		ID:      id,
		Channel: ch,
		Filter:  filter,
	}

	c.subMu.Lock()
	c.subscribers[id] = sub
	c.subMu.Unlock()

	return ch
}

// Unsubscribe removes a subscriber
func (c *SimpleCoordinator) Unsubscribe(id string) {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	if sub, exists := c.subscribers[id]; exists {
		close(sub.Channel)
		delete(c.subscribers, id)
	}
}

// processLoop is the main single-threaded event processing loop
func (c *SimpleCoordinator) processLoop() {
	defer c.wg.Done()

	for {
		// Wait for events or shutdown
		c.queueMu.Lock()
		for c.eventQueue.Len() == 0 {
			// Check if we should shutdown
			select {
			case <-c.ctx.Done():
				c.queueMu.Unlock()
				return
			default:
			}

			// Wait for signal
			c.queueCond.Wait()

			// Check shutdown again after waking up
			select {
			case <-c.ctx.Done():
				c.queueMu.Unlock()
				return
			default:
			}
		}

		// Pop highest priority event
		prioritizedEvent := heap.Pop(&c.eventQueue).(*PrioritizedEvent)
		c.updateStats()
		c.queueMu.Unlock()

		// Process the event
		c.processEvent(prioritizedEvent.Event)
	}
}

// processEvent handles a single event through the FSM
func (c *SimpleCoordinator) processEvent(event events.Event) {
	// Apply event to FSM
	outputEvents, err := c.fsm.Apply(c.ctx, event)
	if err != nil {
		// TODO: Handle errors (dead letter queue, retry, etc.)
		// For now, just log or emit error event
		fmt.Printf("Error processing event %s: %v\n", event.Type(), err)
		return
	}

	// Update stats
	c.stats.mu.Lock()
	c.stats.EventsProcessed++
	c.stats.LastProcessedEvent = time.Now()
	c.stats.mu.Unlock()

	// Fan out output events to subscribers
	c.fanOutEvents(outputEvents)
}

// fanOutEvents sends output events to all subscribers
func (c *SimpleCoordinator) fanOutEvents(events []events.Output) {
	if len(events) == 0 {
		return
	}

	c.subMu.RLock()
	defer c.subMu.RUnlock()

	for _, event := range events {
		for _, sub := range c.subscribers {
			// Apply filter if present
			if sub.Filter != nil && !sub.Filter(event) {
				continue
			}

			// Non-blocking send
			select {
			case sub.Channel <- event:
			default:
				// Channel full - subscriber is slow
				// TODO: Handle slow subscribers (drop, buffer, disconnect, etc.)
				fmt.Printf("Warning: subscriber %s channel full, dropping event\n", sub.ID)
			}
		}
	}
}

// GetStatistics returns current coordinator statistics
func (c *SimpleCoordinator) GetStatistics() CoordinatorStatistics {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	c.queueMu.Lock()
	queueDepth := c.eventQueue.Len()
	c.queueMu.Unlock()

	c.subMu.RLock()
	subCount := len(c.subscribers)
	c.subMu.RUnlock()

	return CoordinatorStatistics{
		EventsProcessed:    c.stats.EventsProcessed,
		EventsQueued:       c.stats.EventsQueued,
		CurrentQueueDepth:  queueDepth,
		Subscribers:        subCount,
		LastProcessedEvent: c.stats.LastProcessedEvent,
	}
}

// updateStats updates queue statistics (must be called with queueMu held)
func (c *SimpleCoordinator) updateStats() {
	c.stats.mu.Lock()
	c.stats.CurrentQueueDepth = c.eventQueue.Len()
	c.stats.mu.Unlock()
}

// ========================================
// Helper Methods for Common Event Types
// ========================================

// SubmitLoadModelEvent is a convenience method for submitting load model events
func (c *SimpleCoordinator) SubmitLoadModelEvent(event *handlers.LoadModelEvent) error {
	return c.SubmitEvent(event, PriorityNormal)
}

// SubmitServerUpdateEvent is a convenience method for high-priority server updates
func (c *SimpleCoordinator) SubmitServerUpdateEvent(event events.Event) error {
	return c.SubmitEvent(event, PriorityHigh)
}

// SubmitStatusUpdateEvent is a convenience method for low-priority status updates
func (c *SimpleCoordinator) SubmitStatusUpdateEvent(event events.Event) error {
	return c.SubmitEvent(event, PriorityLow)
}

// what would be the highes priority?

// server
// models
// experiments and pipelines in the same
