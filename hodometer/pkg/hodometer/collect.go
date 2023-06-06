/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hodometer

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/google/uuid"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

const (
	subscriberName = "hodometer"
)

type Collector interface {
	Collect(ctx context.Context, level MetricsLevel) *UsageMetrics
}

type kubernetesMetrics struct {
	version  string
	isGlobal bool
}

type schedulerMetrics struct {
	version string
}

type experimentMetrics struct {
	count uint
}

type pipelineMetrics struct {
	count uint
}

type serverMetrics struct {
	count             uint
	replicaCount      uint
	multimodelEnabled uint
	overcommitEnabled uint
	memoryBytes       uint
}

type modelMetrics struct {
	count uint
}

var _ Collector = (*SeldonCoreCollector)(nil)

type SeldonCoreCollector struct {
	schedulerClient  scheduler.SchedulerClient
	k8sClient        kubernetes.Interface
	logger           logrus.FieldLogger
	clusterId        string
	certificateStore *seldontls.CertificateStore
}

func NewSeldonCoreCollector(
	logger logrus.FieldLogger,
	schedulerHost string,
	schedulerPlaintxtPort uint,
	schedulerTlsPort uint,
	clusterId string,
) (*SeldonCoreCollector, error) {
	logger = logger.WithField("source", "SeldonCoreCollector")
	var certificateStore *seldontls.CertificateStore
	var err error
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixControlPlane)
	if protocol == seldontls.SecurityProtocolSSL {
		certificateStore, err = seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixControlPlaneServer))
		if err != nil {
			return nil, err
		}
	}
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	var transCreds credentials.TransportCredentials
	var port uint
	if certificateStore == nil {
		logger.Info("Starting plaintxt client to scheduler")
		transCreds = insecure.NewCredentials()
		port = schedulerPlaintxtPort
	} else {
		logger.Info("Starting TLS client to scheduler")
		transCreds = certificateStore.CreateClientTransportCredentials()
		port = schedulerTlsPort
	}
	connectOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
	}
	logger.Infof("Connecting to scheduler at %s:%d", schedulerHost, port)
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", schedulerHost, port),
		connectOptions...,
	)
	if err != nil {
		return nil, err
	}
	client := scheduler.NewSchedulerClient(conn)

	clusterId = parseOrCreateClusterId(clusterId)

	k8sClient := getKubernetesClientset(logger)

	return &SeldonCoreCollector{
		schedulerClient:  client,
		k8sClient:        k8sClient,
		logger:           logger,
		clusterId:        clusterId,
		certificateStore: certificateStore,
	}, nil
}

func parseOrCreateClusterId(clusterId string) string {
	id, err := uuid.Parse(clusterId)
	if err != nil {
		id = uuid.New()
	}
	return id.String()
}

func getKubernetesClientset(logger logrus.FieldLogger) *kubernetes.Clientset {
	logger = logger.WithField("func", "getKubernetesClientset")

	c, err := rest.InClusterConfig()
	if err != nil {
		logger.WithError(err).Warn("cannot create Kubernetes client")
		return nil
	}

	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		logger.WithError(err).Warn("cannot create Kubernetes client")
		return nil
	}

	return clientset
}

func (scc *SeldonCoreCollector) Collect(ctx context.Context, level MetricsLevel) *UsageMetrics {
	logger := scc.logger.WithField("func", "Collect")

	kubernetesResults := make(chan *kubernetesMetrics)
	schedulerResults := make(chan *schedulerMetrics)
	experimentResults := make(chan *experimentMetrics)
	pipelineResults := make(chan *pipelineMetrics)
	serverResults := make(chan *serverMetrics)
	modelResults := make(chan *modelMetrics)

	go func() {
		logger.Info("collecting Kubernetes details")
		kubernetesResults <- scc.collectKubernetes(ctx)
		close(kubernetesResults)
	}()
	go func() {
		logger.Info("collecting scheduler details")
		schedulerResults <- scc.collectScheduler(ctx)
		close(schedulerResults)
	}()
	go func() {
		logger.Info("collecting experiments")
		experimentResults <- scc.collectExperiments(ctx)
		close(experimentResults)
	}()
	go func() {
		logger.Info("collecting pipelines")
		pipelineResults <- scc.collectPipelines(ctx)
		close(pipelineResults)
	}()
	go func() {
		logger.Info("collecting servers")
		serverResults <- scc.collectServers(ctx)
		close(serverResults)
	}()
	go func() {
		logger.Info("collecting models")
		modelResults <- scc.collectModels(ctx)
		close(modelResults)
	}()

	kubernetesDetails := <-kubernetesResults
	schedulerDetails := <-schedulerResults
	experiments := <-experimentResults
	pipelines := <-pipelineResults
	servers := <-serverResults
	models := <-modelResults

	logger.Info("collating metrics")

	um := &UsageMetrics{}
	um.CollectorVersion = BuildVersion
	um.CollectorGitCommit = GitCommit
	um.ClusterId = scc.clusterId

	if kubernetesDetails != nil {
		um.KubernetesMetrics.KubernetesVersion = kubernetesDetails.version
	}
	if schedulerDetails != nil {
		um.SeldonCoreVersion = schedulerDetails.version
	}
	if level > metricsLevelCluster {
		if experiments != nil {
			um.ResourceMetrics.ExperimentCount = experiments.count
		}
		if pipelines != nil {
			um.ResourceMetrics.PipelineCount = pipelines.count
		}
		if servers != nil {
			um.ResourceMetrics.ServerCount = servers.count
			um.ResourceMetrics.ServerReplicaCount = servers.replicaCount
		}
		if models != nil {
			um.ResourceMetrics.ModelCount = models.count
		}
	}
	if level > metricsLevelResource {
		if servers != nil {
			um.FeatureMetrics.MultimodelEnabledCount = servers.multimodelEnabled
			um.FeatureMetrics.OvercommitEnabledCount = servers.overcommitEnabled
			um.FeatureMetrics.ServerMemoryGbSum = float32(servers.memoryBytes) / 1e9
		}
	}

	logger.Debugf("%#v", *um)

	return um
}

func (scc *SeldonCoreCollector) collectKubernetes(ctx context.Context) *kubernetesMetrics {
	if scc.k8sClient == nil || reflect.ValueOf(scc.k8sClient).IsNil() {
		return nil
	}

	km := &kubernetesMetrics{}
	scc.updateKubernetesVersion(km)
	return km
}

func (scc *SeldonCoreCollector) updateKubernetesVersion(metrics *kubernetesMetrics) {
	logger := scc.logger.WithField("func", "updateKubernetesVersion")

	version, err := scc.k8sClient.Discovery().ServerVersion()
	if err != nil {
		logger.WithError(err).Error("unable to retrieve server version")
		return
	}
	metrics.version = version.GitVersion
}

func (scc *SeldonCoreCollector) collectScheduler(ctx context.Context) *schedulerMetrics {
	logger := scc.logger.WithField("func", "collectScheduler")

	request := &scheduler.SchedulerStatusRequest{SubscriberName: subscriberName}
	response, err := scc.schedulerClient.SchedulerStatus(ctx, request)
	if err != nil {
		logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
		return nil
	}

	sm := &schedulerMetrics{}
	sm.version = response.ApplicationVersion
	return sm
}

func (scc *SeldonCoreCollector) collectExperiments(ctx context.Context) *experimentMetrics {
	logger := scc.logger.WithField("func", "collectExperiments")

	request := &scheduler.ExperimentStatusRequest{
		SubscriberName: subscriberName,
		Name:           nil,
	}
	subscription, err := scc.schedulerClient.ExperimentStatus(ctx, request)
	if err != nil {
		logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
		return nil
	}

	numExperiments := uint(0)
	for {
		exp, err := subscription.Recv()
		if err == io.EOF {
			return &experimentMetrics{count: numExperiments}
		}
		if err != nil {
			logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
			return nil
		}

		if exp.Active {
			numExperiments++
		}
	}
}

func (scc *SeldonCoreCollector) collectPipelines(ctx context.Context) *pipelineMetrics {
	logger := scc.logger.WithField("func", "collectPipelines")

	request := &scheduler.PipelineStatusRequest{
		SubscriberName: subscriberName,
		Name:           nil, // Request all pipelines
		AllVersions:    false,
	}

	subscription, err := scc.schedulerClient.PipelineStatus(ctx, request)
	if err != nil {
		logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
		return nil
	}

	metrics := &pipelineMetrics{}
	for {
		p, err := subscription.Recv()
		if err == io.EOF {
			return metrics
		}
		if err != nil {
			logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
			return nil
		}

		updatePipelineMetrics(metrics, p)
	}
}

func updatePipelineMetrics(metrics *pipelineMetrics, status *scheduler.PipelineStatusResponse) {
	if isPipelineActive(status) {
		metrics.count++
	}
}

func isPipelineActive(p *scheduler.PipelineStatusResponse) bool {
	if p == nil || len(p.Versions) == 0 {
		return false
	}

	isActive := false
	for _, v := range p.Versions {
		if v == nil || v.State == nil {
			continue
		}

		if v.State.Status == scheduler.PipelineVersionState_PipelineCreate ||
			v.State.Status == scheduler.PipelineVersionState_PipelineCreating ||
			v.State.Status == scheduler.PipelineVersionState_PipelineReady {
			isActive = true
		}
	}
	return isActive
}

func (scc *SeldonCoreCollector) collectServers(ctx context.Context) *serverMetrics {
	logger := scc.logger.WithField("func", "collectServers")

	request := &scheduler.ServerStatusRequest{
		SubscriberName: subscriberName,
		Name:           nil,
	}
	subscription, err := scc.schedulerClient.ServerStatus(ctx, request)
	if err != nil {
		logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
		return nil
	}

	metrics := &serverMetrics{}
	for {
		s, err := subscription.Recv()
		if err == io.EOF {
			return metrics
		}
		if err != nil {
			logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
			return nil
		}

		updateServerMetrics(metrics, s)
	}
}

func updateServerMetrics(metrics *serverMetrics, status *scheduler.ServerStatusResponse) {
	if status.ExpectedReplicas > 0 || status.AvailableReplicas > 0 {
		metrics.count++
		if status.ExpectedReplicas > 0 {
			metrics.replicaCount += uint(status.ExpectedReplicas)
		} else {
			metrics.replicaCount += uint(status.AvailableReplicas)
		}

		for _, r := range status.Resources {
			if r.OverCommitPercentage > 0 {
				// Overcommitting is redundant/useless without multi-model serving
				metrics.overcommitEnabled++
				metrics.multimodelEnabled++
			} else if r.NumLoadedModels > 1 {
				metrics.multimodelEnabled++
			}

			metrics.memoryBytes += uint(r.TotalMemoryBytes)
		}
	}
}

func (scc *SeldonCoreCollector) collectModels(ctx context.Context) *modelMetrics {
	logger := scc.logger.WithField("func", "collectModels")

	request := &scheduler.ModelStatusRequest{
		SubscriberName: subscriberName,
		Model:          nil,
	}
	subscription, err := scc.schedulerClient.ModelStatus(ctx, request)
	if err != nil {
		logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
		return nil
	}

	metrics := &modelMetrics{}
	for {
		m, err := subscription.Recv()
		if err == io.EOF {
			return metrics
		}
		if err != nil {
			logger.WithError(err).Error("unable to fetch from Seldon Core scheduler")
			return nil
		}

		if !m.Deleted {
			metrics.count++
		}
	}
}
