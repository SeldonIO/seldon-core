package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	"github.com/seldonio/seldon-core/scheduler/pkg/tracing"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/gateway"

	log "github.com/sirupsen/logrus"
)

const (
	flagSchedulerHost       = "scheduler-host"
	flagSchedulerPort       = "scheduler-port"
	flagEnvoyHost           = "envoy-host"
	flagEnvoyPort           = "envoy-port"
	flagLogLevel            = "log-level"
	defaultSchedulerPort    = 9004
	defaultEnvoyPort        = 9000
	kubernetesNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	schedulerHost     string
	schedulerPort     int
	envoyHost         string
	envoyPort         int
	kafkaConfigPath   string
	namespace         string
	logLevel          string
	tracingConfigPath string
)

func init() {

	flag.StringVar(&schedulerHost, flagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPort, flagSchedulerPort, defaultSchedulerPort, "Scheduler port")
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
	nsBytes, err := ioutil.ReadFile(kubernetesNamespacePath)
	if err != nil {
		log.Warn("Using namespace from command line argument")
	} else {
		ns := string(nsBytes)
		log.Infof("Setting namespace from k8s file to %s", ns)
		namespace = ns
	}
}

func runningInsideK8s() bool {
	return namespace != ""
}

func makeSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
}

func getNumberWorkers(logger *log.Logger) int {
	valStr := os.Getenv(gateway.EnvVarNumWorkers)
	if valStr != "" {
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to parse %s", gateway.EnvVarNumWorkers)
		}
		logger.Infof("Setting number of workers to %d", val)
		return int(val)
	}
	logger.Infof("Setting number of workers to default %d", gateway.DefaultNumWorkers)
	return gateway.DefaultNumWorkers
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
		logger.WithError(err).Error("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	kafkaConfigMap, err := config.NewKafkaConfig(kafkaConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load Kafka config")
	}

	inferServerConfig := &gateway.InferenceServerConfig{
		Host:     envoyHost,
		HttpPort: envoyPort,
		GrpcPort: envoyPort,
	}
	kafkaConsumer, err := gateway.NewInferKafkaConsumer(logger, getNumberWorkers(logger), kafkaConfigMap, namespace, inferServerConfig, tracer)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create kafka consumer")
	}
	go kafkaConsumer.Serve()
	defer kafkaConsumer.Stop()

	kafkaSchedulerClient := gateway.NewKafkaSchedulerClient(logger, kafkaConsumer)
	err = kafkaSchedulerClient.ConnectToScheduler(schedulerHost, schedulerPort)
	if err != nil {
		logger.Fatalf("Failed to connect to scheduler")
	}

	go func() {
		err := kafkaSchedulerClient.Start()
		if err != nil {
			logger.WithError(err).Error("Start client failed")
			close(done)
		}
	}()

	// Wait for completion
	<-done
}
