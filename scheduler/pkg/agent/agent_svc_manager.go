/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	boff "github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	agent_pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	sched_pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/drainservice"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	k8s "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type SubServiceReadinessNotification struct {
	err            error
	subserviceName string
	subServiceType interfaces.SubServiceType
}

type AgentServiceManager struct {
	logger                   log.FieldLogger
	inferenceServerConfig    *agent_pb.ReplicaConfig
	stateManager             *LocalStateManager
	rpHTTP                   interfaces.DependencyServiceInterface
	rpGRPC                   interfaces.DependencyServiceInterface
	agentDebugService        interfaces.DependencyServiceInterface
	modelScalingService      interfaces.DependencyServiceInterface
	drainerService           interfaces.DependencyServiceInterface
	metrics                  metrics.AgentMetricsHandler
	isDraining               atomic.Bool
	criticalSubservicesReady atomic.Bool
	isStartup                atomic.Bool
	stop                     atomic.Bool
	modelScalingClientStream agent_pb.AgentService_ModelScalingTriggerClient
	agentConfig              *AgentServiceConfig
	modelTimestamps          sync.Map
	startTime                time.Time
	StorageManager
	SchedulerGrpcClientOptions
	KubernetesOptions
}

type SchedulerGrpcClientOptions struct {
	schedulerHost         string
	schedulerPlaintxtPort int
	schedulerTlsPort      int
	serverName            string
	replicaIdx            uint32
	schedulerConn         *grpc.ClientConn
	callOptions           []grpc.CallOption
	certificateStore      *seldontls.CertificateStore
}

type KubernetesOptions struct {
	secretsHandler *k8s.SecretHandler
	namespace      string
}

type StorageManager struct {
	ModelRepository repository.ModelRepository
}

type AgentServiceConfig struct {
	serverName                               string
	replicaIdx                               uint32
	schedulerHost                            string
	schedulerPlaintxtPort                    int
	schedulerTlsPort                         int
	periodReadySubService                    time.Duration
	maxElapsedTimeReadySubServiceBeforeStart time.Duration
	maxElapsedTimeReadySubServiceAfterStart  time.Duration
	maxLoadElapsedTime                       time.Duration
	maxUnloadElapsedTime                     time.Duration
	maxLoadRetryCount                        uint8
	maxUnloadRetryCount                      uint8
	unloadGraceTime                          time.Duration
}

func NewAgentServiceConfig(
	serverName string,
	replicaIdx uint32,
	schedulerHost string,
	schedulerPlaintxtPort,
	schedulerTlsPort int,
	periodReadySubService,
	maxElapsedTimeReadySubServiceBeforeStart,
	maxElapsedTimeReadySubServiceAfterStart,
	maxLoadElapsedTime,
	maxUnloadElapsedTime time.Duration,
	maxLoadRetryCount,
	maxUnloadRetryCount uint8,
	unloadGraceTime time.Duration,
) *AgentServiceConfig {
	return &AgentServiceConfig{
		serverName:                               serverName,
		replicaIdx:                               replicaIdx,
		schedulerHost:                            schedulerHost,
		schedulerPlaintxtPort:                    schedulerPlaintxtPort,
		schedulerTlsPort:                         schedulerTlsPort,
		periodReadySubService:                    periodReadySubService,
		maxElapsedTimeReadySubServiceBeforeStart: maxElapsedTimeReadySubServiceBeforeStart,
		maxElapsedTimeReadySubServiceAfterStart:  maxElapsedTimeReadySubServiceAfterStart,
		maxLoadElapsedTime:                       maxLoadElapsedTime,
		maxUnloadElapsedTime:                     maxUnloadElapsedTime,
		maxLoadRetryCount:                        maxLoadRetryCount,
		maxUnloadRetryCount:                      maxUnloadRetryCount,
		unloadGraceTime:                          unloadGraceTime,
	}
}

func ParseReplicaConfig(json string) (*agent_pb.ReplicaConfig, error) {
	config := agent_pb.ReplicaConfig{}
	err := protojson.Unmarshal([]byte(json), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func NewAgentServiceManager(
	agentConfig *AgentServiceConfig,
	logger log.FieldLogger,
	modelRepository repository.ModelRepository,
	v2Client interfaces.ModelServerControlPlaneClient,
	replicaConfig *agent_pb.ReplicaConfig,
	namespace string,
	reverseProxyHTTP interfaces.DependencyServiceInterface,
	reverseProxyGRPC interfaces.DependencyServiceInterface,
	agentDebugService interfaces.DependencyServiceInterface,
	modelScalingService interfaces.DependencyServiceInterface,
	drainerService interfaces.DependencyServiceInterface,
	metrics metrics.AgentMetricsHandler,
) *AgentServiceManager {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	modelState := NewModelState()

	stateManager := NewLocalStateManager(
		modelState, logger, v2Client, replicaConfig.GetMemoryBytes(), replicaConfig.GetOverCommitPercentage(), metrics)

	agentDebugService.SetState(stateManager)
	reverseProxyHTTP.SetState(stateManager)
	reverseProxyGRPC.SetState(stateManager)

	am := AgentServiceManager{
		logger:                logger.WithField("Name", "AgentServiceManager"),
		stateManager:          stateManager,
		inferenceServerConfig: replicaConfig,
		rpHTTP:                reverseProxyHTTP,
		rpGRPC:                reverseProxyGRPC,
		agentDebugService:     agentDebugService,
		modelScalingService:   modelScalingService,
		drainerService:        drainerService,
		metrics:               metrics,
		StorageManager: StorageManager{
			ModelRepository: modelRepository,
		},
		SchedulerGrpcClientOptions: SchedulerGrpcClientOptions{
			schedulerHost:         agentConfig.schedulerHost,
			schedulerPlaintxtPort: agentConfig.schedulerPlaintxtPort,
			schedulerTlsPort:      agentConfig.schedulerTlsPort,
			serverName:            agentConfig.serverName,
			replicaIdx:            agentConfig.replicaIdx,
			callOptions:           opts,
			certificateStore:      nil, // Needed to stop 1.48.0 lint failing
		},
		KubernetesOptions: KubernetesOptions{
			namespace: namespace,
		},
		isDraining:               atomic.Bool{},
		criticalSubservicesReady: atomic.Bool{},
		isStartup:                atomic.Bool{},
		stop:                     atomic.Bool{},
		agentConfig:              agentConfig,
		modelTimestamps:          sync.Map{},
		startTime:                time.Now(),
	}
	am.isStartup.Store(true)

	return &am
}

func (am *AgentServiceManager) Ready() bool {
	// We're never returning ready until the agent connects to the scheduler (isStartup becomes false)
	// Similarly, we're never returning ready once the draining process has started
	if am.isStartup.Load() || am.isDraining.Load() {
		return false
	}

	return am.criticalSubservicesReady.Load()
}

func (am *AgentServiceManager) StartControlLoop() error {
	logger := am.logger.WithField("func", "StartControlLoop")

	if am.schedulerConn == nil {
		err := am.createConnection()
		if err != nil {
			logger.WithError(err).Errorf("Failed to create connection to scheduler")
			return err
		}
	}

	// prom metrics
	go am.metrics.AddServerReplicaMetrics(
		am.stateManager.totalMainMemoryBytes,
		float32(am.stateManager.totalMainMemoryBytes)+am.stateManager.GetOverCommitMemoryBytes())

	// model scaling consumption
	go am.modelScalingEventsConsumer()

	// periodic subservices checker for readiness
	go am.startSubServiceChecker()

	// start wait on trigger to drain, this will also unlock any pending /terminate call before returning
	go func() {
		_ = am.drainOnRequest(am.drainerService.(*drainservice.DrainerService))
	}()

	for {
		if am.stop.Load() {
			logger.Info("Stopping")
			return nil
		}

		logFailure := func(err error, delay time.Duration) {
			am.logger.WithError(err).Errorf("Scheduler not ready")
		}
		backOffExp := util.GetClientExponentialBackoff()
		err := boff.RetryNotify(am.HandleSchedulerSubscription, backOffExp, logFailure)
		if err != nil {
			am.logger.WithError(err).Fatal("Failed to connect to the scheduler")
			return err
		}
		logger.Info("Scheduler subscription ended")
	}
}

func (am *AgentServiceManager) StopControlLoop() {
	am.stop.Store(true)
	err := am.closeSchedulerConnection()
	if err != nil {
		am.logger.WithError(err).Warn("Cannot close stream connection to scheduler")
	}
}

func (am *AgentServiceManager) closeSchedulerConnection() error {
	logger := am.logger.WithField("func", "StopSchedulerStream")
	logger.Info("Shutting down stream to scheduler")

	if am.schedulerConn != nil {
		return am.schedulerConn.Close()
	}

	return nil
}

func (am *AgentServiceManager) createConnection() error {
	conn, err := am.getConnection(am.schedulerHost, am.schedulerPlaintxtPort, am.schedulerTlsPort)
	if err != nil {
		return err
	}
	am.SchedulerGrpcClientOptions.schedulerConn = conn
	return nil
}

func (am *AgentServiceManager) isCriticalSubserviceFailure(isStartup bool, subServiceType interfaces.SubServiceType) bool {
	if isStartup {
		// On startup, any non-optional sub-service that fails will set the agent to not ready.
		// This is because, beyond critical data-plane and control-plane services, auxiliary
		// services may be needed on the initial connection to the scheduler. For example, the
		// scheduler requesting the agent to load an initial set of models requires the model
		// repository subservice.
		if subServiceType != interfaces.OptionalService {
			return true
		}
	} else {
		// During normal operation, we would like to only mark the agent as not ready if data-plane
		// services fail. We want the agent to continue to function in case of a control-plane
		// outage.
		if subServiceType == interfaces.CriticalControlPlaneService ||
			subServiceType == interfaces.CriticalDataPlaneService {
			return true
		}
	}
	return false
}

func (am *AgentServiceManager) WaitReadySubServices(isStartup bool) error {
	logger := am.logger.WithField("func", "waitReady")
	wg := &sync.WaitGroup{}

	maxElapsedTime := am.agentConfig.maxElapsedTimeReadySubServiceBeforeStart
	if !isStartup {
		maxElapsedTime = am.agentConfig.maxElapsedTimeReadySubServiceAfterStart
	}
	readyNotifications := make(chan SubServiceReadinessNotification, 2)

	// The total time we wait for all subservices to be ready is maxElapsedTime. All retries will
	// stop after this time, when the `allReadyDeadlineCtx` context timeouts.
	allReadyDeadlineCtx, allReadyDeadlineCancel := context.WithTimeout(context.Background(), maxElapsedTime)
	defer allReadyDeadlineCancel()

	waitModelRepo := func() {
		defer wg.Done()
		subserviceName := "Model Repository (rclone)"

		backoffWithContext := newAgentServiceStatusRetryBackoff(allReadyDeadlineCtx)
		err := boff.RetryNotify(am.ModelRepository.Ready, backoffWithContext,
			logSubserviceNotYetReady(logger, subserviceName))
		readyNotifications <- SubServiceReadinessNotification{
			err:            err,
			subserviceName: subserviceName,
			// For the purposes of agent readiness, we consider the ModelRepository to be a
			// control-plane service. If not ready, existing loaded models are able to
			// continue to work, but new control-plane operations (load, unload) would fail.
			subServiceType: interfaces.AuxControlPlaneService,
		}
	}

	waitInferenceServer := func() {
		defer wg.Done()
		subserviceName := "Inference Server"

		backoffWithContext := newAgentServiceStatusRetryBackoff(allReadyDeadlineCtx)
		err := boff.RetryNotify(am.stateManager.v2Client.Live, backoffWithContext,
			logSubserviceNotYetReady(logger, subserviceName))

		if isStartup {
			// Unload any existing models on server to ensure we start in a clean state
			logger.Infof("Unloading any existing models")
			err = am.UnloadAllModels()
		}

		readyNotifications <- SubServiceReadinessNotification{
			err:            err,
			subserviceName: subserviceName,
			subServiceType: interfaces.CriticalDataPlaneService,
		}
	}

	// wait for subservices from other containers in the pod (rclone, inference server)
	wg.Add(2)
	go waitModelRepo()
	go waitInferenceServer()

	// wait for internal subservices
	// readinessService not part of the list because it is started outside the AgentServiceManager
	// and needs to provide responses to readiness checks for the agent itself even if the
	// AgentServiceManager stops.
	internalSubServices := []interfaces.DependencyServiceInterface{
		am.rpHTTP,
		am.rpGRPC,
		am.agentDebugService,
		am.modelScalingService,
		am.drainerService,
	}
	wg.Add(len(internalSubServices))

	for _, subService := range internalSubServices {
		go func() {
			defer wg.Done()
			isReadyChecker(
				isStartup,
				subService,
				logger,
				readyNotifications,
				allReadyDeadlineCtx,
			)
		}()
	}

	go func() {
		wg.Wait()
		close(readyNotifications)
	}()

	criticalSubservicesNotReady := make([]string, 0)
	for notification := range readyNotifications {
		if notification.err != nil {
			logger.WithError(notification.err).Errorf("Giving up on waiting for %s service to become ready; service marked as failed", notification.subserviceName)
			if am.isCriticalSubserviceFailure(isStartup, notification.subServiceType) {
				criticalSubservicesNotReady = append(criticalSubservicesNotReady, notification.subserviceName)
			} else {
				logger.Warnf("Non-critical subservice no longer ready: %s", notification.subserviceName)
			}
		}
	}

	if len(criticalSubservicesNotReady) > 0 {
		am.criticalSubservicesReady.Store(false)
		return fmt.Errorf("the following critical subservices are no longer ready: %s", strings.Join(criticalSubservicesNotReady, ", "))
	} else {
		if isStartup {
			logger.Infof("All critical agent subservices ready.")
		}
		am.criticalSubservicesReady.Store(true)
	}
	return nil
}

func (am *AgentServiceManager) UnloadAllModels() error {
	logger := am.logger.WithField("func", "UnloadAllModels")

	models, err := am.stateManager.v2Client.GetModels()
	if err != nil {
		return err
	}

	for _, model := range models {
		if model.State == interfaces.ServerModelState_READY || model.State == interfaces.ServerModelState_LOADING {
			logger.Infof("Unloading existing model %s", model)

			v2Err := am.stateManager.v2Client.UnloadModel(model.Name)
			if v2Err != nil {
				if !v2Err.IsNotFound() {
					return v2Err.Err
				} else {
					am.logger.Warnf("Model %s not found on server", model)
				}
			}
		}

		err := am.ModelRepository.RemoveModelVersion(model.Name)
		if err != nil {
			am.logger.WithError(err).Errorf("Model %s could not be removed from repository", model)
		}
	}

	return nil
}

func (am *AgentServiceManager) getConnection(host string, plainTxtPort int, tlsPort int) (*grpc.ClientConn, error) {
	logger := am.logger.WithField("func", "getConnection")

	var err error
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixControlPlane)
	if protocol == seldontls.SecurityProtocolSSL {
		am.certificateStore, err = seldontls.NewCertificateStore(
			seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixControlPlaneServer),
		)
		if err != nil {
			return nil, err
		}
	}

	var transCreds credentials.TransportCredentials
	var port int
	if am.certificateStore == nil {
		logger.Info("Starting plaintxt client to agent server")
		transCreds = insecure.NewCredentials()
		port = plainTxtPort
	} else {
		logger.Info("Starting TLS client to agent server")
		transCreds = am.certificateStore.CreateClientTransportCredentials()
		port = tlsPort
	}

	logger.Infof("Connecting (non-blocking) to scheduler at %s:%d", host, port)

	kacp := util.GetClientKeepAliveParameters()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithKeepaliveParams(kacp),
	}
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (am *AgentServiceManager) HandleSchedulerSubscription() error {
	logger := am.logger.WithField("func", "HandleSchedulerSubscription")
	logger.Info("Call subscribe to scheduler")

	grpcClient := agent_pb.NewAgentServiceClient(am.schedulerConn)

	// Connect to the scheduler for server-side streaming
	stream, err := grpcClient.Subscribe(
		context.Background(),
		&agent_pb.AgentSubscribeRequest{
			ServerName:           am.serverName,
			ReplicaIdx:           am.replicaIdx,
			ReplicaConfig:        am.inferenceServerConfig,
			LoadedModels:         am.stateManager.modelVersions.getVersionsForAllModels(),
			Shared:               true,
			AvailableMemoryBytes: am.stateManager.GetAvailableMemoryBytesWithOverCommit(),
		},
		grpc_retry.WithMax(util.MaxGRPCRetriesOnStream),
	) // TODO make configurable
	if err != nil {
		return err
	}
	logger.Info("Subscribed to scheduler")

	// start model scaling events consumer
	clientStream, err := grpcClient.ModelScalingTrigger(context.Background())
	if err != nil {
		return err
	}

	am.modelScalingClientStream = clientStream
	defer func() {
		_, _ = clientStream.CloseAndRecv()
	}()

	// Mark startup as completed once we have an initial connection to the scheduler
	// This connection may break and will be retried, but we define the agent as "started"
	// once we are able to successfully connect once.
	am.isStartup.Store(false)
	// Start the main control loop for the agent<-scheduler stream
	for {
		if am.stop.Load() {
			logger.Info("Stopping")
			break
		}

		operation, err := stream.Recv()
		if err != nil {
			logger.WithError(err).Error("event recv failed")
			break
		}

		am.logger.Infof("Received operation")

		// Get the time since the start of the agent, this is monotonic as time.Now contains a monotonic clock
		ticksSinceStart := time.Since(am.startTime).Milliseconds()

		switch operation.Operation {
		case agent_pb.ModelOperationMessage_LOAD_MODEL:
			am.logger.Infof("calling load model")

			go func() {
				err := am.LoadModel(operation, ticksSinceStart)
				if err != nil {
					am.logger.WithError(err).Errorf(
						"Failed to handle load model %s:%d",
						operation.GetModelVersion().GetModel().GetMeta().GetName(),
						operation.GetModelVersion().GetVersion(),
					)
				}
			}()

		case agent_pb.ModelOperationMessage_UNLOAD_MODEL:
			am.logger.Infof("calling unload model")

			go func() {
				err := am.UnloadModel(operation, ticksSinceStart)
				if err != nil {
					am.logger.WithError(err).Errorf(
						"Failed to handle unload model %s:%d",
						operation.GetModelVersion().GetModel().GetMeta().GetName(),
						operation.GetModelVersion().GetVersion(),
					)
				}
			}()
		}
	}

	defer func() {
		_ = stream.CloseSend()
	}()

	logger.Info("Exiting")
	return nil
}

func (am *AgentServiceManager) getArtifactConfig(request *agent_pb.ModelOperationMessage) ([]byte, error) {
	model := request.GetModelVersion().GetModel()

	if model.GetModelSpec().StorageConfig == nil {
		return nil, nil
	}

	logger := am.logger.WithField("func", "getArtifactConfig")
	logger.Infof("Getting Rclone configuration")

	switch x := model.GetModelSpec().GetStorageConfig().GetConfig().(type) {
	case *sched_pb.StorageConfig_StorageRcloneConfig:
		return []byte(x.StorageRcloneConfig), nil
	case *sched_pb.StorageConfig_StorageSecretName:
		if am.secretsHandler == nil {
			secretClientSet, err := k8s.CreateClientset()
			if err != nil {
				return nil, err
			}

			if model.GetMeta().GetKubernetesMeta() != nil {
				am.KubernetesOptions.secretsHandler = k8s.NewSecretsHandler(
					secretClientSet,
					model.GetMeta().GetKubernetesMeta().GetNamespace(),
				)
			} else {
				return nil, fmt.Errorf(
					"can't load model %s:%dwith k8s secret %s when namespace not set",
					model.GetMeta().GetName(),
					request.GetModelVersion().GetVersion(),
					x.StorageSecretName,
				)
			}

		}

		config, err := am.secretsHandler.GetSecretConfig(x.StorageSecretName, util.K8sTimeoutDefault)
		if err != nil {
			return nil, err
		}

		return config, nil
	}

	return nil, nil
}

func (am *AgentServiceManager) LoadModel(request *agent_pb.ModelOperationMessage, timestamp int64) error {
	if request == nil || request.ModelVersion == nil {
		return fmt.Errorf("empty request received for load model")
	}

	logger := am.logger.WithField("func", "LoadModel")

	modelName := request.GetModelVersion().GetModel().GetMeta().GetName()
	modelVersion := request.GetModelVersion().GetVersion()
	modelWithVersion := util.GetVersionedModelName(modelName, modelVersion)
	pinnedModelVersion := util.GetPinnedModelVersion()

	am.stateManager.cache.Lock(modelWithVersion)
	defer am.stateManager.cache.Unlock(modelWithVersion)

	logger.Infof("Load model %s:%d", modelName, modelVersion)
	// if it is out of order message, ignore it
	ignore := ignoreIfOutOfOrder(modelWithVersion, timestamp, &am.modelTimestamps)
	if ignore {
		logger.Warnf("Ignoring out of order message for model %s:%d", modelName, modelVersion)
		return nil
	}
	defer am.modelTimestamps.Store(modelWithVersion, timestamp)

	// Get Rclone configuration
	config, err := am.getArtifactConfig(request)
	if err != nil {
		am.sendModelEventError(modelName, modelVersion, agent_pb.ModelEventMessage_LOAD_FAILED, err)
		return err
	}

	// Copy model artifact
	chosenVersionPath, err := am.ModelRepository.DownloadModelVersion(
		modelWithVersion,
		pinnedModelVersion,
		request.GetModelVersion().GetModel().GetModelSpec(),
		config,
	)
	if err != nil {
		am.sendModelEventError(modelName, modelVersion, agent_pb.ModelEventMessage_LOAD_FAILED, err)
		am.cleanup(modelWithVersion)
		return err
	}
	logger.Infof("Chose path %s for model %s:%d", *chosenVersionPath, modelName, modelVersion)

	modelConfig, err := am.ModelRepository.GetModelRuntimeInfo(modelWithVersion)
	if err != nil {
		logger.Errorf("there was a problem getting the config for model: %s", modelName)
	}

	// TODO: consider whether we need the actual protos being sent to `LoadModelVersion`?
	modifiedModelVersionRequest := getModifiedModelVersion(
		modelWithVersion,
		pinnedModelVersion,
		request.GetModelVersion(),
		modelConfig,
	)

	loaderFn := func() error {
		return am.stateManager.LoadModelVersion(modifiedModelVersionRequest)
	}

	if err := backoffWithMaxNumRetry(loaderFn, am.agentConfig.maxLoadRetryCount, am.agentConfig.maxLoadElapsedTime, logger); err != nil {
		am.sendModelEventError(modelName, modelVersion, agent_pb.ModelEventMessage_LOAD_FAILED, err)
		am.cleanup(modelWithVersion)
		return err
	}

	// if scheduler ask for autoscaling, add pointers in model scaling stats
	// we have done it via the scaling service as not to expose here all the model scaling stats
	// that we have and then call Add on each one of them
	if request.AutoscalingEnabled {
		logger.Debugf("Enabling autoscaling checks for model %s", modelWithVersion)
		if err := am.modelScalingService.(*modelscaling.StatsAnalyserService).AddModel(modelWithVersion); err != nil {
			logger.WithError(err).Warnf("Cannot add model %s to scaling service", modelWithVersion)
		}
	}

	logger.Infof("Load model %s:%d success", modelName, modelVersion)

	return am.sendAgentEvent(modelName, modelVersion, modelConfig, agent_pb.ModelEventMessage_LOADED)
}

func (am *AgentServiceManager) UnloadModel(request *agent_pb.ModelOperationMessage, timestamp int64) error {
	if request == nil || request.GetModelVersion() == nil {
		return fmt.Errorf("empty request received for unload model")
	}

	logger := am.logger.WithField("func", "UnloadModel")

	// As envoy is eventually consistent, we need to wait for a grace period before unloading the model
	// to give envoy time to drain the connections and reflect the cluster changes
	// this should be ~500ms
	time.Sleep(am.agentConfig.unloadGraceTime)

	modelName := request.GetModelVersion().GetModel().GetMeta().GetName()
	modelVersion := request.GetModelVersion().GetVersion()
	modelWithVersion := util.GetVersionedModelName(modelName, modelVersion)
	pinnedModelVersion := util.GetPinnedModelVersion()

	am.stateManager.cache.Lock(modelWithVersion)
	defer am.stateManager.cache.Unlock(modelWithVersion)

	logger.Infof("Unload model %s:%d", modelName, modelVersion)
	// if it is out of order message, ignore it
	ignore := ignoreIfOutOfOrder(modelWithVersion, timestamp, &am.modelTimestamps)
	if ignore {
		logger.Warnf("Ignoring out of order message for model %s:%d", modelName, modelVersion)
		return nil
	}
	defer am.modelTimestamps.Store(modelWithVersion, timestamp)

	// we do not care about model versions here
	// model runtime info is retrieved from the existing version, so nil is passed here
	modifiedModelVersionRequest := getModifiedModelVersion(modelWithVersion, pinnedModelVersion, request.GetModelVersion(), nil)

	unloaderFn := func() error {
		return am.stateManager.UnloadModelVersion(modifiedModelVersionRequest)
	}
	if err := backoffWithMaxNumRetry(unloaderFn, am.agentConfig.maxUnloadRetryCount, am.agentConfig.maxUnloadElapsedTime, logger); err != nil {
		am.sendModelEventError(modelName, modelVersion, agent_pb.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	// remove pointers in model scaling stats
	// we have done it via the scaling service as not to expose here all the model scaling stats that we have and then call Delete on
	// each one of them
	// note that we do not check if the model is already enabled for autoscaling, should we?
	if err := am.modelScalingService.(*modelscaling.StatsAnalyserService).DeleteModel(modelWithVersion); err != nil {
		logger.WithError(err).Warnf(
			"Cannot delete model %s from scaling service, likely that it was not enabled in the first place",
			modelWithVersion,
		)
	}

	err := am.ModelRepository.RemoveModelVersion(modelWithVersion)
	if err != nil {
		am.sendModelEventError(modelName, modelVersion, agent_pb.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	logger.Infof("Unload model %s:%d success", modelName, modelVersion)
	return am.sendAgentEvent(modelName, modelVersion, nil, agent_pb.ModelEventMessage_UNLOADED)
}

func (am *AgentServiceManager) cleanup(modelWithVersion string) {
	logger := am.logger.WithField("func", "cleanup")
	err := am.ModelRepository.RemoveModelVersion(modelWithVersion)
	if err != nil {
		logger.Errorf("could not remove model %s - %v", modelWithVersion, err)
		return
	}
	logger.Infof("removed model %s", modelWithVersion)
}

func (am *AgentServiceManager) sendModelEventError(
	modelName string,
	modelVersion uint32,
	event agent_pb.ModelEventMessage_Event,
	err error,
) {
	am.logger.WithError(err).Errorf("Failed to load model, sending error to scheduler")
	grpcClient := agent_pb.NewAgentServiceClient(am.schedulerConn)
	modelEventResponse, err := grpcClient.AgentEvent(context.Background(), &agent_pb.ModelEventMessage{
		ServerName:           am.serverName,
		ReplicaIdx:           am.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         modelVersion,
		Event:                event,
		Message:              err.Error(),
		AvailableMemoryBytes: am.stateManager.GetAvailableMemoryBytesWithOverCommit(),
	})
	if err != nil {
		am.logger.WithError(err).Errorf("Failed to send error back to scheduler on load model")
		return
	}
	am.logger.WithField("modelEventResponse", modelEventResponse).Infof("Sent agent model event to scheduler")
}

func (am *AgentServiceManager) sendAgentEvent(
	modelName string,
	modelVersion uint32,
	modelRuntimeInfo *sched_pb.ModelRuntimeInfo,
	event agent_pb.ModelEventMessage_Event,
) error {
	// if the server is draining and the model load has succeeded, we need to "cancel"
	if am.isDraining.Load() {
		if event == agent_pb.ModelEventMessage_LOADED {
			am.sendModelEventError(
				modelName,
				modelVersion,
				agent_pb.ModelEventMessage_LOAD_FAILED,
				fmt.Errorf("server replica is draining"),
			)
			return nil
		}
	}

	grpcClient := agent_pb.NewAgentServiceClient(am.schedulerConn)
	_, err := grpcClient.AgentEvent(context.Background(), &agent_pb.ModelEventMessage{
		ServerName:           am.serverName,
		ReplicaIdx:           am.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         modelVersion,
		Event:                event,
		AvailableMemoryBytes: am.stateManager.GetAvailableMemoryBytesWithOverCommit(),
		RuntimeInfo:          modelRuntimeInfo,
	})
	return err
}

func (am *AgentServiceManager) drainOnRequest(drainer *drainservice.DrainerService) error {
	drainer.WaitOnTrigger()
	am.isDraining.Store(true)

	err := am.sendAgentDrainEvent()
	if err != nil {
		am.logger.WithError(err).Warn("Could not drain agent / server")
	}

	drainer.SetSchedulerDone()
	return err
}

func (am *AgentServiceManager) sendAgentDrainEvent() error {
	grpcClient := agent_pb.NewAgentServiceClient(am.schedulerConn)
	response, err := grpcClient.AgentDrain(context.Background(), &agent_pb.AgentDrainRequest{
		ServerName: am.serverName,
		ReplicaIdx: am.replicaIdx,
	})
	if response != nil {
		am.logger.Infof("Agent drain process result %t", response.GetSuccess())
	}
	return err
}

func (am *AgentServiceManager) sendModelScalingTriggerEvent(
	modelName string,
	modelVersion uint32,
	scalingType modelscaling.ModelScalingEventType,
	amount uint32,
	data map[string]uint32,
) error {
	triggerType := agent_pb.ModelScalingTriggerMessage_SCALE_UP
	if scalingType == modelscaling.ScaleDownEvent {
		triggerType = agent_pb.ModelScalingTriggerMessage_SCALE_DOWN
	}

	err := am.modelScalingClientStream.Send(&agent_pb.ModelScalingTriggerMessage{
		ServerName:   am.serverName,
		ReplicaIdx:   am.replicaIdx,
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Trigger:      triggerType,
		Amount:       amount,
		Metrics:      data,
	})
	return err
}

func (am *AgentServiceManager) modelScalingEventsConsumer() {
	ch := am.modelScalingService.(*modelscaling.StatsAnalyserService).GetEventChannel()
	for am.modelScalingService.Ready() {
		e := <-ch
		modelName, modelVersion, err := util.GetOrignalModelNameAndVersion(e.StatsData.ModelName)
		if err != nil {
			am.logger.WithError(err).Warnf(
				"Trigger model scaling event %d for model %s failed",
				e.EventType,
				e.StatsData.ModelName,
			)
			continue
		}

		am.logger.Debugf(
			"Trigger model scaling event %d for model %s:%d with value %d",
			e.EventType,
			modelName,
			modelVersion,
			e.StatsData.Value,
		)

		err = am.sendModelScalingTriggerEvent(
			modelName, modelVersion, e.EventType, e.StatsData.Value, nil,
		)
		if err != nil {
			am.logger.WithError(err).Warnf(
				"Sending model scaling event %d for model %s failed",
				e.EventType,
				e.StatsData.ModelName,
			)
			continue
		}
	}
}

func (am *AgentServiceManager) startSubServiceChecker() {
	ticker := time.NewTicker(am.agentConfig.periodReadySubService)
	defer ticker.Stop()
	for !am.stop.Load() {
		<-ticker.C
		if err := am.WaitReadySubServices(false); err != nil {
			am.StopControlLoop()
		}
	}
}
