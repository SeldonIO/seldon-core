package main

import (
	"flag"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/kafka"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/k8s"
	predictor2 "github.com/seldonio/seldon-core/executor/predictor"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	configPath    = flag.String("config", "", "Path to kubconfig")
	modelName     = flag.String("model_name", "", "Name of the model inside the predictor")
	predictorName = flag.String("predictor", "", "Name of the predictor inside the SeldonDeployment")
	sdepName      = flag.String("sdep", "", "Seldon deployment name")
	namespace     = flag.String("namespace", "default", "Namespace")
	hostname      = flag.String("hostname", "localhost", "The hostname of client service")
	httpPort      = flag.Int("http_port", 9000, "Port of the client service")
	protocol      = flag.String("protocol", "seldon", "The payload protocol")
	filename      = flag.String("file", "", "Load graph from file")
	broker        = flag.String("broker", "", "The kafka broker as host:port")
)

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

	if *hostname == "" {
		log.Fatalf("Required argument hostname missing")
	}

	if !(*protocol == api.ProtocolSeldon || *protocol == api.ProtocolTensorflow) {
		log.Fatal("Invalid protocol: must be seldon or tensorflow")
	}

	predictor, err := predictor2.GetPredictor(*predictorName, *filename, *sdepName, *namespace, configPath)
	if err != nil {
		log.Error(err, "Failed to get predictor")
		os.Exit(-1)
	}

	annotations, err := k8s.GetAnnotations()
	if err != nil {
		log.Error(err, "Failed to load annotations")
	}

	client, err := rest.NewJSONRestClient(*protocol, *sdepName, predictor, annotations)

	logf.SetLogger(logf.ZapLogger(false))
	logger := logf.Log.WithName("entrypoint")

	kafkaProxy := kafka.NewKafkaProxy(client, *modelName, *predictorName, *sdepName, *namespace, *broker, *hostname, int32(*httpPort), logger)

	err = kafkaProxy.Consume()
	if err != nil {
		logger.Error(err, "consume failed - ending")
	} else {
		logger.Info("Stopping")
	}

}
