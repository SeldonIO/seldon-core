package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api"
	seldonclient "github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
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
	// OpenAPI values
	openapiFilePath     = "./openapi/seldon.json"
	openapiPredPath     = "/seldon/{namespace}/{deployment}/api/v1.0/predictions"
	openapiFeedbackPath = "/seldon/{namespace}/{deployment}/api/v1.0/feedback"
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

func embedSeldonDeploymentValuesToSwaggerFile(namespace string, sdepName string) error {
	openapiInputBytes, err := ioutil.ReadFile(openapiFilePath)
	if err != nil {
		return err
	}
	var openapiInterface interface{}
	if err := json.Unmarshal(openapiInputBytes, &openapiInterface); err != nil {
		return err
	}

	jsonFormatError := errors.New("Incorrect format for OpenAPI JSON file")

	replacer := strings.NewReplacer(
		"{namespace}", namespace,
		"{deployment}", sdepName)

	// Ensure json is a map before performing actions
	if openapiJson, ok := openapiInterface.(map[string]interface{}); ok {
		// Remove the servers element to ensure it uses the same URL
		delete(openapiJson, "servers")

		// Get the paths key value to remove the parameters from each of the URLs
		if pathsJson, ok := openapiJson["paths"].(map[string]interface{}); ok {
			// Delete the parameters field from the predictions path
			if openapiPredPathJson, ok := pathsJson[openapiPredPath].(map[string]interface{}); ok {
				if openapiPredPathPostJson, ok := openapiPredPathJson["post"].(map[string]interface{}); ok {
					delete(openapiPredPathPostJson, "parameters")
				} else {
					return jsonFormatError
				}
			} else {
				return jsonFormatError
			}

			// Rename the predictions path to use the namespace and deploymentName instead of placeholder values
			openapiPredPathReplaced := replacer.Replace(openapiPredPath)
			pathsJson[openapiPredPathReplaced] = pathsJson[openapiPredPath]
			delete(pathsJson, openapiPredPath)

			// Delete the parameters field from the feedback path
			if openapiFeedbackPathJson, ok := pathsJson[openapiFeedbackPath].(map[string]interface{}); ok {
				if openapiFeedbackPathPostJson, ok := openapiFeedbackPathJson["post"].(map[string]interface{}); ok {
					delete(openapiFeedbackPathPostJson, "parameters")
				} else {
					return jsonFormatError
				}
			} else {
				return jsonFormatError
			}

			// Rename the predictions path to use the namespace and deploymentName instead of placeholder values
			openapiFeedbackPathReplaced := replacer.Replace(openapiFeedbackPath)
			pathsJson[openapiFeedbackPathReplaced] = pathsJson[openapiFeedbackPath]
			delete(pathsJson, openapiFeedbackPath)
		} else {
			return jsonFormatError
		}

	} else {
		return jsonFormatError
	}

	// We use MarshalIndent so that the output can be humanly visible and indented
	openapiOutputBytes, err := json.MarshalIndent(openapiInterface, "", "    ")

	if err := ioutil.WriteFile(openapiFilePath, openapiOutputBytes, 0644); err != nil {
		return err
	}

	return nil
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

	if !(*protocol == api.ProtocolSeldon || *protocol == api.ProtocolTensorflow) {
		log.Fatal("Protocol must be seldon or tensorflow")
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
	err = embedSeldonDeploymentValuesToSwaggerFile(*namespace, *sdepName)
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
		if *protocol == "seldon" {
			clientGrpc = seldon.NewSeldonGrpcClient(predictor, *sdepName, annotations)
		} else {
			clientGrpc = tensorflow.NewTensorflowGrpcClient(predictor, *sdepName, annotations)
		}
		runGrpcServer(logger, predictor, clientGrpc, *grpcPort, serverUrl, *namespace, *protocol, *sdepName, annotations)

	}

}
