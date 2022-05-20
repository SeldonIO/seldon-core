package main

import (
	"context"
	"net/url"
	"time"

	"github.com/seldonio/seldon-core/hodometer/pkg/hodometer"
	"github.com/sirupsen/logrus"
)

var (
	// Multiple attempts per day in case of connectivity issues
	interval = 8 * time.Hour
)

func main() {
	logger := logrus.New()

	args, failures, err := parseArgs()
	if err != nil {
		for _, f := range failures {
			logger.WithError(f.failure).Error(f.arg)
		}
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

	urls := []*url.URL{args.publishUrl}
	urls = append(urls, args.extraPublishUrls...)
	jp, err := hodometer.NewJsonPublisher(logger, urls)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	punctuator.Run(
		"collect metrics and publish",
		func() {
			ctx := context.Background()
			metrics := scc.Collect(ctx, args.metricsLevel)
			_ = jp.Publish(ctx, metrics)
		},
	)
}
