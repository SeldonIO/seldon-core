package main

import (
	"github.com/seldonio/seldon-core/hodometer/pkg/receiver"
	"github.com/sirupsen/logrus"
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

	receiver.Listen()
}
