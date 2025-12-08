package steps

import (
	"net/http"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/godog/k8sclient"
	"github.com/sirupsen/logrus"
)

type World struct {
	namespace            string
	KubeClient           *k8sclient.K8sClient
	WatcherStorage       k8sclient.WatcherStorage
	StartingClusterState string //todo: this will be a combination of starting state awareness of core 2 such as the
	//todo:  server config,seldon config and seldon runtime to be able to reconcile to starting state should we change
	//todo: the state such as reducing replicas to 0 of scheduler to test unavailability
	CurrentModel *Model
	infer        inference
	logger       *logrus.Entry
}

type Config struct {
	Namespace      string
	Logger         *logrus.Entry
	KubeClient     *k8sclient.K8sClient
	WatcherStorage k8sclient.WatcherStorage
	IngressHost    string
	HTTPPort       uint
	GRPCPort       uint
}

type inference struct {
	host             string
	http             *http.Client
	grpc             v2_dataplane.GRPCInferenceServiceClient
	httpPort         uint
	grpcPort         uint
	lastHTTPResponse *http.Response
	lastGRPCResponse *v2_dataplane.ModelInferResponse
}

func NewWorld(c Config) *World {
	w := &World{
		namespace:      c.Namespace,
		KubeClient:     c.KubeClient,
		WatcherStorage: c.WatcherStorage,
		infer:          inference{host: c.IngressHost, http: &http.Client{}, httpPort: c.HTTPPort, grpcPort: c.GRPCPort},
	}
	if c.Logger != nil {
		w.logger = c.Logger
	}
	return w
}
