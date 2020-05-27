package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api"
	seldonclient "github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/kfserving"
	kfproto "github.com/seldonio/seldon-core/executor/api/grpc/kfserving/proto"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/grpc/tensorflow"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/api/tracing"
	"github.com/seldonio/seldon-core/executor/k8s"
	loghandler "github.com/seldonio/seldon-core/executor/logger"
	"github.com/seldonio/seldon-core/executor/proto/tensorflow/serving"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/signal"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"syscall"
	"time"
)

var (
	configPath     = flag.String("config", "", "Path to kubconfig")
	sdepName       = flag.String("sdep", "", "Seldon deployment name")
	namespace      = flag.String("namespace", "", "Namespace")
	predictorName  = flag.String("predictor", "", "Name of the predictor inside the SeldonDeployment")
	httpPort       = flag.Int("http_port", 8080, "Executor port")
	grpcPort       = flag.Int("grpc_port", 8000, "Executor port")
	wait           = flag.Duration("graceful_timeout", time.Second*15, "Graceful shutdown secs")
	protocol       = flag.String("protocol", "seldon", "The payload protocol")
	transport      = flag.String("transport", "rest", "The network transport http or grpc")
	filename       = flag.String("file", "", "Load graph from file")
	hostname       = flag.String("hostname", "localhost", "The hostname of the running server")
	logWorkers     = flag.Int("logger_workers", 5, "Number of workers handling payload logging")
	prometheusPath = flag.String("prometheus_path", "/metrics", "The prometheus metrics path")
)

func getPredictorFromEnv() (*v1.PredictorSpec, error) {
	b64Predictor := os.Getenv("ENGINE_PREDICTOR")
	if b64Predictor != "" {
		bytes, err := base64.StdEncoding.DecodeString(b64Predictor)
		if err != nil {
			return nil, err
		}
		predictor := v1.PredictorSpec{}
		if err := json.Unmarshal(bytes, &predictor); err != nil {
			return nil, err
		} else {
			return &predictor, nil
		}
	} else {
		return nil, nil
	}
}

func getPredictorFromFile(predictorName string, filename string) (*v1.PredictorSpec, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(filename, "yaml") {
		var sdep v1.SeldonDeployment
		err = yaml.Unmarshal(dat, &sdep)
		if err != nil {
			return nil, err
		}
		for _, predictor := range sdep.Spec.Predictors {
			if predictor.Name == predictorName {
				return &predictor, nil
			}
		}
		return nil, fmt.Errorf("Predictor not found %s", predictorName)
	} else {
		return nil, fmt.Errorf("Unsupported file type %s", filename)
	}
}

func getServerUrl(hostname string, port int) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("http://%s:%d/", hostname, port))
}

func runHttpServer(logger logr.Logger, predictor *v1.PredictorSpec, client seldonclient.SeldonApiClient, port int,
	probesOnly bool, serverUrl *url.URL, namespace string, protocol string, deploymentName string, prometheusPath string) {

	// Create REST API
	seldonRest := rest.NewServerRestApi(predictor, client, probesOnly, serverUrl, namespace, protocol, deploymentName, prometheusPath)
	seldonRest.Initialise()
	srv := seldonRest.CreateHttpServer(port)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error(err, "Server error")
		}
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

func runGrpcServer(logger logr.Logger, predictor *v1.PredictorSpec, client seldonclient.SeldonApiClient, port int, serverUrl *url.URL, namespace string, protocol string, deploymentName string, annotations map[string]string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer, err := grpc.CreateGrpcServer(predictor, deploymentName, annotations, logger)
	if err != nil {
		log.Fatalf("Failed to create grpc server: %v", err)
	}
	switch protocol {
	case api.ProtocolSeldon:
		seldonGrpcServer := seldon.NewGrpcSeldonServer(predictor, client, serverUrl, namespace)
		proto.RegisterSeldonServer(grpcServer, seldonGrpcServer)
	case api.ProtocolTensorflow:
		tensorflowGrpcServer := tensorflow.NewGrpcTensorflowServer(predictor, client, serverUrl, namespace)
		serving.RegisterPredictionServiceServer(grpcServer, tensorflowGrpcServer)
		serving.RegisterModelServiceServer(grpcServer, tensorflowGrpcServer)
	case api.ProtocolKfserving:
		kfservingGrpcServer := kfserving.NewGrpcKFServingServer(predictor, client, serverUrl, namespace)
		kfproto.RegisterGRPCInferenceServiceServer(grpcServer, kfservingGrpcServer)
	}
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Errorf("Grpc server error: %v", err)
	}
}

func main() {
	flag.Parse()

	if *sdepName == "" {
		log.Fatal("Seldon deployment name must be provided")
	}

	if *namespace == "" {
		log.Fatal("Namespace must be provied")
	}

	if *predictorName == "" {
		log.Fatal("Predictor must be provied")
	}

	if !(*protocol == api.ProtocolSeldon || *protocol == api.ProtocolTensorflow || *protocol == api.ProtocolKfserving) {
		log.Fatal("Protocol must be seldon, tensorflow or kfserving")
	}

	if !(*transport == "rest" || *transport == "grpc") {
		log.Fatal("Only rest and grpc supported")
	}

	serverUrl, err := getServerUrl(*hostname, *httpPort)
	if err != nil {
		log.Fatal("Failed to create server url from", *hostname, *httpPort)
	}

	logf.SetLogger(logf.ZapLogger(false))
	logger := logf.Log.WithName("entrypoint")

	logger.Info("Flags", "transport", *transport)

	var predictor *v1.PredictorSpec
	if *filename != "" {
		logger.Info("Trying to get predictor from file")
		predictor, err = getPredictorFromFile(*predictorName, *filename)
		if err != nil {
			logger.Error(err, "Failed to get predictor from file")
			panic(err)
		}
	} else {
		logger.Info("Trying to get predictor from Env")
		predictor, err = getPredictorFromEnv()
		if err != nil {
			logger.Error(err, "Failed to get predictor from Env")
			panic(err)
		} else if predictor == nil {
			logger.Info("Trying to get predictor from API")
			seldonDeploymentClient := seldonclient.NewSeldonDeploymentClient(configPath)
			predictor, err = seldonDeploymentClient.GetPredictor(*sdepName, *namespace, *predictorName)
			if err != nil {
				logger.Error(err, "Failed to find predictor", "name", predictor)
				panic(err)
			}
		}
	}

	// Ensure standard OpenAPI seldon API file has this deployment's values
	err = rest.EmbedSeldonDeploymentValuesInSwaggerFile(*namespace, *sdepName)
	if err != nil {
		log.Error(err, "Failed to embed variables on OpenAPI template")
	}

	annotations, err := k8s.GetAnnotations()
	if err != nil {
		log.Error(err, "Failed to load annotations")
	}

	//Start Logger Dispacther
	loghandler.StartDispatcher(*logWorkers, logger, *sdepName, *namespace, *predictorName)

	//Init Tracing
	closer, err := tracing.InitTracing()
	if err != nil {
		log.Fatal("Could not initialize jaeger tracer", err.Error())
	}
	defer closer.Close()

	if *transport == "rest" {
		clientRest, err := rest.NewJSONRestClient(*protocol, *sdepName, predictor, annotations)
		if err != nil {
			log.Fatalf("Failed to create http client: %v", err)
		}
		logger.Info("Running http server ", "port", *httpPort)
		runHttpServer(logger, predictor, clientRest, *httpPort, false, serverUrl, *namespace, *protocol, *sdepName, *prometheusPath)
	} else {
		logger.Info("Running http probes only server ", "port", *httpPort)
		go runHttpServer(logger, predictor, nil, *httpPort, true, serverUrl, *namespace, *protocol, *sdepName, *prometheusPath)
		logger.Info("Running grpc server ", "port", *grpcPort)
		var clientGrpc seldonclient.SeldonApiClient
		switch *protocol {
		case api.ProtocolSeldon:
			clientGrpc = seldon.NewSeldonGrpcClient(predictor, *sdepName, annotations)
		case api.ProtocolTensorflow:
			clientGrpc = tensorflow.NewTensorflowGrpcClient(predictor, *sdepName, annotations)
		case api.ProtocolKfserving:
			clientGrpc = kfserving.NewKFServingGrpcClient(predictor, *sdepName, annotations)
		default:
			log.Fatalf("Failed to create grpc client. Unknown protocol %s: %v", *protocol, err)
		}
		runGrpcServer(logger, predictor, clientGrpc, *grpcPort, serverUrl, *namespace, *protocol, *sdepName, annotations)

	}

}
