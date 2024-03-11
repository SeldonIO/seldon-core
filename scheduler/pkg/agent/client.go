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
	"sync/atomic"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/drainservice"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	k8s "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type Client struct {
	logger                   log.FieldLogger
	configChan               chan config.AgentConfiguration
	replicaConfig            *agent.ReplicaConfig
	stateManager             *LocalStateManager
	rpHTTP                   interfaces.DependencyServiceInterface
	rpGRPC                   interfaces.DependencyServiceInterface
	agentDebugService        interfaces.DependencyServiceInterface
	modelScalingService      interfaces.DependencyServiceInterface
	drainerService           interfaces.DependencyServiceInterface
	metrics                  metrics.AgentMetricsHandler
	isDraining               atomic.Bool
	stop                     atomic.Bool
	modelScalingClientStream agent.AgentService_ModelScalingTriggerClient
	settings                 *ClientSettings
	ClientServices
	SchedulerGrpcClientOptions
	KubernetesOptions
}

type SchedulerGrpcClientOptions struct {
	schedulerHost         string
	schedulerPlaintxtPort int
	schedulerTlsPort      int
	serverName            string
	replicaIdx            uint32
	conn                  *grpc.ClientConn
	callOptions           []grpc.CallOption
	certificateStore      *seldontls.CertificateStore
}

type KubernetesOptions struct {
	secretsHandler *k8s.SecretHandler
	namespace      string
}

type ClientServices struct {
	ModelRepository repository.ModelRepository
}

type ClientSettings struct {
	serverName                               string
	replicaIdx                               uint32
	schedulerHost                            string
	schedulerPlaintxtPort                    int
	schedulerTlsPort                         int
	periodReadySubService                    time.Duration
	maxElapsedTimeReadySubServiceBeforeStart time.Duration
	maxElapsedTimeReadySubServiceAfterStart  time.Duration
	maxLoadRetryCount                        uint8
	maxUnloadRetryCount                      uint8
}

func NewClientSettings(
	serverName string,
	replicaIdx uint32,
	schedulerHost string,
	schedulerPlaintxtPort,
	schedulerTlsPort int,
	periodReadySubService,
	maxElapsedTimeReadySubServiceBeforeStart,
	maxElapsedTimeReadySubServiceAfterStart time.Duration,
	maxLoadRetryCount,
	maxUnloadRetryCount uint8,
) *ClientSettings {
	return &ClientSettings{
		serverName:                               serverName,
		replicaIdx:                               replicaIdx,
		schedulerHost:                            schedulerHost,
		schedulerPlaintxtPort:                    schedulerPlaintxtPort,
		schedulerTlsPort:                         schedulerTlsPort,
		periodReadySubService:                    periodReadySubService,
		maxElapsedTimeReadySubServiceBeforeStart: maxElapsedTimeReadySubServiceBeforeStart,
		maxElapsedTimeReadySubServiceAfterStart:  maxElapsedTimeReadySubServiceAfterStart,
		maxLoadRetryCount:                        maxLoadRetryCount,
		maxUnloadRetryCount:                      maxUnloadRetryCount,
	}
}

func ParseReplicaConfig(json string) (*agent.ReplicaConfig, error) {
	config := agent.ReplicaConfig{}
	err := protojson.Unmarshal([]byte(json), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func NewClient(
	settings *ClientSettings,
	logger log.FieldLogger,
	modelRepository repository.ModelRepository,
	v2Client interfaces.ModelServerControlPlaneClient,
	replicaConfig *agent.ReplicaConfig,
	namespace string,
	reverseProxyHTTP interfaces.DependencyServiceInterface,
	reverseProxyGRPC interfaces.DependencyServiceInterface,
	agentDebugService interfaces.DependencyServiceInterface,
	modelScalingService interfaces.DependencyServiceInterface,
	drainerService interfaces.DependencyServiceInterface,
	metrics metrics.AgentMetricsHandler,
) *Client {

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

	return &Client{
		logger:              logger.WithField("Name", "Client"),
		configChan:          make(chan config.AgentConfiguration),
		stateManager:        stateManager,
		replicaConfig:       replicaConfig,
		rpHTTP:              reverseProxyHTTP,
		rpGRPC:              reverseProxyGRPC,
		agentDebugService:   agentDebugService,
		modelScalingService: modelScalingService,
		drainerService:      drainerService,
		metrics:             metrics,
		ClientServices: ClientServices{
			ModelRepository: modelRepository,
		},
		SchedulerGrpcClientOptions: SchedulerGrpcClientOptions{
			schedulerHost:         settings.schedulerHost,
			schedulerPlaintxtPort: settings.schedulerPlaintxtPort,
			schedulerTlsPort:      settings.schedulerTlsPort,
			serverName:            settings.serverName,
			replicaIdx:            settings.replicaIdx,
			callOptions:           opts,
			certificateStore:      nil, // Needed to stop 1.48.0 lint failing
		},
		KubernetesOptions: KubernetesOptions{
			namespace: namespace,
		},
		isDraining: atomic.Bool{},
		stop:       atomic.Bool{},
		settings:   settings,
	}
}

func (c *Client) Start() error {
	logger := c.logger.WithField("func", "Start")

	if c.conn == nil {
		err := c.createConnection()
		if err != nil {
			logger.WithError(err).Errorf("Failed to create connection to scheduler")
			return err
		}
	}

	// prom metrics
	go c.metrics.AddServerReplicaMetrics(
		c.stateManager.totalMainMemoryBytes,
		float32(c.stateManager.totalMainMemoryBytes)+c.stateManager.GetOverCommitMemoryBytes())

	// model scaling consumption
	go c.modelScalingEventsConsumer()

	// periodic subservices checker for readiness
	go c.startSubServiceChecker()

	// start wait on trigger to drain, this will also unlock any pending /terminate call before returning
	go func() {
		_ = c.drainOnRequest(c.drainerService.(*drainservice.DrainerService))
	}()

	for {
		if c.stop.Load() {
			logger.Info("Stopping")
			return nil
		}

		logFailure := func(err error, delay time.Duration) {
			c.logger.WithError(err).Errorf("Scheduler not ready")
		}
		backOffExp := backoff.NewExponentialBackOff()
		backOffExp.MaxElapsedTime = 0 // Never stop due to large time between calls
		err := backoff.RetryNotify(c.StartService, backOffExp, logFailure)
		if err != nil {
			c.logger.WithError(err).Fatal("Failed to start client")
			return err
		}
		logger.Info("Subscribe ended")
	}
}

func (c *Client) Stop() {
	c.stop.Store(true)
	err := c.closeSchedulerConnection()
	if err != nil {
		c.logger.WithError(err).Warn("Cannot close stream connection to scheduler")
	}
}

func (c *Client) closeSchedulerConnection() error {
	logger := c.logger.WithField("func", "StopSchedulerStream")
	logger.Info("Shutting down stream to scheduler")

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *Client) createConnection() error {
	conn, err := c.getConnection(c.schedulerHost, c.schedulerPlaintxtPort, c.schedulerTlsPort)
	if err != nil {
		return err
	}
	c.SchedulerGrpcClientOptions.conn = conn
	return nil
}

func (c *Client) WaitReadySubServices(isStartup bool) error {
	logger := c.logger.WithField("func", "waitReady")

	maxElapsedTime := c.settings.maxElapsedTimeReadySubServiceBeforeStart
	if !isStartup {
		maxElapsedTime = c.settings.maxElapsedTimeReadySubServiceAfterStart
	}
	backoffWithMax := backoff.NewExponentialBackOff()
	backoffWithMax.MaxElapsedTime = maxElapsedTime

	// Wait for model repo to be ready
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Rclone not ready")
	}

	//TODO make retry configurable
	err := backoff.RetryNotify(c.ModelRepository.Ready, backoffWithMax, logFailure)
	if err != nil {
		logger.WithError(err).Error("Failed to wait for model repository to be ready")
		return err
	}

	// Wait for Inference server to be ready
	logFailure = func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Inference server not ready")
	}

	err = backoff.RetryNotify(c.stateManager.v2Client.Live, backoffWithMax, logFailure)
	if err != nil {
		logger.WithError(err).Error("Failed to wait for inference server to be ready")
		return err
	}

	if isStartup {
		// Unload any existing models on server to ensure we start in a clean state
		logger.Infof("Unloading any existing models")
		err = c.UnloadAllModels()
		if err != nil {
			return err
		}
	}

	// http reverse proxy
	if err := isReadyChecker(
		isStartup, c.rpHTTP, logger, "Rest proxy not ready",
		maxElapsedTime,
	); err != nil {
		return err
	}

	// grpc reverse proxy
	if err := isReadyChecker(isStartup, c.rpGRPC, logger, "Grpc proxy not ready",
		maxElapsedTime,
	); err != nil {
		return err
	}

	// agent debug service
	if err := isReadyChecker(isStartup, c.agentDebugService, logger, "Agent debug service not ready",
		maxElapsedTime,
	); err != nil {
		return err
	}

	// model scaling service
	if err := isReadyChecker(isStartup, c.modelScalingService, logger, "Scaling service not ready",
		maxElapsedTime,
	); err != nil {
		return err
	}

	// drainer service
	if err := isReadyChecker(isStartup, c.drainerService, logger, "Inference server drainer service not ready",
		maxElapsedTime,
	); err != nil {
		return err
	}

	return nil
}

func (c *Client) UnloadAllModels() error {
	logger := c.logger.WithField("func", "UnloadAllModels")

	models, err := c.stateManager.v2Client.GetModels()
	if err != nil {
		return err
	}

	for _, model := range models {
		if model.State == interfaces.ServerModelState_READY || model.State == interfaces.ServerModelState_LOADING {
			logger.Infof("Unloading existing model %s", model)

			v2Err := c.stateManager.v2Client.UnloadModel(model.Name)
			if v2Err != nil {
				if !v2Err.IsNotFound() {
					return v2Err.Err
				} else {
					c.logger.Warnf("Model %s not found on server", model)
				}
			}
		}

		err := c.ModelRepository.RemoveModelVersion(model.Name)
		if err != nil {
			c.logger.WithError(err).Errorf("Model %s could not be removed from repository", model)
		}
	}

	return nil
}

func (c *Client) getConnection(host string, plainTxtPort int, tlsPort int) (*grpc.ClientConn, error) {
	logger := c.logger.WithField("func", "getConnection")

	var err error
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixControlPlane)
	if protocol == seldontls.SecurityProtocolSSL {
		c.certificateStore, err = seldontls.NewCertificateStore(
			seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixControlPlaneServer),
		)
		if err != nil {
			return nil, err
		}
	}

	var transCreds credentials.TransportCredentials
	var port int
	if c.certificateStore == nil {
		logger.Info("Starting plaintxt client to agent server")
		transCreds = insecure.NewCredentials()
		port = plainTxtPort
	} else {
		logger.Info("Starting TLS client to agent server")
		transCreds = c.certificateStore.CreateClientTransportCredentials()
		port = tlsPort
	}

	logger.Infof("Connecting (non-blocking) to scheduler at %s:%d", host, port)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) StartService() error {
	logger := c.logger.WithField("func", "StartService")
	logger.Info("Call subscribe to scheduler")

	grpcClient := agent.NewAgentServiceClient(c.conn)

	// Connect to the scheduler for server-side streaming
	stream, err := grpcClient.Subscribe(
		context.Background(),
		&agent.AgentSubscribeRequest{
			ServerName:           c.serverName,
			ReplicaIdx:           c.replicaIdx,
			ReplicaConfig:        c.replicaConfig,
			LoadedModels:         c.stateManager.modelVersions.getVersionsForAllModels(),
			Shared:               true,
			AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytesWithOverCommit(),
		},
		grpc_retry.WithMax(1),
	) //TODO make configurable
	if err != nil {
		return err
	}

	logger.Info("Subscribed to scheduler")

	// start model scaling events consumer
	clientStream, err := grpcClient.ModelScalingTrigger(context.Background())
	if err != nil {
		return err
	}

	c.modelScalingClientStream = clientStream
	defer func() {
		_, _ = clientStream.CloseAndRecv()
	}()

	// Start the main control loop for the agent<-scheduler stream
	for {
		if c.stop.Load() {
			logger.Info("Stopping")
			break
		}

		operation, err := stream.Recv()
		if err != nil {
			logger.WithError(err).Error("event recv failed")
			break
		}

		c.logger.Infof("Received operation")

		switch operation.Operation {
		case agent.ModelOperationMessage_LOAD_MODEL:
			c.logger.Infof("calling load model")

			go func() {
				err := c.LoadModel(operation)
				if err != nil {
					c.logger.WithError(err).Errorf(
						"Failed to handle load model %s:%d",
						operation.GetModelVersion().GetModel().GetMeta().GetName(),
						operation.GetModelVersion().GetVersion(),
					)
				}
			}()

		case agent.ModelOperationMessage_UNLOAD_MODEL:
			c.logger.Infof("calling unload model")

			go func() {
				err := c.UnloadModel(operation)
				if err != nil {
					c.logger.WithError(err).Errorf(
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

func (c *Client) getArtifactConfig(request *agent.ModelOperationMessage) ([]byte, error) {
	model := request.GetModelVersion().GetModel()

	if model.GetModelSpec().StorageConfig == nil {
		return nil, nil
	}

	logger := c.logger.WithField("func", "getArtifactConfig")
	logger.Infof("Getting Rclone configuration")

	switch x := model.GetModelSpec().GetStorageConfig().GetConfig().(type) {
	case *pbs.StorageConfig_StorageRcloneConfig:
		return []byte(x.StorageRcloneConfig), nil
	case *pbs.StorageConfig_StorageSecretName:
		if c.secretsHandler == nil {
			secretClientSet, err := k8s.CreateClientset()
			if err != nil {
				return nil, err
			}

			if model.GetMeta().GetKubernetesMeta() != nil {
				c.KubernetesOptions.secretsHandler = k8s.NewSecretsHandler(
					secretClientSet,
					model.GetMeta().GetKubernetesMeta().GetNamespace(),
				)
			} else {
				return nil, fmt.Errorf(
					"Can't load model %s:%dwith k8s secret %s when namespace not set",
					model.GetMeta().GetName(),
					request.GetModelVersion().GetVersion(),
					x.StorageSecretName,
				)
			}

		}

		config, err := c.secretsHandler.GetSecretConfig(x.StorageSecretName)
		if err != nil {
			return nil, err
		}

		return config, nil
	}

	return nil, nil
}

func (c *Client) LoadModel(request *agent.ModelOperationMessage) error {
	if request == nil || request.ModelVersion == nil {
		return fmt.Errorf("Empty request received for load model")
	}

	logger := c.logger.WithField("func", "LoadModel")

	modelName := request.GetModelVersion().GetModel().GetMeta().GetName()
	modelVersion := request.GetModelVersion().GetVersion()
	modelWithVersion := util.GetVersionedModelName(modelName, modelVersion)
	pinnedModelVersion := util.GetPinnedModelVersion()

	c.stateManager.cache.Lock(modelWithVersion)
	defer c.stateManager.cache.Unlock(modelWithVersion)

	logger.Infof("Load model %s:%d", modelName, modelVersion)

	// Get Rclone configuration
	config, err := c.getArtifactConfig(request)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}

	// Copy model artifact
	chosenVersionPath, err := c.ModelRepository.DownloadModelVersion(
		modelWithVersion,
		pinnedModelVersion,
		request.GetModelVersion().GetModel().GetModelSpec(),
		config,
	)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	logger.Infof("Chose path %s for model %s:%d", *chosenVersionPath, modelName, modelVersion)

	// TODO: consider whether we need the actual protos being sent to `LoadModelVersion`?
	modifiedModelVersionRequest := getModifiedModelVersion(
		modelWithVersion,
		pinnedModelVersion,
		request.GetModelVersion(),
	)
	loaderFn := func() error {
		return c.stateManager.LoadModelVersion(modifiedModelVersionRequest)
	}
	if err := backoffWithMaxNumRetry(loaderFn, c.settings.maxLoadRetryCount, logger); err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}

	// if scheduler ask for autoscaling, add pointers in model scaling stats
	// we have done it via the scaling service as not to expose here all the model scaling stats
	// that we have and then call Add on each one of them
	if request.AutoscalingEnabled {
		logger.Debugf("Enabling autoscaling checks for model %s", modelWithVersion)
		if err := c.modelScalingService.(*modelscaling.StatsAnalyserService).AddModel(modelWithVersion); err != nil {
			logger.WithError(err).Warnf("Cannot add model %s to scaling service", modelWithVersion)
		}
	}

	logger.Infof("Load model %s:%d success", modelName, modelVersion)
	return c.sendAgentEvent(modelName, modelVersion, agent.ModelEventMessage_LOADED)
}

func (c *Client) UnloadModel(request *agent.ModelOperationMessage) error {
	if request == nil || request.GetModelVersion() == nil {
		return fmt.Errorf("Empty request received for unload model")
	}

	logger := c.logger.WithField("func", "UnloadModel")

	modelName := request.GetModelVersion().GetModel().GetMeta().GetName()
	modelVersion := request.GetModelVersion().GetVersion()
	modelWithVersion := util.GetVersionedModelName(modelName, modelVersion)
	pinnedModelVersion := util.GetPinnedModelVersion()

	c.stateManager.cache.Lock(modelWithVersion)
	defer c.stateManager.cache.Unlock(modelWithVersion)

	logger.Infof("Unload model %s:%d", modelName, modelVersion)

	// we do not care about model versions here
	modifiedModelVersionRequest := getModifiedModelVersion(modelWithVersion, pinnedModelVersion, request.GetModelVersion())

	unloaderFn := func() error {
		return c.stateManager.UnloadModelVersion(modifiedModelVersionRequest)
	}
	if err := backoffWithMaxNumRetry(unloaderFn, c.settings.maxUnloadRetryCount, logger); err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	// remove pointers in model scaling stats
	// we have done it via the scaling service as not to expose here all the model scaling stats that we have and then call Delete on
	// each one of them
	// note that we do not check if the model is already enabled for autoscaling, should we?
	if err := c.modelScalingService.(*modelscaling.StatsAnalyserService).DeleteModel(modelWithVersion); err != nil {
		logger.WithError(err).Warnf(
			"Cannot delete model %s from scaling service, likely that it was not enabled in the first place",
			modelWithVersion,
		)
	}

	err := c.ModelRepository.RemoveModelVersion(modelWithVersion)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	logger.Infof("Unload model %s:%d success", modelName, modelVersion)
	return c.sendAgentEvent(modelName, modelVersion, agent.ModelEventMessage_UNLOADED)
}

func (c *Client) sendModelEventError(
	modelName string,
	modelVersion uint32,
	event agent.ModelEventMessage_Event,
	err error,
) {
	c.logger.WithError(err).Errorf("Failed to load model, sending error to scheduler")
	grpcClient := agent.NewAgentServiceClient(c.conn)
	modelEventResponse, err := grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         modelVersion,
		Event:                event,
		Message:              err.Error(),
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytesWithOverCommit(),
	})
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to send error back to scheduler on load model")
		return
	}
	c.logger.WithField("modelEventResponse", modelEventResponse).Infof("Sent agent model event to scheduler")
}

func (c *Client) sendAgentEvent(
	modelName string,
	modelVersion uint32,
	event agent.ModelEventMessage_Event,
) error {
	// if the server is draining and the model load has succeeded, we need to "cancel"
	if c.isDraining.Load() {
		if event == agent.ModelEventMessage_LOADED {
			c.sendModelEventError(
				modelName,
				modelVersion,
				agent.ModelEventMessage_LOAD_FAILED,
				fmt.Errorf("server replica is draining"),
			)
			return nil
		}
	}

	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err := grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         modelVersion,
		Event:                event,
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytesWithOverCommit(),
	})
	return err
}

func (c *Client) drainOnRequest(drainer *drainservice.DrainerService) error {
	drainer.WaitOnTrigger()
	c.isDraining.Store(true)

	err := c.sendAgentDrainEvent()
	if err != nil {
		c.logger.WithError(err).Warn("Could not drain agent / server")
	}

	drainer.SetSchedulerDone()
	return err
}

func (c *Client) sendAgentDrainEvent() error {
	grpcClient := agent.NewAgentServiceClient(c.conn)
	response, err := grpcClient.AgentDrain(context.Background(), &agent.AgentDrainRequest{
		ServerName: c.serverName,
		ReplicaIdx: c.replicaIdx,
	})
	if response != nil {
		c.logger.Infof("Agent drain process result %t", response.GetSuccess())
	}
	return err
}

func (c *Client) sendModelScalingTriggerEvent(
	modelName string,
	modelVersion uint32,
	scalingType modelscaling.ModelScalingEventType,
	amount uint32,
	data map[string]uint32,
) error {
	triggerType := agent.ModelScalingTriggerMessage_SCALE_UP
	if scalingType == modelscaling.ScaleDownEvent {
		triggerType = agent.ModelScalingTriggerMessage_SCALE_DOWN
	}

	err := c.modelScalingClientStream.Send(&agent.ModelScalingTriggerMessage{
		ServerName:   c.serverName,
		ReplicaIdx:   c.replicaIdx,
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Trigger:      triggerType,
		Amount:       amount,
		Metrics:      data,
	})
	return err
}

func (c *Client) modelScalingEventsConsumer() {
	ch := c.modelScalingService.(*modelscaling.StatsAnalyserService).GetEventChannel()
	for c.modelScalingService.Ready() {
		e := <-ch
		modelName, modelVersion, err := util.GetOrignalModelNameAndVersion(e.StatsData.ModelName)
		if err != nil {
			c.logger.WithError(err).Warnf(
				"Trigger model scaling event %d for model %s failed",
				e.EventType,
				e.StatsData.ModelName,
			)
			continue
		}

		c.logger.Debugf(
			"Trigger model scaling event %d for model %s:%d with %s %d",
			e.EventType,
			modelName,
			modelVersion,
			e.StatsData.Key,
			e.StatsData.Value,
		)

		err = c.sendModelScalingTriggerEvent(
			modelName, modelVersion, e.EventType, e.StatsData.Value, nil,
		)
		if err != nil {
			c.logger.WithError(err).Warnf(
				"Sending model scaling event %d for model %s failed",
				e.EventType,
				e.StatsData.ModelName,
			)
			continue
		}
	}
}

func (c *Client) startSubServiceChecker() {
	ticker := time.NewTicker(c.settings.periodReadySubService)
	defer ticker.Stop()
	for !c.stop.Load() {
		<-ticker.C
		if err := c.WaitReadySubServices(false); err != nil {
			c.Stop()
		}
	}
}
