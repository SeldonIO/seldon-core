package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/seldonio/seldon-core/scheduler/pkg/otel"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/pipeline"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	flagHttpPort            = "http-port"
	flagGrpcPort            = "grpc-port"
	flagLogLevel            = "log-level"
	kubernetesNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	httpPort   int
	grpcPort   int
	logLevel   string
	namespace  string
	configPath string
)

func init() {

	flag.IntVar(&httpPort, flagHttpPort, 9010, "http-port")
	flag.IntVar(&grpcPort, flagGrpcPort, 9011, "grpc-port")
	flag.StringVar(&namespace, "namespace", "", "Namespace")
	flag.StringVar(&logLevel, flagLogLevel, "debug", "Log level - examples: debug, info, error")
	flag.StringVar(
		&configPath,
		"config-path",
		"/mnt/config",
		"Path to folder with configuration files. Will assume agent.yaml or agent.json in this folder",
	)
}

func makeSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
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

	var clientset kubernetes.Interface
	if runningInsideK8s() {
		clientset, err = k8s.CreateClientset()
		if err != nil { //TODO change to Error from Fatal?
			logger.WithError(err).Fatal("Failed to create kubernetes clientset")
		}
	}

	tracer, err := otel.NewTracer("seldon-pipelinegateway")
	if err != nil {
		logger.WithError(err).Error("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
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

	km := pipeline.NewKafkaManager(logger, namespace)
	km.StartConfigListener(agentConfigHandler)

	httpServer := pipeline.NewGatewayHttpServer(httpPort, logger, nil, km)
	go func() {
		if err := httpServer.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.WithError(err).Error("Failed to start http server")
				close(done)
			}
		}
	}()

	grpcServer := pipeline.NewGatewayGrpcServer(grpcPort, logger, km)
	go func() {
		if err := grpcServer.Start(); err != nil {
			logger.WithError(err).Error("Failed to start grpc server")
			close(done)
		}
	}()

	// Wait for completion
	<-done
	logger.Infof("Shutting down http server")
	if err := httpServer.Stop(); err != nil {
		logger.WithError(err).Error("Failed to stop http server")
	}
	logger.Infof("Shutting down grpc server")
	grpcServer.Stop()

}
