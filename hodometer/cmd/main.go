package main

import (
	"context"
	"time"

	hodometer "github.com/seldonio/seldon-core/hodometer/pkg"
	"github.com/sirupsen/logrus"
)

var (
	// Multiple attempts per day in case of connectivity issues
	interval = 8 * time.Hour
)

func main() {
	logger := logrus.New()

	args, err := parseArgs(logger)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	logger.SetLevel(args.logLevel)

	build := hodometer.GetBuildDetails()
	logger.WithFields(build).Info("build information")

	punctuator := hodometer.NewPunctuator(logger, interval)

	scc, err := hodometer.NewSeldonCoreCollector(
		logger,
		args.schedulerHost,
		args.schedulerPort,
		args.clusterId,
	)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	jp := hodometer.NewJsonPublisher(logger)

	punctuator.Run(
		"collect metrics and publish",
		func() {
			ctx := context.Background()
			metrics := scc.Collect(ctx, args.metricsLevel)
			_ = jp.Publish(ctx, metrics)
		},
	)
}
