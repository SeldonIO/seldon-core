/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/hodometer/v2/pkg/receiver"
)

func main() {
	baseLogger := logrus.New()
	logger := baseLogger.
		WithField("source", "main").
		WithField("func", "main")

	args, err := parseArgs(logger)
	if err != nil {
		baseLogger.WithError(err).Fatal()
	}

	baseLogger.SetLevel(args.logLevel)

	var recorder receiver.Recorder
	switch args.recordLevel {
	case recordLevelNone:
		logger.Info("not recording events")
		recorder = receiver.NewNoopRecorder()
	case recordLevelSummary:
		logger.Info("recording event summaries")
		recorder = receiver.NewCountingRecorder(receiver.NewNoopRecorder())
	case recordLevelAll:
		logger.Info("recording full events")
		recorder = receiver.NewOrderedRecorder(logger)
	}

	receiver := receiver.NewReceiver(
		logger,
		args.listenPort,
		recorder,
	)

	if err := receiver.Listen(); err != nil {
		baseLogger.WithError(err).Error("Listener stopped")
	}
}
