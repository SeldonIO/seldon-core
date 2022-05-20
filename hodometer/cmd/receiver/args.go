package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	envListenPort  = "LISTEN_PORT"
	envLogLevel    = "LOG_LEVEL"
	envRecordLevel = "RECORD_LEVEL"
)

const (
	defaultListenPort  = 8001
	defaultLogLevel    = logrus.InfoLevel
	defaultRecordLevel = recordLevelSummary
)

type RecordLevel string

const (
	recordLevelNone    RecordLevel = "NONE"
	recordLevelSummary RecordLevel = "SUMMARY"
	recordLevelAll     RecordLevel = "ALL"
)

type cliArgs struct {
	listenPort  uint
	logLevel    logrus.Level
	recordLevel RecordLevel
}

type parseFailure struct {
	arg     string
	failure error
}

func parseArgs(logger logrus.FieldLogger) (*cliArgs, error) {
	logger = logger.WithField("func", "parseArgs")
	failures := []parseFailure{}

	var err error

	listenPortFromEnv := os.Getenv(envListenPort)
	var listenPort uint
	if listenPortFromEnv == "" {
		listenPort = defaultListenPort
	} else {
		listenPort64, err := strconv.ParseUint(listenPortFromEnv, 10, 64)
		if err != nil {
			failures = append(
				failures,
				parseFailure{
					arg:     envListenPort,
					failure: err,
				},
			)
		}
		listenPort = uint(listenPort64)
	}

	logLevelFromEnv := os.Getenv(envLogLevel)
	var logLevel logrus.Level
	if logLevelFromEnv == "" {
		logLevel = defaultLogLevel
	} else {
		logLevel, err = logrus.ParseLevel(logLevelFromEnv)
		if err != nil {
			failures = append(
				failures,
				parseFailure{
					arg:     envLogLevel,
					failure: err,
				},
			)
		}
	}

	recordLevelFromEnv := os.Getenv(envRecordLevel)
	var recordLevel RecordLevel
	if recordLevelFromEnv == "" {
		recordLevel = defaultRecordLevel
	} else {
		normalised := strings.ToUpper(
			strings.TrimSpace(recordLevelFromEnv),
		)

		switch normalised {
		case "NONE":
			recordLevel = recordLevelNone
		case "SUMMARY":
			recordLevel = recordLevelSummary
		case "ALL":
			recordLevel = recordLevelAll
		default:
			failures = append(
				failures,
				parseFailure{
					arg: envRecordLevel,
					failure: fmt.Errorf(
						"unrecognised record level %s", recordLevelFromEnv,
					),
				},
			)
		}
	}

	if len(failures) > 0 {
		for _, f := range failures {
			logger.WithError(f.failure).Error(f.arg)
		}
		return nil, errors.New("failed to parse all required arguments")
	}

	return &cliArgs{
		listenPort:  uint(listenPort),
		logLevel:    logLevel,
		recordLevel: recordLevel,
	}, nil
}
