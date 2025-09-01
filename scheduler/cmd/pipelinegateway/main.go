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

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	health "github.com/seldonio/seldon-core/scheduler/v2/pkg/health-probe"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/status"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	flagHttpPort              = "http-port"
	flagGrpcPort              = "grpc-port"
	flagLogLevel              = "log-level"
	flagMetricsPort           = "metrics-port"
	kubernetesNamespacePath   = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	flagSchedulerHost         = "scheduler-host"
	flagSchedulerPlaintxtPort = "scheduler-plaintxt-port"
	flagSchedulerTlsPort      = "scheduler-tls-port"
	flagEnvoyHost             = "envoy-host"
	flagEnvoyPort             = "envoy-port"
)

const (
	defaultHttpPort              = 9010
	defaultGrpcPort              = 9011
	defaultMetricsPort           = 9006
	defaultSchedulerPlaintxtPort = 9004
	defaultSchedulerTLSPort      = 9044
	defaultEnvoyPort             = 9000
	serviceTag                   = "seldon-pipelinegateway"
)

var (
	httpPort              int
	grpcPort              int
	metricsPort           int
	logLevel              string
	namespace             string
	kafkaConfigPath       string
	tracingConfigPath     string
	schedulerHost         string
	schedulerPlaintxtPort int
	schedulerTlsPort      int
	envoyHost             string
	envoyPort             int
)

func init() {

	flag.IntVar(&httpPort, flagHttpPort, defaultHttpPort, "http-port")
	flag.IntVar(&grpcPort, flagGrpcPort, defaultGrpcPort, "grpc-port")
	flag.IntVar(&metricsPort, flagMetricsPort, defaultMetricsPort, "metrics-port")
	flag.StringVar(&namespace, "namespace", "", "Namespace")
	flag.StringVar(&logLevel, flagLogLevel, "debug", "Log level - examples: debug, info, error")
	flag.StringVar(
		&kafkaConfigPath,
		"kafka-config-path",
		"/mnt/config/kafka.json",
		"path to kafka configuration file",
	)
	flag.StringVar(&tracingConfigPath, "tracing-config-path", "", "Tracing config path")
	flag.StringVar(&schedulerHost, flagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPlaintxtPort, flagSchedulerPlaintxtPort, defaultSchedulerPlaintxtPort, "Scheduler plaintxt port")
	flag.IntVar(&schedulerTlsPort, flagSchedulerTlsPort, defaultSchedulerTLSPort, "Scheduler TLS port")
	flag.StringVar(&envoyHost, flagEnvoyHost, "0.0.0.0", "Envoy host")
	flag.IntVar(&envoyPort, flagEnvoyPort, defaultEnvoyPort, "Envoy port")
}

func makeSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
}

// TODO: move to a common util
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

// TODO: move to a common util
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

	done := make(chan bool, 1)
	go makeSignalHandler(logger, done)

	updateNamespace()

	tracer, err := tracing.NewTraceProvider(serviceTag, &tracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	kafkaConfigMap, err := kafka_config.NewKafkaConfig(kafkaConfigPath, logLevel)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load Kafka config")
	}

	maxNumConsumers := getEnVar(logger, pipeline.EnvMaxNumConsumers, pipeline.DefaultMaxNumConsumers)
	km, err := pipeline.NewKafkaManager(
		logger, namespace, kafkaConfigMap, tracer, maxNumConsumers)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create kafka manager")
	}
	defer km.Stop()

	promMetrics, err := metrics.NewPrometheusPipelineMetrics(logger)
	if err != nil {
		logger.WithError(err).Fatalf("Can't create prometheus metrics")
	}
	go func() {
		err := promMetrics.Start(metricsPort)
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.WithError(err).Error("Can't start metrics server")
		close(done)
	}()
	defer func() { _ = promMetrics.Stop() }()

	tlsOptions, err := util.CreateUpstreamDataplaneServerTLSOptions()
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create TLS Options")
	}

	// Handle pipeline status updates
	statusManager := status.NewPipelineStatusManager()
	schedulerClient := pipeline.NewPipelineSchedulerClient(logger, statusManager, km)
	go func() {
		if err := schedulerClient.Start(schedulerHost, schedulerPlaintxtPort, schedulerTlsPort); err != nil {
			logger.WithError(err).Error("Start client failed")
		}
		logger.Info("Scheduler client ended - closing done")
		close(done)
	}()

	healthManager, err := initHealthProbe(schedulerClient)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create health probe")
	}

	restModelChecker, err := status.NewModelRestStatusCaller(logger, envoyHost, envoyPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create REST modelchecker")
	}
	pipelineReadyChecker := status.NewSimpleReadyChecker(statusManager, restModelChecker)

	grpcServer := pipeline.NewGatewayGrpcServer(grpcPort, logger, km, promMetrics, &tlsOptions, pipelineReadyChecker)
	go func() {
		if err := grpcServer.Start(); err != nil {
			logger.WithError(err).Error("Failed to start grpc server")
			close(done)
		}
	}()

	httpServer := pipeline.NewGatewayHttpServer(httpPort, logger, km, promMetrics, &tlsOptions, pipelineReadyChecker, healthManager)
	go func() {
		if err := httpServer.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.WithError(err).Error("Failed to start http server")
				close(done)
			}
		}
	}()

	// Wait for completion
	<-done
	logger.Infof("Shutting down http server")
	if err := httpServer.Stop(); err != nil {
		logger.WithError(err).Error("Failed to stop http server")
	}
	logger.Infof("Shutting down scheduler client")
	grpcServer.Stop()
	schedulerClient.Stop()
}

func initHealthProbe(schedulerClient *pipeline.PipelineSchedulerClient) (health.Manager, error) {
	manager := health.NewManager()
	manager.AddCheck(func() error {
		if !schedulerClient.IsConnected() {
			return fmt.Errorf("not connected to scheduler")
		}
		return nil
	}, health.ProbeStartUp)

	// note this will not attempt connection handshake until req is sent
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", grpcPort),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}),
		grpc.WithKeepaliveParams(util.GetClientKeepAliveParameters()),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("error creating gRPC connection: %v", err)
	}
	gRPCClient := v2_dataplane.NewGRPCInferenceServiceClient(conn)

	manager.AddCheck(func() error {
		_, err := gRPCClient.ServerReady(context.Background(), &v2_dataplane.ServerReadyRequest{})
		if err != nil {
			return fmt.Errorf("gRPC server check failed calling ServerReady: %w", err)
		}
		return nil
	}, health.ProbeReadiness, health.ProbeStartUp, health.ProbeLiveness)

	return manager, nil
}
