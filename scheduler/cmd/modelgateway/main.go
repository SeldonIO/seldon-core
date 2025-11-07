/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	health_probe "github.com/seldonio/seldon-core/scheduler/v2/pkg/health-probe"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/gateway"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/schema"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/version"
)

const (
	flagSchedulerHost            = "scheduler-host"
	flagSchedulerPlaintxtPort    = "scheduler-plaintxt-port"
	flagSchedulerTlsPort         = "scheduler-tls-port"
	flagEnvoyHost                = "envoy-host"
	flagEnvoyPort                = "envoy-port"
	flagLogLevel                 = "log-level"
	flagHealthPort               = "health-probe-port"
	defaultSchedulerPlaintxtPort = 9004
	defaultSchedulerTLSPort      = 9044
	defaultEnvoyPort             = 9000
	defaultHealthProbePort       = 9999
	kubernetesNamespacePath      = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	displayVersion         bool
	schedulerHost          string
	schedulerPlaintxtPort  int
	schedulerTlsPort       int
	envoyHost              string
	envoyPort              int
	kafkaConfigPath        string
	namespace              string
	logLevel               string
	tracingConfigPath      string
	healthProbeServicePort int
	// TODO: add file watcher cfg using koanf and in the future read all file config in one file
	k = koanf.New(".")
)

func init() {
	flag.BoolVar(&displayVersion, "version", false, "display version and exit")
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
	flag.IntVar(&healthProbeServicePort, flagHealthPort, defaultHealthProbePort, "K8s health probe port")
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

	if displayVersion {
		logger.Infof("Version %s", version.Tag)
		os.Exit(0)
	}

	logIntLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", logLevel)
	}
	logger.Infof("Version %s", version.Tag)
	logger.Infof("Setting log level to %s", logLevel)
	logger.SetLevel(logIntLevel)

	updateNamespace()

	errChan := make(chan error, 5)

	tlsOptionsControlPlane, err := tls.CreateControlPlaneTLSOptions(
		tls.Prefix(tls.EnvSecurityPrefixControlPlaneClient),
		tls.ValidationPrefix(tls.EnvSecurityPrefixControlPlaneServer))
	if err != nil {
		logger.WithError(err).Fatal("Failed to create control-plane TLS Options")
	}

	tracer, err := tracing.NewTraceProvider("seldon-modelgateway", &tracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	kafkaConfigMap, err := kafka_config.NewKafkaConfig(kafkaConfigPath, logLevel)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load Kafka config")
	}

	schemaRegistryClient, err := schema.NewSchemaRegistryClient(logger, k)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create schema registry client")
	}

	if schemaRegistryClient != nil {
		logger.Infof("Schema registry client was set with host on %s", schemaRegistryClient.Config().SchemaRegistryURL)
	} else {
		logger.Debugf("Schema registry not set")
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
		WorkerTimeout:         getEnVar(logger, gateway.EnvVarWorkerTimeoutMs, gateway.DefaultWorkerTimeoutMs),
	}
	kafkaConsumer, err := gateway.NewKafkaConsumerManager(logger, &consumerConfig,
		getEnVar(logger, gateway.EnvMaxNumConsumers, gateway.DefaultMaxNumConsumers), schemaRegistryClient)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create consumer manager")
	}
	defer kafkaConsumer.Stop()

	kafkaSchedulerClient := gateway.NewKafkaSchedulerClient(logger, kafkaConsumer, tlsOptionsControlPlane)
	err = kafkaSchedulerClient.ConnectToScheduler(schedulerHost, schedulerPlaintxtPort, schedulerTlsPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to scheduler")
	}
	defer kafkaSchedulerClient.Stop()

	go func() {
		err := kafkaSchedulerClient.Start()
		if err != nil {
			logger.WithError(err).Error("Start client failed")
		}
		logger.Infof("Scheduler client ended")
		errChan <- err
	}()

	healthServer := initHealthProbeServer(logger, kafkaSchedulerClient, kafkaConsumer, errChan)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		if err := healthServer.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Health server shutdown failed")
		}
		cancel()
	}()

	// Wait for completion
	waitForTermSignalOrErr(logger, errChan)

	logger.Info("Graceful shutdown triggered")
}

func initHealthProbeServer(logger *log.Logger, schedulerClient *gateway.KafkaSchedulerClient, kafkaConsumer *gateway.KafkaConsumerManager, errChan chan<- error) *health_probe.HTTPServer {
	healthManager := health_probe.NewManager()

	healthManager.AddCheck(func() error {
		if !schedulerClient.IsConnected() {
			return errors.New("not connected to scheduler")
		}
		return nil
	}, health_probe.ProbeStartUp)

	healthManager.AddCheck(func() error {
		if err := kafkaConsumer.Healthy(); err != nil {
			return fmt.Errorf("kafka is not healthy: %w", err)
		}
		return nil
	}, health_probe.ProbeLiveness, health_probe.ProbeReadiness)

	healthServer := health_probe.NewHTTPServer(healthProbeServicePort, healthManager, logger)
	go func() {
		var err error
		if err = healthServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Error("HTTP health server failed")
		}
		errChan <- err
	}()

	return healthServer
}

func waitForTermSignalOrErr(logger *log.Logger, errChan <-chan error) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.WithError(err).Error("Shutting down due to error")
	case <-exit:
		logger.Info("Shutting down due to SIGTERM or SIGINT")
	}
}
