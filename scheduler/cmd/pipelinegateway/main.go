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
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	health "github.com/seldonio/seldon-core/scheduler/v2/pkg/health-probe"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/status"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/schema"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	"github.com/seldonio/seldon-core/scheduler/v2/version"
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
	flagHealthPort            = "health-probe-port"
)

const (
	defaultHttpPort              = 9010
	defaultGrpcPort              = 9011
	defaultMetricsPort           = 9006
	defaultSchedulerPlaintxtPort = 9004
	defaultSchedulerTLSPort      = 9044
	defaultEnvoyPort             = 9000
	defaultHealthProbePort       = 9999
	serviceTag                   = "seldon-pipelinegateway"
)

var (
	displayVersion         bool
	httpPort               int
	grpcPort               int
	metricsPort            int
	logLevel               string
	namespace              string
	kafkaConfigPath        string
	tracingConfigPath      string
	schedulerHost          string
	schedulerPlaintxtPort  int
	schedulerTlsPort       int
	envoyHost              string
	envoyPort              int
	healthProbeServicePort int
	// TODO: add file watcher cfg using koanf and in the future read all file config in one file
	k = koanf.New(".")
)

func init() {
	flag.BoolVar(&displayVersion, "version", false, "display version and exit")
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
	flag.IntVar(&healthProbeServicePort, flagHealthPort, defaultHealthProbePort, "Health probe port")
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

	defer logger.Info("Graceful shutdown complete")

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

	schemaRegistryClient, err := schema.NewSchemaRegistryClient(logger, k)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create schema registry client")
	}

	if schemaRegistryClient != nil {
		logger.Infof("Schema registry client was set with host on %s", schemaRegistryClient.Config().SchemaRegistryURL)
	} else {
		logger.Debugf("Schema registry not set")
	}

	maxNumConsumers := getEnVar(logger, pipeline.EnvMaxNumConsumers, pipeline.DefaultMaxNumConsumers)
	km, err := pipeline.NewKafkaManager(
		logger, namespace, kafkaConfigMap, tracer, maxNumConsumers, schemaRegistryClient)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create kafka manager")
	}
	defer km.Stop()

	errChan := make(chan error, 10)

	promMetrics, err := metrics.NewPrometheusPipelineMetrics(logger)
	if err != nil {
		logger.WithError(err).Fatalf("Can't create prometheus metrics")
	}
	go func() {
		err := promMetrics.Start(metricsPort)
		if !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Error("Can't start metrics server")
		}
		errChan <- err
	}()
	defer func() { _ = promMetrics.Stop() }()

	tlsEnvoyOptions, err := util.CreateUpstreamDataplaneServerTLSOptions()
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create TLS Options")
	}

	tlsOptions, err := tls.CreateControlPlaneTLSOptions(
		tls.Prefix(tls.EnvSecurityPrefixControlPlaneClient),
		tls.ValidationPrefix(tls.EnvSecurityPrefixControlPlaneServer))
	if err != nil {
		logger.WithError(err).Fatal("Failed to create TLS Options")
	}

	// Handle pipeline status updates
	statusManager := status.NewPipelineStatusManager()
	schedulerClient := pipeline.NewPipelineSchedulerClient(logger, statusManager, km, tlsOptions)
	go func() {
		if err := schedulerClient.Start(schedulerHost, schedulerPlaintxtPort, schedulerTlsPort); err != nil {
			logger.WithError(err).Error("Start client failed")
		}
		logger.Info("Scheduler client ended")
		errChan <- err
	}()

	restModelChecker, err := status.NewModelRestStatusCaller(logger, envoyHost, envoyPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create REST modelchecker")
	}
	pipelineReadyChecker := status.NewSimpleReadyChecker(statusManager, restModelChecker)

	grpcServer := pipeline.NewGatewayGrpcServer(grpcPort, logger, km, promMetrics, &tlsEnvoyOptions, pipelineReadyChecker)
	go func() {
		if err := grpcServer.Start(); err != nil {
			logger.WithError(err).Error("Failed to start grpc server")
			errChan <- err
		}
	}()

	httpServer := pipeline.NewGatewayHttpServer(httpPort, logger, km, promMetrics, &tlsEnvoyOptions, pipelineReadyChecker)
	go func() {
		if err := httpServer.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.WithError(err).Error("Failed to start http server")
				errChan <- err
			}
		}
	}()

	healthManager, err := initHealthProbe(schedulerClient, km, &tlsEnvoyOptions, logger, healthProbeServicePort, errChan, httpPort, httpServer)
	if err != nil {
		logger.WithError(err).Error("Failed to start http health server")
		errChan <- err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := healthManager.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Failed to shutdown health probe server")
		}
		cancel()
	}()

	waitForTermSignalOrErr(logger, errChan)

	logger.Infof("Shutting down http server")
	if err := httpServer.Stop(); err != nil {
		logger.WithError(err).Error("Failed to stop http server")
	}
	logger.Infof("Shutting down scheduler client")
	grpcServer.Stop()
	schedulerClient.Stop()
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

func initHealthProbe(schedulerClient *pipeline.PipelineSchedulerClient, kafka *pipeline.KafkaManager,
	tlsOptions *util.TLSOptions, log *log.Logger, listenPort int, errChan chan<- error, httpPort int, gwHTTPServer *pipeline.GatewayHttpServer) (*health.HTTPServer, error) {
	manager := health.NewManager()

	schedulerHealth(manager, schedulerClient)
	if err := gRPCHealth(tlsOptions, manager); err != nil {
		return nil, fmt.Errorf("failed setting up gRPC health probe: %w", err)
	}
	httpHealth(tlsOptions, manager, gwHTTPServer.HealthPath(), httpPort)
	kafkaHealth(manager, kafka)

	server := health.NewHTTPServer(listenPort, manager, log)
	go func() {
		errChan <- server.Start()
	}()

	return server, nil
}

func schedulerHealth(manager health.Manager, schedulerClient *pipeline.PipelineSchedulerClient) {
	manager.AddCheck(func() error {
		if !schedulerClient.IsConnected() {
			return fmt.Errorf("not connected to scheduler")
		}
		return nil
	}, health.ProbeStartUp)
}

func kafkaHealth(manager health.Manager, kafka *pipeline.KafkaManager) {
	manager.AddCheck(func() error {
		if kafka.ProducerClosed() {
			return fmt.Errorf("kafka producer closed")
		}
		if !kafka.ConsumersActive() {
			return fmt.Errorf("kafka consumer(s) not active")
		}
		return nil
	}, health.ProbeReadiness, health.ProbeLiveness)
}

func gRPCHealth(tlsOptions *util.TLSOptions, manager health.Manager) error {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.DefaultConfig,
	}), grpc.WithKeepaliveParams(util.GetClientKeepAliveParameters()))

	if tlsOptions.TLS {
		opts = append(opts, grpc.WithTransportCredentials(tlsOptions.Cert.CreateClientTransportCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// note this will not attempt connection handshake until req is sent
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", grpcPort), opts...)
	if err != nil {
		return fmt.Errorf("error creating gRPC connection: %v", err)
	}
	gRPCClient := v2_dataplane.NewGRPCInferenceServiceClient(conn)

	manager.AddCheck(func() error {
		_, err := gRPCClient.ServerReady(context.Background(), &v2_dataplane.ServerReadyRequest{})
		if err != nil {
			return fmt.Errorf("gRPC server check failed calling ServerReady: %w", err)
		}
		return nil
	}, health.ProbeReadiness, health.ProbeStartUp, health.ProbeLiveness)

	return nil
}

func httpHealth(tlsOptions *util.TLSOptions, manager health.Manager, path string, port int) {
	httpClient := &http.Client{}
	url := fmt.Sprintf("http://localhost:%d%s", port, path)
	if tlsOptions.TLS {
		url = fmt.Sprintf("https://localhost:%d%s", port, path)
		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsOptions.Cert.CreateClientTLSConfig(),
		}
	}

	manager.AddCheck(func() error {
		resp, err := httpClient.Get(url)
		if err != nil {
			return fmt.Errorf("failed HTTPs health request: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed HTTP health request got %d expected 200", resp.StatusCode)
		}
		return nil
	}, health.ProbeReadiness, health.ProbeStartUp, health.ProbeLiveness)
}
