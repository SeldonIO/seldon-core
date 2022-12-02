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
	"context"
	"net/url"
	"time"

	"github.com/seldonio/seldon-core/hodometer/v2/pkg/hodometer"
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
		args.schedulerPlaintxtPort,
		args.schedulerTlsPort,
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
