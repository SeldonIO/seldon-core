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

func (h *EventHub) RegisterModelEventHandler(
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

	handler := h.newModelEventHandler(logger, events)
	h.bus.RegisterHandler(name, handler)
}

func (h *EventHub) newModelEventHandler(
	logger log.FieldLogger,
	events chan ModelEventMsg,
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
