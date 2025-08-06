/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package coordinator

import (
	"context"
	"reflect"

	busV3 "github.com/mustafaturan/bus/v3"
	log "github.com/sirupsen/logrus"
)

func (h *EventHub) RegisterPipelineStreamsEventHandler(
	name string,
	queueSize int,
	logger log.FieldLogger,
	handle func(event PipelineStreamsEventMsg),
) {
	events := make(chan PipelineStreamsEventMsg, queueSize)
	h.addPipelineStreamsEventHandlerChannel(events)

	go func() {
		for e := range events {
			handle(e)
		}
	}()

	handler := h.newPipelineStreamsEventHandler(logger, events)
	h.bus.RegisterHandler(name, handler)
}

func (h *EventHub) newPipelineStreamsEventHandler(
	logger log.FieldLogger,
	events chan PipelineStreamsEventMsg,
) busV3.Handler {
	handlePipelineStreamsEventMessage := func(_ context.Context, e busV3.Event) {
		l := logger.WithField("func", "handlePipelineStreamsEventMessage")
		l.Debugf("Received event on %s from %s (ID: %s, TxID: %s)", e.Topic, e.Source, e.ID, e.TxID)

		me, ok := e.Data.(PipelineStreamsEventMsg)
		if !ok {
			l.Warnf(
				"Event (ID %s, TxID %s) on topic %s from %s is not a PipelineStreamsEventMsg: %s",
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
		Matcher: topicPipelineStreamsEvents,
		Handle:  handlePipelineStreamsEventMessage,
	}
}

func (h *EventHub) addPipelineStreamsEventHandlerChannel(c chan PipelineStreamsEventMsg) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.pipelineStreamsEventHandlerChannels = append(h.pipelineStreamsEventHandlerChannels, c)
}

func (h *EventHub) PublishPipelineStreamsEvent(source string, event PipelineStreamsEventMsg) {
	err := h.bus.EmitWithOpts(
		context.Background(),
		topicPipelineStreamsEvents,
		event,
		busV3.WithSource(source),
	)
	if err != nil {
		h.logger.WithError(err).Errorf(
			"unable to publish pipeline streams event message from %s to %s",
			source,
			topicPipelineStreamsEvents,
		)
	}
}
