package agent

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"

	backoff "github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	k8s "github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/repository"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type ClientServiceInterface interface {
	SetState(state *LocalStateManager)
	Start() error
	Ready() bool
	Stop() error
	Name() string
}

type Client struct {
	logger             log.FieldLogger
	configChan         chan config.AgentConfiguration
	replicaConfig      *agent.ReplicaConfig
	stateManager       *LocalStateManager
	rpHTTP             ClientServiceInterface
	rpGRPC             ClientServiceInterface
	clientDebugService ClientServiceInterface
	metrics            metrics.MetricsHandler
	ClientServices
	SchedulerGrpcClientOptions
	KubernetesOptions
}

type SchedulerGrpcClientOptions struct {
	schedulerHost string
	schedulerPort int
	serverName    string
	replicaIdx    uint32
	conn          *grpc.ClientConn
	callOptions   []grpc.CallOption
}

type KubernetesOptions struct {
	secretsHandler *k8s.SecretHandler
	namespace      string
}

type ClientServices struct {
	ModelRepository repository.ModelRepository
}

func ParseReplicaConfig(json string) (*agent.ReplicaConfig, error) {
	config := agent.ReplicaConfig{}
	err := protojson.Unmarshal([]byte(json), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func NewClient(serverName string,
	replicaIdx uint32,
	schedulerHost string,
	schedulerPort int,
	logger log.FieldLogger,
	modelRepository repository.ModelRepository,
	v2Client *V2Client,
	replicaConfig *agent.ReplicaConfig,
	inferenceSvcName string,
	namespace string,
	reverseProxyHTTP ClientServiceInterface,
	reverseProxyGRPC ClientServiceInterface,
	clientDebugService ClientServiceInterface,
	metrics metrics.MetricsHandler,
) *Client {

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	modelState := NewModelState()

	stateManager := NewLocalStateManager(
		modelState, logger, v2Client, replicaConfig.GetMemoryBytes(), replicaConfig.GetOverCommitPercentage(), metrics)

	clientDebugService.SetState(stateManager)
	reverseProxyHTTP.SetState(stateManager)
	reverseProxyGRPC.SetState(stateManager)

	return &Client{
		logger:             logger.WithField("Name", "Client"),
		configChan:         make(chan config.AgentConfiguration),
		stateManager:       stateManager,
		replicaConfig:      replicaConfig,
		rpHTTP:             reverseProxyHTTP,
		rpGRPC:             reverseProxyGRPC,
		clientDebugService: clientDebugService,
		metrics:            metrics,
		ClientServices: ClientServices{
			ModelRepository: modelRepository,
		},
		SchedulerGrpcClientOptions: SchedulerGrpcClientOptions{
			schedulerHost: schedulerHost,
			schedulerPort: schedulerPort,
			serverName:    serverName,
			replicaIdx:    replicaIdx,
			callOptions:   opts,
		},
		KubernetesOptions: KubernetesOptions{
			namespace: namespace,
		},
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
	return nil
}

func (c *Client) createConnection() error {
	c.logger.Infof("Creating connection to %s:%d", c.schedulerHost, c.schedulerPort)
	conn, err := getConnection(c.schedulerHost, c.schedulerPort)
	if err != nil {
		return err
	}
	c.SchedulerGrpcClientOptions.conn = conn
	return nil
}

func (c *Client) WaitReady() error {
	logger := c.logger.WithField("func", "waitReady")

	// Wait for model repo to be ready
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Rclone not ready")
	}
	logger.Infof("Waiting for Model Repository to be ready")
	//TODO make retry configurable
	err := backoff.RetryNotify(c.ModelRepository.Ready, backoff.NewExponentialBackOff(), logFailure)
	if err != nil {
		return err
	}

	// Wait for V2 Server to be ready
	logFailure = func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Server not ready")
	}
	logger.Infof("Waiting for inference server to be ready")
	err = backoff.RetryNotify(c.stateManager.v2Client.Ready, backoff.NewExponentialBackOff(), logFailure)
	if err != nil {
		return err
	}

	// http reverse proxy
	if err := startSubService(c.rpHTTP, logger); err != nil {
		return err
	}

	// grpc reverse proxy
	if err := startSubService(c.rpGRPC, logger); err != nil {
		return err
	}

	// debug service
	if err := startSubService(c.clientDebugService, logger); err != nil {
		return err
	}

	return nil
}

func startSubService(service ClientServiceInterface, logger *log.Entry) error {
	// debug service
	logger.Infof("Starting and waiting for %s", service.Name())
	err := service.Start()
	if err != nil {
		return err
	}

	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("%s service not ready", service.Name())
	}

	readyToError := func() error {
		if service.Ready() {
			return nil
		} else {
			return fmt.Errorf("Service %s not ready", service.Name())
		}
	}
	err = backoff.RetryNotify(readyToError, backoff.NewExponentialBackOff(), logFailure)
	return err
}

func getConnection(host string, port int) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(grpc_retry.UnaryClientInterceptor(retryOpts...), otelgrpc.UnaryClientInterceptor())),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) StartService() error {
	c.logger.Infof("Call subscribe to scheduler")
	grpcClient := agent.NewAgentServiceClient(c.conn)

	stream, err := grpcClient.Subscribe(context.Background(), &agent.AgentSubscribeRequest{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ReplicaConfig:        c.replicaConfig,
		LoadedModels:         c.stateManager.modelVersions.getVersionsForAllModels(),
		Shared:               true,
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytesWithOverCommit(),
	}, grpc_retry.WithMax(100)) //TODO make configurable
	if err != nil {
		return err
	}

	go c.metrics.AddServerReplicaMetrics(
		c.stateManager.totalMainMemoryBytes,
		float32(c.stateManager.totalMainMemoryBytes)+c.stateManager.GetOverCommitMemoryBytes())

	for {
		operation, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		c.logger.Infof("Received operation")
		switch operation.Operation {
		case agent.ModelOperationMessage_LOAD_MODEL:
			c.logger.Infof("calling load model")
			go func() {
				err := c.LoadModel(operation)
				if err != nil {
					c.logger.WithError(err).Errorf("Failed to handle load model")
				}
			}()

		case agent.ModelOperationMessage_UNLOAD_MODEL:
			c.logger.Infof("calling unload model")
			go func() {
				err := c.UnloadModel(operation)
				if err != nil {
					c.logger.WithError(err).Errorf("Failed to handle unload model")
				}
			}()
		}
	}
	return nil
}

func (c *Client) getArtifactConfig(request *agent.ModelOperationMessage) ([]byte, error) {
	if request.GetModelVersion().GetModel().GetModelSpec().StorageConfig == nil {
		return nil, nil
	}
	logger := c.logger.WithField("func", "getArtifactConfig")
	logger.Infof("Getting Rclone configuration")
	switch x := request.GetModelVersion().GetModel().GetModelSpec().StorageConfig.Config.(type) {
	case *pbs.StorageConfig_StorageRcloneConfig:
		return []byte(x.StorageRcloneConfig), nil
	case *pbs.StorageConfig_StorageSecretName:
		if c.secretsHandler == nil {
			secretClientSet, err := k8s.CreateClientset()
			if err != nil {
				return nil, err
			}
			if request.GetModelVersion().GetModel().GetMeta().GetKubernetesMeta() != nil {
				c.KubernetesOptions.secretsHandler = k8s.NewSecretsHandler(secretClientSet, request.GetModelVersion().GetModel().GetMeta().GetKubernetesMeta().GetNamespace())
			} else {
				return nil, fmt.Errorf("Can't load model %s:%dwith k8s secret %s when namespace not set", request.GetModelVersion().GetModel().GetMeta().GetName(), request.GetModelVersion().GetVersion(), x.StorageSecretName)
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
	logger := c.logger.WithField("func", "LoadModel")
	if request == nil || request.ModelVersion == nil {
		return fmt.Errorf("Empty request received for load model")
	}
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
		modelWithVersion, pinnedModelVersion, request.GetModelVersion().GetModel().GetModelSpec().ArtifactVersion,
		request.GetModelVersion().GetModel().GetModelSpec().Uri, config)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	logger.Infof("Chose path %s for model %s:%d", *chosenVersionPath, modelName, modelVersion)

	// TODO: do we need the actual protos being sent
	modifiedModelVersionRequest := getModifiedModelVersion(modelWithVersion, pinnedModelVersion, request.GetModelVersion())
	err = c.stateManager.LoadModelVersion(modifiedModelVersionRequest)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}

	logger.Infof("Load model %s:%d success", modelName, modelVersion)
	return c.sendAgentEvent(modelName, modelVersion, agent.ModelEventMessage_LOADED)
}

func (c *Client) UnloadModel(request *agent.ModelOperationMessage) error {
	logger := c.logger.WithField("func", "UnloadModel")
	if request == nil || request.GetModelVersion() == nil {
		return fmt.Errorf("Empty request received for unload model")
	}
	modelName := request.GetModelVersion().GetModel().GetMeta().GetName()
	modelVersion := request.GetModelVersion().GetVersion()
	modelWithVersion := util.GetVersionedModelName(modelName, modelVersion)
	pinnedModelVersion := util.GetPinnedModelVersion()

	c.stateManager.cache.Lock(modelWithVersion)
	defer c.stateManager.cache.Unlock(modelWithVersion)

	logger.Infof("Unload model %s:%d", modelName, modelVersion)

	_, err := c.ModelRepository.RemoveModelVersion(modelWithVersion, pinnedModelVersion)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	// we do not care about model versions here
	modifiedModelVersionRequest := getModifiedModelVersion(modelWithVersion, pinnedModelVersion, request.GetModelVersion())
	if err := c.stateManager.UnloadModelVersion(modifiedModelVersionRequest); err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	logger.Infof("Unload model %s:%d success", modelName, modelVersion)
	return c.sendAgentEvent(modelName, modelVersion, agent.ModelEventMessage_UNLOADED)
}

func (c *Client) sendModelEventError(modelName string, modelVersion uint32, event agent.ModelEventMessage_Event, err error) {
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         modelVersion,
		Event:                event,
		Message:              err.Error(),
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytesWithOverCommit(),
	})
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to send error back on load model")
	}
}

func (c *Client) sendAgentEvent(modelName string, modelVersion uint32, event agent.ModelEventMessage_Event) error {
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

func getModifiedModelVersion(modelId string, version uint32, originalModelVersion *agent.ModelVersion) *agent.ModelVersion {
	mv := proto.Clone(originalModelVersion)
	mv.(*agent.ModelVersion).Model.Meta.Name = modelId
	mv.(*agent.ModelVersion).Version = version
	return mv.(*agent.ModelVersion)
}
