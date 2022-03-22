package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/gateway"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
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
	schedulerHost string
	schedulerPort int
	envoyHost     string
	envoyPort     int
	configPath    string
	namespace     string
	logLevel      string
)

func init() {

	flag.StringVar(&schedulerHost, flagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPort, flagSchedulerPort, defaultSchedulerPort, "Scheduler port")
	flag.StringVar(&envoyHost, flagEnvoyHost, "0.0.0.0", "Envoy host")
	flag.IntVar(&envoyPort, flagEnvoyPort, defaultEnvoyPort, "Envoy port")
	flag.StringVar(&namespace, "namespace", "", "Namespace")
	flag.StringVar(
		&configPath,
		"config-path",
		"/mnt/config",
		"Path to folder with configuration files. Will assume agent.yaml or agent.json in this folder",
	)
	flag.StringVar(&logLevel, flagLogLevel, "debug", "Log level - examples: debug, info, error")

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

	var clientset kubernetes.Interface
	if runningInsideK8s() {
		clientset, err = k8s.CreateClientset()
		if err != nil { //TODO change to Error from Fatal?
			logger.WithError(err).Fatal("Failed to create kubernetes clientset")
		}
	}

	// Start Agent configuration handler
	agentConfigHandler, err := config.NewAgentConfigHandler(configPath, namespace, logger, clientset)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create stream config handler")
	}
	defer func() {
		_ = agentConfigHandler.Close()
		logger.Info("Closed config handler")
	}()

	kafkaManager := gateway.NewKafkaManager(logger, &gateway.KafkaServerConfig{
		Host:     envoyHost,
		HttpPort: envoyPort,
		GrpcPort: envoyPort,
	})
	defer func() { _ = kafkaManager.Stop() }()

	kafkaManager.StartConfigListener(agentConfigHandler)

	kafkaSchedulerClient := gateway.NewKafkaSchedulerClient(logger, kafkaManager)
	err = kafkaSchedulerClient.ConnectToScheduler(schedulerHost, schedulerPort)
	if err != nil {
		logger.Fatalf("Failed to connect to scheduler")
	}

	go func() {
		err := kafkaSchedulerClient.SubscribeModelEvents(context.Background())
		if err != nil {
			logger.WithError(err).Error("Subscribe model events failed")
			close(done)
		}
	}()

	// Wait for completion
	<-done
}
