package coordinator

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	busV3 "github.com/mustafaturan/bus/v3"
	log "github.com/sirupsen/logrus"
)

const (
	topicModelEvents = "model.event"
)

type SequenceGenerator struct {
	counter int64
}

func (g *SequenceGenerator) Generate() string {
	next := atomic.AddInt64(&g.counter, 1)
	return fmt.Sprintf("%d", next)
}

var _ busV3.IDGenerator = (*SequenceGenerator)(nil)

type EventHub struct {
	bus                       *busV3.Bus
	logger                    log.FieldLogger
	modelEventHandlerChannels []chan ModelEventMsg
	lock                      sync.RWMutex
	closed                    bool
}

// NewEventHub creates a new EventHub with topics pre-registered.
// The logger l does not need fields preset.
func NewEventHub(l log.FieldLogger) (*EventHub, error) {
	generator := &SequenceGenerator{}
	bus, err := busV3.NewBus(generator)
	if err != nil {
		return nil, err
	}

	hub := EventHub{
		logger: l.WithField("source", "EventHub"),
		bus:    bus,
	}

	hub.bus.RegisterTopics(topicModelEvents)

	return &hub, nil
}

func (h *EventHub) Close() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.closed = true

	for _, c := range h.modelEventHandlerChannels {
		close(c)
	}
}

func (h *EventHub) RegisterHandler(
	name string,
	queueSize int,
	logger log.FieldLogger,
	handle func(event ModelEventMsg),
) {
	events := make(chan ModelEventMsg, queueSize)
	h.addModelEventHandlerChannel(events)

	go func() {
		for e := range events {
			handle(e)
		}
	}()

	handler := h.newModelEventHandler(logger, events, handle)
	h.bus.RegisterHandler(name, handler)
}

func (h *EventHub) newModelEventHandler(
	logger log.FieldLogger,
	events chan ModelEventMsg,
	handle func(event ModelEventMsg),
) busV3.Handler {
	handleModelEventMessage := func(_ context.Context, e busV3.Event) {
		l := logger.WithField("func", "handleModelEventMessage")
		l.Debugf("Received event on %s from %s (ID: %s, TxID: %s)", e.Topic, e.Source, e.ID, e.TxID)

		me, ok := e.Data.(ModelEventMsg)
		if !ok {
			l.Warnf(
				"Event (ID %s, TxID %s) on topic %s from %s is not a ModelEventMsg: %s",
				e.ID,
				e.TxID,
				e.Topic,
				e.Source,
				reflect.TypeOf(e.Data).String(),
			)
			return
		}

		h.lock.RLock()
		if h.closed {
			return
		}
		events <- me
		h.lock.RUnlock()
	}

	return busV3.Handler{
		Matcher: topicModelEvents,
		Handle:  handleModelEventMessage,
	}
}

func (h *EventHub) addModelEventHandlerChannel(c chan ModelEventMsg) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.modelEventHandlerChannels = append(h.modelEventHandlerChannels, c)
}

func (h *EventHub) PublishModelEvent(source string, event ModelEventMsg) {
	err := h.bus.EmitWithOpts(
		context.Background(),
		topicModelEvents,
		event,
		busV3.WithSource(source),
	)
	if err != nil {
		h.logger.WithError(err).Errorf(
			"unable to publish model event message from %s to %s",
			source,
			topicModelEvents,
		)
	}
}
