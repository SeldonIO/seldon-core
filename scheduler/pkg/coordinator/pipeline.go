package coordinator

import (
	"context"
	"reflect"

	busV3 "github.com/mustafaturan/bus/v3"
	log "github.com/sirupsen/logrus"
)

func (h *EventHub) RegisterPipelineEventHandler(
	name string,
	queueSize int,
	logger log.FieldLogger,
	handle func(event PipelineEventMsg),
) {
	events := make(chan PipelineEventMsg, queueSize)
	h.addPipelineEventHandlerChannel(events)

	go func() {
		for e := range events {
			handle(e)
		}
	}()

	handler := h.newPipelineEventHandler(logger, events)
	h.bus.RegisterHandler(name, handler)
}

func (h *EventHub) newPipelineEventHandler(
	logger log.FieldLogger,
	events chan PipelineEventMsg,
) busV3.Handler {
	handlePipelineEventMessage := func(_ context.Context, e busV3.Event) {
		l := logger.WithField("func", "handlePipelineEventMessage")
		l.Debugf("Received event on %s from %s (ID: %s, TxID: %s)", e.Topic, e.Source, e.ID, e.TxID)

		me, ok := e.Data.(PipelineEventMsg)
		if !ok {
			l.Warnf(
				"Event (ID %s, TxID %s) on topic %s from %s is not a PipelineEventMsg: %s",
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
		Matcher: topicPipelineEvents,
		Handle:  handlePipelineEventMessage,
	}
}

func (h *EventHub) addPipelineEventHandlerChannel(c chan PipelineEventMsg) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.pipelineEventHandlerChannels = append(h.pipelineEventHandlerChannels, c)
}

func (h *EventHub) PublishPipelineEvent(source string, event PipelineEventMsg) {
	err := h.bus.EmitWithOpts(
		context.Background(),
		topicPipelineEvents,
		event,
		busV3.WithSource(source),
	)
	if err != nil {
		h.logger.WithError(err).Errorf(
			"unable to publish pipeline event message from %s to %s",
			source,
			topicPipelineEvents,
		)
	}
}
