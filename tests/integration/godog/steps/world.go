/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"net/http"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	v "github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	log "github.com/sirupsen/logrus"
)

type World struct {
	namespace            string
	kubeClient           *k8sclient.K8sClient
	corek8sClient        v.Interface
	watcherStorage       k8sclient.WatcherStorage
	StartingClusterState string //todo: this will be a combination of starting state awareness of core 2 such as the
	//todo:  server config,seldon config and seldon runtime to be able to reconcile to starting state should we change
	//todo: the state such as reducing replicas to 0 of scheduler to test unavailability
	currentModel *Model
	infer        inference
	logger       *log.Logger
	Label        map[string]string
}

type Config struct {
	Namespace      string
	Logger         *log.Logger
	KubeClient     *k8sclient.K8sClient
	K8sClient      v.Interface
	WatcherStorage k8sclient.WatcherStorage
	GRPC           v2_dataplane.GRPCInferenceServiceClient
	IngressHost    string
	HTTPPort       uint
	GRPCPort       uint
	SSL            bool
}

type inference struct {
	ssl              bool
	host             string
	http             *http.Client
	grpc             v2_dataplane.GRPCInferenceServiceClient
	httpPort         uint
	lastHTTPResponse *http.Response
	lastGRPCResponse lastGRPCResponse
}

type lastGRPCResponse struct {
	response *v2_dataplane.ModelInferResponse
	err      error
}

func NewWorld(c Config) (*World, error) {
	label := map[string]string{
		"scenario": randomString(6),
	}

	w := &World{
		namespace:      c.Namespace,
		kubeClient:     c.KubeClient,
		watcherStorage: c.WatcherStorage,
		currentModel:   NewModel(label, c.Namespace, c.K8sClient, c.Logger),
		infer: inference{
			host:     c.IngressHost,
			http:     &http.Client{},
			grpc:     c.GRPC,
			httpPort: c.HTTPPort,
			ssl:      c.SSL},
		Label: label,
	}

	if c.Logger != nil {
		w.logger = c.Logger
	}
	return w, nil
}
