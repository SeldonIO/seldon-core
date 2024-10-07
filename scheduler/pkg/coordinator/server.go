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

func (h *EventHub) RegisterServerEventHandler(
	name string,
	queueSize int,
	logger log.FieldLogger,
	handle func(event ServerEventMsg),
) {
	events := make(chan ServerEventMsg, queueSize)
	h.addServerEventHandlerChannel(events)

	go func() {
		for e := range events {
			handle(e)
		}
	}()

	handler := h.newServerEventHandler(logger, events, handle)
	h.bus.RegisterHandler(name, handler)
}

func (h *EventHub) newServerEventHandler(
	logger log.FieldLogger,
	events chan ServerEventMsg,
	_ func(event ServerEventMsg),
) busV3.Handler {
	handleServerEventMessage := func(_ context.Context, e busV3.Event) {
		l := logger.WithField("func", "handleServerEventMessage")
		l.Debugf("Received event on %s from %s (ID: %s, TxID: %s)", e.Topic, e.Source, e.ID, e.TxID)

		me, ok := e.Data.(ServerEventMsg)
		if !ok {
			l.Warnf(
				"Event (ID %s, TxID %s) on topic %s from %s is not a ServerEventMsg: %s",
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
		// Propagate the busV3.Event source to the ServerEventMsg
		// This is useful for logging, but also in case we want to distinguish
		// the action to take based on where the event came from.
		me.Source = e.Source
		events <- me
		h.lock.RUnlock()
	}

	return busV3.Handler{
		Matcher: topicServerEvents,
		Handle:  handleServerEventMessage,
	}
}

func (h *EventHub) addServerEventHandlerChannel(c chan ServerEventMsg) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.serverEventHandlerChannels = append(h.serverEventHandlerChannels, c)
}

func (h *EventHub) PublishServerEvent(source string, event ServerEventMsg) {
	err := h.bus.EmitWithOpts(
		context.Background(),
		topicServerEvents,
		event,
		busV3.WithSource(source),
	)
	if err != nil {
		h.logger.WithError(err).Errorf(
			"unable to publish server event message from %s to %s",
			source,
			topicServerEvents,
		)
	}
}
