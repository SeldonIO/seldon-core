/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"fmt"
	"net/http"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	SSL            bool
}

type inference struct {
	ssl              bool
	host             string
	http             *http.Client
	grpc             v2_dataplane.GRPCInferenceServiceClient
	httpPort         uint
	lastHTTPResponse *http.Response
	lastGRPCResponse *v2_dataplane.ModelInferResponse
}

func NewWorld(c Config) (*World, error) {
	// TODO TLS for gRPC when c.SSL == true
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", c.IngressHost, c.GRPCPort), opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc client: %w", err)
	}
	grpcClient := v2_dataplane.NewGRPCInferenceServiceClient(conn)

	w := &World{
		namespace:      c.Namespace,
		KubeClient:     c.KubeClient,
		WatcherStorage: c.WatcherStorage,
		infer: inference{
			host:     c.IngressHost,
			http:     &http.Client{},
			httpPort: c.HTTPPort,
			grpc:     grpcClient,
			ssl:      c.SSL},
	}

	if c.Logger != nil {
		w.logger = c.Logger
	}
	return w, nil
}
