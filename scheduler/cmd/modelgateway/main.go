/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/gateway"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
)

const (
	flagSchedulerHost            = "scheduler-host"
	flagSchedulerPlaintxtPort    = "scheduler-plaintxt-port"
	flagSchedulerTlsPort         = "scheduler-tls-port"
	flagEnvoyHost                = "envoy-host"
	flagEnvoyPort                = "envoy-port"
	flagLogLevel                 = "log-level"
	defaultSchedulerPlaintxtPort = 9004
	defaultSchedulerTLSPort      = 9044
	defaultEnvoyPort             = 9000
	kubernetesNamespacePath      = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	schedulerHost         string
	schedulerPlaintxtPort int
	schedulerTlsPort      int
	envoyHost             string
	envoyPort             int
	kafkaConfigPath       string
	namespace             string
	logLevel              string
	tracingConfigPath     string
)

func init() {

	flag.StringVar(&schedulerHost, flagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPlaintxtPort, flagSchedulerPlaintxtPort, defaultSchedulerPlaintxtPort, "Scheduler plaintxt port")
	flag.IntVar(&schedulerTlsPort, flagSchedulerTlsPort, defaultSchedulerTLSPort, "Scheduler TLS port")
	flag.StringVar(&envoyHost, flagEnvoyHost, "0.0.0.0", "Envoy host")
	flag.IntVar(&envoyPort, flagEnvoyPort, defaultEnvoyPort, "Envoy port")
	flag.StringVar(&namespace, "namespace", "", "Namespace")
	flag.StringVar(
		&kafkaConfigPath,
		"kafka-config-path",
		"/mnt/config/kafka.json",
		"Path to kafka configuration file",
	)
	flag.StringVar(&logLevel, flagLogLevel, "debug", "Log level - examples: debug, info, error")
	flag.StringVar(&tracingConfigPath, "tracing-config-path", "", "Tracing config path")
}

func updateNamespace() {
	nsBytes, err := os.ReadFile(kubernetesNamespacePath)
	if err != nil {
		log.Warn("Using namespace from command line argument")
	} else {
		ns := string(nsBytes)
		log.Infof("Setting namespace from k8s file to %s", ns)
		namespace = ns
	}
}

func makeSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
}

func getEnVar(logger *log.Logger, key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to parse %s", key)
		}
		logger.Infof("Got %s = %d", key, val)
		return int(val)
	}
	logger.Infof("Returning default %s = %d", key, defaultValue)
	return defaultValue
}

func main() {
	logger := log.New()
	flag.Parse()

	logIntLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", logLevel)
	}
	logger.Infof("Setting log level to %s", logLevel)
	logger.SetLevel(logIntLevel)

	updateNamespace()

	done := make(chan bool, 1)

	go makeSignalHandler(logger, done)

	tracer, err := tracing.NewTraceProvider("seldon-modelgateway", &tracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	kafkaConfigMap, err := kafka_config.NewKafkaConfig(kafkaConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load Kafka config")
	}

	inferServerConfig := &gateway.InferenceServerConfig{
		Host:     envoyHost,
		HttpPort: envoyPort,
		GrpcPort: envoyPort,
	}
	consumerConfig := gateway.ManagerConfig{
		SeldonKafkaConfig:     kafkaConfigMap,
		Namespace:             namespace,
		InferenceServerConfig: inferServerConfig,
		TraceProvider:         tracer,
		NumWorkers:            getEnVar(logger, gateway.EnvVarNumWorkers, gateway.DefaultNumWorkers),
	}
	kafkaConsumer, err := gateway.NewConsumerManager(logger, &consumerConfig,
		getEnVar(logger, gateway.EnvMaxNumConsumers, gateway.DefaultMaxNumConsumers))
	if err != nil {
		logger.WithError(err).Fatal("Failed to create consumer manager")
	}
	defer kafkaConsumer.Stop()

	kafkaSchedulerClient := gateway.NewKafkaSchedulerClient(logger, kafkaConsumer)
	err = kafkaSchedulerClient.ConnectToScheduler(schedulerHost, schedulerPlaintxtPort, schedulerTlsPort)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to connect to scheduler")
	}
	defer kafkaSchedulerClient.Stop()

	go func() {
		err := kafkaSchedulerClient.Start()
		if err != nil {
			logger.WithError(err).Error("Start client failed")
		}
		logger.Infof("Scheduler client ended - closing done")
		close(done)
	}()

	// Wait for completion
	<-done

}
