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
	"github.com/seldonio/seldon-core/tests/integration/godog/components"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	log "github.com/sirupsen/logrus"
)

type World struct {
	namespace         string
	kubeClient        *k8sclient.K8sClient
	corek8sClient     v.Interface
	watcherStorage    k8sclient.WatcherStorage
	env               *components.EnvManager //this is a combination of components for the cluster
	currentModel      *Model
	currentPipeline   *Pipeline
	currentExperiment *Experiment
	server            *server
	infer             inference
	infra             *Infrastructure
	logger            log.FieldLogger
	Label             map[string]string
}

type Config struct {
	Namespace      string
	Logger         log.FieldLogger
	KubeClient     *k8sclient.K8sClient
	K8sClient      v.Interface
	WatcherStorage k8sclient.WatcherStorage
	Env            *components.EnvManager
	GRPC           v2_dataplane.GRPCInferenceServiceClient
	IngressHost    string
	HTTPPort       uint
	GRPCPort       uint
	SSL            bool
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
		namespace:         c.Namespace,
		kubeClient:        c.KubeClient,
		watcherStorage:    c.WatcherStorage,
		env:               c.Env,
		currentModel:      newModel(label, c.Namespace, c.K8sClient, c.Logger, c.WatcherStorage),
		currentExperiment: newExperiment(label, c.Namespace, c.K8sClient, c.Logger, c.WatcherStorage),
		currentPipeline:   newPipeline(label, c.Namespace, c.K8sClient, c.Logger, c.WatcherStorage),
		server:            newServer(label, c.Namespace, c.K8sClient, c.Logger, c.KubeClient),
		infer: inference{
			host:     c.IngressHost,
			http:     &http.Client{},
			grpc:     c.GRPC,
			httpPort: c.HTTPPort,
			log:      c.Logger,
			ssl:      c.SSL},
		infra: newInfrastructure(c.Env, c.Logger),
		Label: label,
	}

	if c.Logger != nil {
		w.logger = c.Logger
	}
	return w, nil
}
