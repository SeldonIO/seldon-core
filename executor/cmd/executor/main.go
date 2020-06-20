package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api"
	seldonclient "github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/grpc/tensorflow"
	"github.com/seldonio/seldon-core/executor/api/kafka"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/api/tracing"
	"github.com/seldonio/seldon-core/executor/api/util"
	"github.com/seldonio/seldon-core/executor/k8s"
	loghandler "github.com/seldonio/seldon-core/executor/logger"
	predictor2 "github.com/seldonio/seldon-core/executor/predictor"
	"github.com/seldonio/seldon-core/executor/proto/tensorflow/serving"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	zapf "sigs.k8s.io/controller-runtime/pkg/log/zap"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

const (
	logLevelEnvVar          = "SELDON_LOG_LEVEL"
	logLevelDefault         = "INFO"
	debugEnvVar             = "SELDON_DEBUG"
	ENV_VAR_CERT_MOUNT_PATH = "SELDON_CERT_MOUNT_PATH"
)

var (
	serverType = flag.String("server_type", "rpc", "Server type: rpc or kafka")

	debugDefault = false

	configPath     = flag.String("config", "", "Path to kubconfig")
	sdepName       = flag.String("sdep", "", "Seldon deployment name")
	namespace      = flag.String("namespace", "", "Namespace")
	predictorName  = flag.String("predictor", "", "Name of the predictor inside the SeldonDeployment")
	port           = flag.Int("port", 8080, "Executor port")
	wait           = flag.Duration("graceful_timeout", time.Second*15, "Graceful shutdown secs")
	protocol       = flag.String("protocol", "seldon", "The payload protocol")
	transport      = flag.String("transport", "rest", "The network transport mechanism rest, grpc")
	filename       = flag.String("file", "", "Load graph from file")
	hostname       = flag.String("hostname", "", "The hostname of the running server")
	logWorkers     = flag.Int("logger_workers", 5, "Number of workers handling payload logging")
	prometheusPath = flag.String("prometheus_path", "/metrics", "The prometheus metrics path")
	kafkaBroker    = flag.String("kafka_broker", "", "The kafka broker as host:port")
	kafkaTopicIn   = flag.String("kafka_input_topic", "", "The kafka input topic")
	kafkaTopicOut  = flag.String("kafka_output_topic", "", "The kafka output topic")
	kafkaFullGraph = flag.Bool("kafka_full_graph", false, "Use kafka for internal graph processing")
	kafkaWorkers   = flag.Int("kafka_workers", 4, "Number of kafka workers")
	debug          = flag.Bool(
		"debug",
		util.GetEnvAsBool(debugEnvVar, debugDefault),
		"Enable debug mode. Logs will be sampled and less structured.",
	)
	logLevel = flag.String(
		"log_level",
		util.GetEnv(logLevelEnvVar, logLevelDefault),
		"Log level.",
	)

	envCertMountPath = util.GetEnv(ENV_VAR_CERT_MOUNT_PATH, "")
)

func getServerUrl(hostname string, port int) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("http://%s:%d/", hostname, port))
}

func runHttpServer(lis net.Listener, logger logr.Logger, predictor *v1.PredictorSpec, client seldonclient.SeldonApiClient, port int,
	probesOnly bool, serverUrl *url.URL, namespace string, protocol string, deploymentName string, prometheusPath string) {

	// Create REST API
	seldonRest := rest.NewServerRestApi(predictor, client, probesOnly, serverUrl, namespace, protocol, deploymentName, prometheusPath)
	seldonRest.Initialise()
	srv := seldonRest.CreateHttpServer(port)

	go func() {
		if err := srv.Serve(lis); err != nil {
			logger.Error(err, "Server error")
		}
		logger.Info("server started")
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) and SIGTERM
	// SIGKILL, SIGQUIT will not be caught.
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), *wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logger.Info("shutting down")
	os.Exit(0)

}

func runGrpcServer(lis net.Listener, logger logr.Logger, predictor *v1.PredictorSpec, client seldonclient.SeldonApiClient, port int, serverUrl *url.URL, namespace string, protocol string, deploymentName string, annotations map[string]string) {
	grpcServer, err := grpc.CreateGrpcServer(predictor, deploymentName, annotations, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}
	if protocol == api.ProtocolSeldon {
		seldonGrpcServer := seldon.NewGrpcSeldonServer(predictor, client, serverUrl, namespace)
		proto.RegisterSeldonServer(grpcServer, seldonGrpcServer)
	} else {
		tensorflowGrpcServer := tensorflow.NewGrpcTensorflowServer(predictor, client, serverUrl, namespace)
		serving.RegisterPredictionServiceServer(grpcServer, tensorflowGrpcServer)
		serving.RegisterModelServiceServer(grpcServer, tensorflowGrpcServer)
	}
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Error(err, "gRPC server error")
	}
}

func setupLogger() {
	level := zap.InfoLevel
	switch *logLevel {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
	case "WARNING":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	case "FATAL":
		level = zap.FatalLevel
	}

	atomicLevel := zap.NewAtomicLevelAt(level)

	logger := zapf.New(
		zapf.UseDevMode(*debug),
		zapf.Level(&atomicLevel),
	)

	logf.SetLogger(logger)
}

func main() {
	flag.Parse()

	if *sdepName == "" {
		log.Fatal("Required argument sdep missing")
	}

	if *namespace == "" {
		log.Fatal("Required argument namespace missing")
	}

	if *predictorName == "" {
		log.Fatal("Required argument predictor missing")
	}

	if !(*protocol == api.ProtocolSeldon || *protocol == api.ProtocolTensorflow) {
		log.Fatal("Invalid protocol: must be seldon or tensorflow")
	}

	if *serverType == "kafka" {
		// Get Broker
		if *kafkaBroker == "" {
			*kafkaBroker = os.Getenv(kafka.ENV_KAFKA_BROKER)
			if *kafkaBroker == "" {
				log.Fatal("Required argument kafka_broker missing")
			}
		}
		// Get input topic
		if *kafkaTopicIn == "" {
			*kafkaTopicIn = os.Getenv(kafka.ENV_KAFKA_INPUT_TOPIC)
			if *kafkaTopicIn == "" {
				log.Fatal("Required argument kafka_input_topic missing")
			}
		}
		// Get output topic
		if *kafkaTopicOut == "" {
			*kafkaTopicOut = os.Getenv(kafka.ENV_KAFKA_OUTPUT_TOPIC)
			if *kafkaTopicOut == "" {
				log.Fatal("Required argument kafka_output_topic missing")
			}
		}
		// Get Full Graph
		kafkaFullGraphFromEnv := os.Getenv(kafka.ENV_KAFKA_FULL_GRAPH)
		if kafkaFullGraphFromEnv != "" {
			kafkaFullGraphFromEnvBool, err := strconv.ParseBool(kafkaFullGraphFromEnv)
			if err != nil {
				log.Fatalf("Failed to parse %s %s", kafka.ENV_KAFKA_FULL_GRAPH, kafkaFullGraphFromEnv)
			} else {
				*kafkaFullGraph = kafkaFullGraphFromEnvBool
			}
		}

		//Kafka workers
		kafkaWorkersFromEnv := os.Getenv(kafka.ENV_KAFKA_WORKERS)
		if kafkaWorkersFromEnv != "" {
			kafkaWorkersFromEnvInt, err := strconv.Atoi(kafkaWorkersFromEnv)
			if err != nil {
				log.Fatalf("Failed to parse %s %s", kafka.ENV_KAFKA_WORKERS, kafkaWorkersFromEnv)
			} else {
				*kafkaWorkers = kafkaWorkersFromEnvInt
			}
		}

	}

	setupLogger()
	logger := logf.Log.WithName("entrypoint")

	// Set hostname
	if *hostname == "" {
		*hostname = os.Getenv("POD_NAME")
		if *hostname == "" {
			logger.Info("Hostname unset will use localhost")
			*hostname = "localhost"
		} else {
			logger.Info("Hostname found from env", "hostname", *hostname)
		}
	} else {
		logger.Info("Hostname provided on command line", "hostname", *hostname)
	}
	serverUrl, err := getServerUrl(*hostname, *port)
	if err != nil {
		log.Fatal("Failed to create server url from", *hostname, *port)
	}

	predictor, err := predictor2.GetPredictor(*predictorName, *filename, *sdepName, *namespace, configPath)
	if err != nil {
		logger.Error(err, "Failed to get predictor")
		os.Exit(-1)

	}

	// Ensure standard OpenAPI seldon API file has this deployment's values
	err = rest.EmbedSeldonDeploymentValuesInSwaggerFile(*namespace, *sdepName)
	if err != nil {
		logger.Error(err, "Failed to embed variables on OpenAPI template")
	}

	annotations, err := k8s.GetAnnotations()
	if err != nil {
		logger.Error(err, "Failed to load annotations")
	}

	//Start Logger Dispacther
	loghandler.StartDispatcher(*logWorkers, logger, *sdepName, *namespace, *predictorName)

	//Init Tracing
	closer, err := tracing.InitTracing()
	if err != nil {
		log.Fatal("Could not initialize jaeger tracer", err.Error())
	}
	defer closer.Close()
	// Create a listener at the desired port.
	var lis net.Listener
	if len(envCertMountPath) > 0 {
		logger.Info("Creating TLS listener", "port", *port)
		certPath := path.Join(envCertMountPath, "tls.crt")
		keyPath := path.Join(envCertMountPath, "tls.key")
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			log.Fatalf("Error certificate could not be found: %v", err)
		}
		lis, err = tls.Listen("tcp", fmt.Sprintf(":%d", *port), &tls.Config{Certificates: []tls.Certificate{cert}})
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	} else {
		logger.Info("Creating non-TLS listener", "port", *port)
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	}
	defer lis.Close()

	// Create a cmux object.
	tcpm := cmux.New(lis)

	// Declare the match for different services required.
	httpl := tcpm.Match(cmux.HTTP1Fast())
	grpcl := tcpm.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	logger.Info("Running grpc server ", "port", *port)
	var clientGrpc seldonclient.SeldonApiClient
	if *protocol == "seldon" {
		clientGrpc = seldon.NewSeldonGrpcClient(predictor, *sdepName, annotations)
	} else {
		clientGrpc = tensorflow.NewTensorflowGrpcClient(predictor, *sdepName, annotations)
	}
	go runGrpcServer(grpcl, logger, predictor, clientGrpc, *port, serverUrl, *namespace, *protocol, *sdepName, annotations)

	clientRest, err := rest.NewJSONRestClient(*protocol, *sdepName, predictor, annotations)
	if err != nil {
		log.Fatalf("Failed to create http client: %v", err)
	}
	logger.Info("Running http server ", "port", *port)
	go runHttpServer(httpl, logger, predictor, clientRest, *port, false, serverUrl, *namespace, *protocol, *sdepName, *prometheusPath)

	if *serverType == "kafka" {
		logger.Info("Starting kafka server")
		kafkaServer, err := kafka.NewKafkaServer(*kafkaFullGraph, *kafkaWorkers, *sdepName, *namespace, *protocol, *transport, annotations, serverUrl, predictor, *kafkaBroker, *kafkaTopicIn, *kafkaTopicOut, logger)
		if err != nil {
			log.Fatalf("Failed to create kafka server: %v", err)
		}
		go func() {
			err = kafkaServer.Serve()
			if err != nil {
				log.Fatal("Failed to serve kafka", err)
			}
		}()
	}

	// Start cmux serving.
	if err := tcpm.Serve(); !strings.Contains(err.Error(),
		"use of closed network connection") {
		log.Fatal(err)
	}

}
