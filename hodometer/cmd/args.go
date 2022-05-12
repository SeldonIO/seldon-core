package main

import (
	"errors"
	"os"
	"strconv"
	"strings"

	hodometer "github.com/seldonio/seldon-core/hodometer/pkg"
	"github.com/sirupsen/logrus"
)

const (
	envSchedulerHost = "SCHEDULER_HOST"
	envSchedulerPort = "SCHEDULER_PORT"
	envMetricsLevel  = "METRICS_LEVEL"
	envLogLevel      = "LOG_LEVEL"
	envClusterId     = "CLUSTER_ID"
)

type cliArgs struct {
	schedulerHost string
	schedulerPort uint
	metricsLevel  hodometer.MetricsLevel
	logLevel      logrus.Level
	clusterId     string
}

type parseFailure struct {
	arg     string
	failure error
}

func parseArgs(logger logrus.FieldLogger) (*cliArgs, error) {
	logger = logger.WithField("func", "parseArgs")
	failures := []parseFailure{}

	schedulerHost := os.Getenv(envSchedulerHost)
	if schedulerHost == "" {
		failures = append(
			failures,
			parseFailure{envSchedulerHost, errors.New("value not set")},
		)
	}

	schedulerPortFromEnv := os.Getenv(envSchedulerPort)
	if schedulerPortFromEnv == "" {
		failures = append(
			failures,
			parseFailure{envSchedulerPort, errors.New("value not set")},
		)
	}
	schedulerPort, err := strconv.ParseUint(schedulerPortFromEnv, 10, 64)
	if err != nil {
		failures = append(
			failures,
			parseFailure{envSchedulerPort, err},
		)
	}

	metricsLevelFromEnv := os.Getenv(envMetricsLevel)
	if metricsLevelFromEnv == "" {
		failures = append(
			failures,
			parseFailure{envMetricsLevel, errors.New("value not set")},
		)
	}
	metricsLevel, err := hodometer.MetricsLevelFrom(metricsLevelFromEnv)
	if err != nil {
		failures = append(
			failures,
			parseFailure{envMetricsLevel, err},
		)
	}

	logLevelFromEnv := os.Getenv(envLogLevel)
	var logLevel logrus.Level
	if logLevelFromEnv == "" {
		logLevel = logrus.InfoLevel
	} else {
		logLevel, err = logrus.ParseLevel(logLevelFromEnv)
		if err != nil {
			failures = append(
				failures,
				parseFailure{envLogLevel, err},
			)
		}
	}

	clusterIdFromEnv := os.Getenv(envClusterId)
	clusterId := strings.TrimSpace(clusterIdFromEnv)

	if len(failures) > 0 {
		for _, f := range failures {
			logger.WithError(f.failure).Error(f.arg)
		}
		return nil, errors.New("failed to parse all required arguments")
	}

	return &cliArgs{
		schedulerHost: schedulerHost,
		schedulerPort: uint(schedulerPort),
		metricsLevel:  metricsLevel,
		logLevel:      logLevel,
		clusterId:     clusterId,
	}, nil
}
