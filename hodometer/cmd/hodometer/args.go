package main

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/seldonio/seldon-core/hodometer/pkg/hodometer"
	"github.com/sirupsen/logrus"
)

const (
	envSchedulerHost         = "SCHEDULER_HOST"
	envSchedulerPlaintxtPort = "SCHEDULER_PLAINTXT_PORT"
	envSchedulerTlsPort      = "SCHEDULER_TLS_PORT"
	envMetricsLevel          = "METRICS_LEVEL"
	envLogLevel              = "LOG_LEVEL"
	envClusterId             = "CLUSTER_ID"
	envPublishUrl            = "PUBLISH_URL"
	envExtraPublishUrls      = "EXTRA_PUBLISH_URLS"
)

const (
	defaultLogLevel   = logrus.InfoLevel
	defaultPublishUrl = "http://hodometer-test.seldon.io"
)

type cliArgs struct {
	schedulerHost         string
	schedulerPlaintxtPort uint
	schedulerTlsPort      uint
	metricsLevel          hodometer.MetricsLevel
	logLevel              logrus.Level
	clusterId             string
	publishUrl            *url.URL
	extraPublishUrls      []*url.URL
}

type parseFailure struct {
	arg     string
	failure error
}

func parseArgs() (*cliArgs, []parseFailure, error) {
	failures := []parseFailure{}
	args := &cliArgs{}

	failures = setSchedulerHost(args, failures)
	failures = setSchedulerPlaintxtPort(args, failures)
	failures = setSchedulerTlsPort(args, failures)
	failures = setMetricsLevel(args, failures)
	failures = setLogLevel(args, failures)
	failures = setClusterId(args, failures)
	failures = setPublishUrl(args, failures)
	failures = setExtraPublishUrls(args, failures)

	if len(failures) > 0 {
		return nil, failures, errors.New("failed to parse all required arguments")
	}

	return args, nil, nil
}

func setSchedulerHost(args *cliArgs, failures []parseFailure) []parseFailure {
	schedulerHost := os.Getenv(envSchedulerHost)
	if schedulerHost == "" {
		failures = append(
			failures,
			parseFailure{envSchedulerHost, errors.New("value not set")},
		)
		return failures
	}

	args.schedulerHost = schedulerHost
	return failures
}

func setSchedulerPlaintxtPort(args *cliArgs, failures []parseFailure) []parseFailure {
	schedulerPortFromEnv := os.Getenv(envSchedulerPlaintxtPort)
	if schedulerPortFromEnv == "" {
		failures = append(
			failures,
			parseFailure{envSchedulerPlaintxtPort, errors.New("value not set")},
		)
		return failures
	}
	schedulerPort, err := strconv.ParseUint(schedulerPortFromEnv, 10, 64)
	if err != nil {
		failures = append(
			failures,
			parseFailure{envSchedulerPlaintxtPort, err},
		)
		return failures
	}
	args.schedulerPlaintxtPort = uint(schedulerPort)
	return failures
}

func setSchedulerTlsPort(args *cliArgs, failures []parseFailure) []parseFailure {
	schedulerPortFromEnv := os.Getenv(envSchedulerTlsPort)
	if schedulerPortFromEnv == "" {
		failures = append(
			failures,
			parseFailure{envSchedulerTlsPort, errors.New("value not set")},
		)
		return failures
	}
	schedulerPort, err := strconv.ParseUint(schedulerPortFromEnv, 10, 64)
	if err != nil {
		failures = append(
			failures,
			parseFailure{envSchedulerTlsPort, err},
		)
		return failures
	}
	args.schedulerTlsPort = uint(schedulerPort)
	return failures
}

func setMetricsLevel(args *cliArgs, failures []parseFailure) []parseFailure {
	metricsLevelFromEnv := os.Getenv(envMetricsLevel)
	if metricsLevelFromEnv == "" {
		failures = append(
			failures,
			parseFailure{envMetricsLevel, errors.New("value not set")},
		)
		return failures
	}
	metricsLevel, err := hodometer.MetricsLevelFrom(metricsLevelFromEnv)
	if err != nil {
		failures = append(
			failures,
			parseFailure{envMetricsLevel, err},
		)
		return failures
	}

	args.metricsLevel = metricsLevel
	return failures
}

func setLogLevel(args *cliArgs, failures []parseFailure) []parseFailure {
	var logLevel logrus.Level
	var err error

	logLevelFromEnv := os.Getenv(envLogLevel)
	if logLevelFromEnv == "" {
		logLevel = defaultLogLevel
	} else {
		logLevel, err = logrus.ParseLevel(logLevelFromEnv)
		if err != nil {
			failures = append(
				failures,
				parseFailure{envLogLevel, err},
			)
			return failures
		}
	}

	args.logLevel = logLevel
	return failures
}

func setClusterId(args *cliArgs, failures []parseFailure) []parseFailure {
	clusterIdFromEnv := os.Getenv(envClusterId)
	clusterId := strings.TrimSpace(clusterIdFromEnv)

	args.clusterId = clusterId
	return failures
}

func setPublishUrl(args *cliArgs, failures []parseFailure) []parseFailure {
	publishUrlFromEnv := os.Getenv(envPublishUrl)
	if publishUrlFromEnv == "" {
		publishUrlFromEnv = defaultPublishUrl
	}
	publishUrl, err := url.Parse(publishUrlFromEnv)
	if err != nil {
		failures = append(
			failures,
			parseFailure{envPublishUrl, err},
		)
		return failures
	}

	args.publishUrl = publishUrl
	return failures
}

func setExtraPublishUrls(args *cliArgs, failures []parseFailure) []parseFailure {
	extraUrlsFromEnv := os.Getenv(envExtraPublishUrls)
	if extraUrlsFromEnv == "" {
		return failures
	}

	rawExtraUrls := strings.Split(extraUrlsFromEnv, ",")
	extraUrls := make([]*url.URL, len(rawExtraUrls))
	for i, raw := range rawExtraUrls {
		normalised := strings.TrimSpace(raw)
		u, err := url.Parse(normalised)
		if err != nil {
			failures = append(
				failures,
				parseFailure{
					arg:     envExtraPublishUrls,
					failure: err,
				},
			)
			return failures
		}
		extraUrls[i] = u
	}

	args.extraPublishUrls = extraUrls
	return failures
}
