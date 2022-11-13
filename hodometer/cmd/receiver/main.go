/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	if err := receiver.Listen(); err != nil {
		baseLogger.WithError(err).Error("Listener stopped")
	}
}
