package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"strconv"

	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api"
	seldonclient "github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/kfserving"
	kfproto "github.com/seldonio/seldon-core/executor/api/grpc/kfserving/inference"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/grpc/tensorflow"
	"github.com/seldonio/seldon-core/executor/api/kafka"
	"github.com/seldonio/seldon-core/executor/api/rabbitmq"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/api/tracing"
	"github.com/seldonio/seldon-core/executor/api/util"
	"github.com/seldonio/seldon-core/executor/k8s"
	loghandler "github.com/seldonio/seldon-core/executor/logger"
	predictor2 "github.com/seldonio/seldon-core/executor/predictor"
	"github.com/seldonio/seldon-core/executor/proto/tensorflow/serving"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	zapf "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	logLevelEnvVar        = "SELDON_LOG_LEVEL"
	logLevelDefault       = "INFO"
	debugEnvVar           = "SELDON_DEBUG"
	certMountPathEnvVar   = "SELDON_CERT_MOUNT_PATH"
	certFileEnvVar        = "SELDON_CERT_FILE_NAME"
	certKeyFileNameEnvVar = "SELDON_CERT_KEY_FILE_NAME"
)

var (
	serverType = flag.String("server_type", "rpc", "Server type: rpc or kafka")

	debugDefault = false

	configPath        = flag.String("config", "", "Path to kubconfig")
	sdepName          = flag.String("sdep", "", "Seldon deployment name")
	namespace         = flag.String("namespace", "", "Namespace")
	predictorName     = flag.String("predictor", "", "Name of the predictor inside the SeldonDeployment")
	httpPort          = flag.Int("http_port", 8080, "Executor http port")
	grpcPort          = flag.Int("grpc_port", 5000, "Executor grpc port")
	wait              = flag.Duration("graceful_timeout", time.Second*15, "Graceful shutdown secs")
	delay             = flag.Duration("shutdown_delay", 0, "Shutdown delay secs")
	protocol          = flag.String("protocol", "seldon", "The payload protocol")
	transport         = flag.String("transport", "rest", "The network transport mechanism rest, grpc")
	filename          = flag.String("file", "", "Load graph from file")
	hostname          = flag.String("hostname", "", "The hostname of the running server")
	logWorkers        = flag.Int("logger_workers", 10, "Number of workers handling payload logging")
	logWorkBufferSize = flag.Int("log_work_buffer_size", loghandler.DefaultWorkQueueSize, "Limit of buffered logs in memory while waiting for downstream request ingestion")
	logWriteTimeoutMs = flag.Int("log_write_timeout_ms", loghandler.DefaultWriteTimeoutMilliseconds, "Timeout before giving up writing log if buffer is full. If <= 0 will immediately drop log on full log buffer.")
	prometheusPath    = flag.String("prometheus_path", "/metrics", "The prometheus metrics path")
	kafkaBroker       = flag.String("kafka_broker", "", "The kafka broker as host:port")
	kafkaTopicIn      = flag.String("kafka_input_topic", "", "The kafka input topic")
	kafkaTopicOut     = flag.String("kafka_output_topic", "", "The kafka output topic")
	kafkaFullGraph    = flag.Bool("kafka_full_graph", false, "Use kafka for internal graph processing")
	kafkaWorkers      = flag.Int("kafka_workers", 4, "Number of kafka workers")
	kafkaAutoCommit   = flag.Bool("kafka_auto_commit", true, "Use auto committing in the kafka consumer")
	logKafkaBroker    = flag.String("log_kafka_broker", "", "The kafka log broker")
	logKafkaTopic     = flag.String("log_kafka_topic", "", "The kafka log topic")
	rabbitmqFullGraph = flag.Bool("rabbitmq_full_graph", false, "Use rabbitmq for internal graph processing")
	rabbitmqBroker    = flag.String("rabbitmq_broker", "", "The rabbitmq broker as host:port")
	rabbitmqQueueIn   = flag.String("rabbitmq_input_queue", "", "The rabbitmq input queue")
	rabbitmqQueueOut  = flag.String("rabbitmq_output_queue", "", "The rabbitmq output queue")
	fullHealthChecks  = flag.Bool("full_health_checks", false, "Full health checks via chosen protocol API")
	debug             = flag.Bool(
		"debug",
		util.GetEnvAsBool(debugEnvVar, debugDefault),
		"Enable debug mode. Logs will be sampled and less structured.",
	)
	logLevel = flag.String(
		"log_level",
		util.GetEnv(logLevelEnvVar, logLevelDefault),
		"Log level.",
	)

	certMountPath   = util.GetEnv(certMountPathEnvVar, "")
	certFileName    = util.GetEnv(certFileEnvVar, "tls.crt")
	certKeyFileName = util.GetEnv(certFileEnvVar, "tls.key")
)

func getServerUrl(hostname string, port int) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("http://%s:%d/", hostname, port))
}

func runHttpServer(wg *sync.WaitGroup, shutdown chan bool, lis net.Listener, logger logr.Logger, predictor *v1.PredictorSpec, client seldonclient.SeldonApiClient, port int, probesOnly bool, serverUrl *url.URL, namespace string, protocol string, deploymentName string, prometheusPath string, fullHealthChecks bool) {
	wg.Add(1)
	defer wg.Done()
	defer lis.Close()

	// Create REST API
	seldonRest := rest.NewServerRestApi(predictor, client, probesOnly, serverUrl, namespace, protocol, deploymentName, prometheusPath, fullHealthChecks)
	seldonRest.Initialise()
	srv := seldonRest.CreateHttpServer(port)

	var isShutdown bool
	go func() {
		logger.Info("http server started")
		if err := srv.Serve(lis); err != nil {
			if !(isShutdown && errors.Is(err, http.ErrServerClosed)) {
				logger.Error(err, "http server error")
			}
		}
	}()

	// Block until we receive our signal.
	isShutdown = <-shutdown

	// Doesn't block if no connections, otherwise this waits for context
	// deadline. As shutdown is coordinated across multiple server by the
	// caller, we just use a background context and wait indefinitely if there
	// are active connections.
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error(err, "http server shutdown error")
	}
	logger.Info("http server shutdown")
}

func runGrpcServer(wg *sync.WaitGroup, shutdown chan bool, lis net.Listener, logger logr.Logger, predictor *v1.PredictorSpec, client seldonclient.SeldonApiClient, serverUrl *url.URL, namespace string, protocol string, deploymentName string, annotations map[string]string) {
	wg.Add(1)
	defer wg.Done()
	defer lis.Close()
	grpcServer, err := grpc.CreateGrpcServer(predictor, deploymentName, annotations, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}
	switch protocol {
	case api.ProtocolSeldon:
		seldonGrpcServer := seldon.NewGrpcSeldonServer(predictor, client, serverUrl, namespace)
		proto.RegisterSeldonServer(grpcServer, seldonGrpcServer)
		// Register reflection service on gRPC server.
		reflection.Register(grpcServer)
	case api.ProtocolTensorflow:
		tensorflowGrpcServer := tensorflow.NewGrpcTensorflowServer(predictor, client, serverUrl, namespace)
		serving.RegisterPredictionServiceServer(grpcServer, tensorflowGrpcServer)
		serving.RegisterModelServiceServer(grpcServer, tensorflowGrpcServer)
	case api.ProtocolV2, api.ProtocolKFServing:
		kfservingGrpcServer := kfserving.NewGrpcKFServingServer(predictor, client, serverUrl, namespace)
		kfproto.RegisterGRPCInferenceServiceServer(grpcServer, kfservingGrpcServer)
	}

	go func() {
		logger.Info("gRPC server started")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(err, "gRPC server error")
		}
	}()

	<-shutdown // wait for signal
	grpcServer.GracefulStop()
	logger.Info("gRPC server shutdown")
}

func waitForShutdown(logger logr.Logger, wg *sync.WaitGroup, chs ...chan bool) {
	// wait for and then signal servers have exited
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) and SIGTERM
	// SIGKILL, SIGQUIT will not be caught.
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	// Block until we receive our signal.
	sig := <-c
	logger.Info("shutdown signal received", "signal", sig, "shutdown_delay", *delay)
	time.Sleep(*delay) // shutdown_delay

	// Create a deadline to wait for graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), *wait)
	defer cancel()

	// send signals to server channels to initiate shutdown.
	for _, ch := range chs {
		ch <- true
	}

	select {
	case <-done:
		return // servers shutdown
	case <-ctx.Done():
		logger.Error(ctx.Err(), "graceful_timeout exceeded, exiting immediately", "graceful_timeout", *wait)
		os.Exit(1)
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

	if !(*protocol == api.ProtocolSeldon || *protocol == api.ProtocolTensorflow || *protocol == api.ProtocolV2 || *protocol == api.ProtocolKFServing) {
		log.Fatal("Protocol must be seldon, tensorflow or v2")
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

		// Get Kafka Auto Commit
		kafkaAutoCommitFromEnv := os.Getenv(kafka.ENV_KAFKA_AUTO_COMMIT)
		if kafkaAutoCommitFromEnv != "" {
			kafkaAutoCommitFromEnvBool, err := strconv.ParseBool(kafkaAutoCommitFromEnv)
			if err != nil {
				log.Fatalf("Failed to parse %s %s", kafka.ENV_KAFKA_AUTO_COMMIT, kafkaAutoCommitFromEnv)
			} else {
				*kafkaAutoCommit = kafkaAutoCommitFromEnvBool
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

	if *serverType == "rabbitmq" {
		// Get Broker
		if *rabbitmqBroker == "" {
			*rabbitmqBroker = os.Getenv(rabbitmq.ENV_RABBITMQ_BROKER_URL)
			if *rabbitmqBroker == "" {
				log.Fatal("Required argument rabbitmq_broker missing")
			}
		}
		// Get input topic
		if *rabbitmqQueueIn == "" {
			*rabbitmqQueueIn = os.Getenv(rabbitmq.ENV_RABBITMQ_INPUT_QUEUE)
			if *rabbitmqQueueIn == "" {
				log.Fatal("Required argument rabbitmq_input_queue missing")
			}
		}
		// Get output queue
		if *rabbitmqQueueOut == "" {
			*rabbitmqQueueOut = os.Getenv(rabbitmq.ENV_RABBITMQ_OUTPUT_QUEUE)
			if *rabbitmqQueueOut == "" {
				log.Fatal("Required argument rabbitmq_output_topic missing")
			}
		}
		// Get Full Graph
		rabbitmqFullGraphFromEnv := os.Getenv(rabbitmq.ENV_RABBITMQ_FULL_GRAPH)
		if rabbitmqFullGraphFromEnv != "" {
			rabbitmqFullGraphFromEnvBool, err := strconv.ParseBool(rabbitmqFullGraphFromEnv)
			if err != nil {
				log.Fatalf("Failed to parse %s %s", rabbitmq.ENV_RABBITMQ_FULL_GRAPH, rabbitmqFullGraphFromEnv)
			} else {
				*rabbitmqFullGraph = rabbitmqFullGraphFromEnvBool
			}
		}
	}

	if !(*transport == "rest" || *transport == "grpc") {
		log.Fatal("Only rest and grpc supported")
	}

	serverUrl, err := getServerUrl(*hostname, *httpPort)
	if err != nil {
		log.Fatal("Failed to create server url from", *hostname, *httpPort)
	}

	setupLogger()
	logger := logf.Log.WithName("entrypoint")

	logger.Info("Full health checks ", "value", fullHealthChecks)

	// Set runtime.GOMAXPROCS to respect container limits if the env var GOMAXPROCS is not set or is invalid, preventing CPU throttling.
	undo, err := maxprocs.Set(maxprocs.Logger(func(format string, a ...interface{}) {
		logger.WithName("maxprocs").Info(fmt.Sprintf(format, a...))
	}))
	defer undo()
	if err != nil {
		logger.Error(err, "failed to set GOMAXPROCS")
	}

	// Set hostname
	if *hostname == "" {
		*hostname = os.Getenv("POD_NAME")
		if *hostname == "" {
			logger.Info("Hostname unset will use localhost")
			*hostname = "localhost"
		} else {
			logger.Info("Hostname found from env", "hostname", *hostname)
		}
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
	err = loghandler.StartDispatcher(*logWorkers, *logWorkBufferSize, *logWriteTimeoutMs, logger, *sdepName, *namespace, *predictorName, *logKafkaBroker, *logKafkaTopic, *protocol)
	if err != nil {
		log.Fatal("Failed to start log dispatcher", err)
	}

	//Init Tracing
	closer, err := tracing.InitTracing()
	if err != nil {
		log.Fatal("Could not initialize jaeger tracer", err.Error())
	}
	defer closer.Close()

	if *serverType == "kafka" {
		logger.Info("Starting kafka server")
		kafkaServer, err := kafka.NewKafkaServer(*kafkaFullGraph, *kafkaWorkers, *sdepName, *namespace, *protocol, *transport, annotations, serverUrl, predictor, *kafkaBroker, *kafkaTopicIn, *kafkaTopicOut, logger, *fullHealthChecks, *kafkaAutoCommit)
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

	if *serverType == "rabbitmq" {
		logger.Info("Starting rabbitmq server")
		rabbitMqServer, err := rabbitmq.NewRabbitMqServer(*rabbitmqFullGraph, *sdepName, *namespace, *protocol, *transport, annotations, serverUrl, predictor, *rabbitmqBroker, *rabbitmqQueueIn, *rabbitmqQueueOut, logger, *fullHealthChecks)
		if err != nil {
			log.Fatalf("Failed to create rabbitmq server: %v", err)
		}
		go func() {
			err = rabbitMqServer.Serve()
			if err != nil {
				log.Fatal("Failed to serve rabbitmq", err)
			}
		}()
	}

	clientRest, err := rest.NewJSONRestClient(*protocol, *sdepName, predictor, annotations)
	if err != nil {
		log.Fatalf("Failed to create http client: %v", err)
	}

	var clientGrpc seldonclient.SeldonApiClient
	switch *protocol {
	case api.ProtocolSeldon:
		clientGrpc = seldon.NewSeldonGrpcClient(predictor, *sdepName, annotations)
	case api.ProtocolTensorflow:
		clientGrpc = tensorflow.NewTensorflowGrpcClient(predictor, *sdepName, annotations)
	case api.ProtocolV2, api.ProtocolKFServing:
		clientGrpc = kfserving.NewKFServingGrpcClient(predictor, *sdepName, annotations)
	default:
		log.Fatalf("Failed to create grpc client. Unknown protocol %s: %v", *protocol, err)
	}

	wg := sync.WaitGroup{}
	logger.Info("Running http server ", "port", *httpPort)
	httpStop := make(chan bool, 1)
	go runHttpServer(&wg, httpStop, createListener(*httpPort, logger), logger, predictor, clientRest, *httpPort, false, serverUrl, *namespace, *protocol, *sdepName, *prometheusPath, *fullHealthChecks)

	logger.Info("Running grpc server ", "port", *grpcPort)
	grpcStop := make(chan bool, 1)
	go runGrpcServer(&wg, grpcStop, createListener(*grpcPort, logger), logger, predictor, clientGrpc, serverUrl, *namespace, *protocol, *sdepName, annotations)
	waitForShutdown(logger, &wg, httpStop, grpcStop)
}

func createListener(port int, logger logr.Logger) net.Listener {
	// Create a listener at the desired port.
	var lis net.Listener
	var err error
	if len(certMountPath) > 0 {
		logger.Info("Creating TLS listener", "port", port)
		certPath := path.Join(certMountPath, certFileName)
		keyPath := path.Join(certMountPath, certKeyFileName)
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			log.Fatalf("Error certificate could not be found: %v", err)
		}
		lis, err = tls.Listen("tcp", fmt.Sprintf(":%d", port), &tls.Config{Certificates: []tls.Certificate{cert}})
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	} else {
		logger.Info("Creating non-TLS listener", "port", port)
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	}
	return lis
}
