package agent

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"

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
	Ready() error
	Stop() error
}

type Client struct {
	logger             log.FieldLogger
	configChan         chan config.AgentConfiguration
	replicaConfig      *agent.ReplicaConfig
	stateManager       *LocalStateManager
	rpHTTP             ClientServiceInterface
	clientDebugService ClientServiceInterface
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
	clientDebugService ClientServiceInterface,
) *Client {

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	modelState := NewModelState()

	stateManager := NewLocalStateManager(modelState, logger, v2Client, int64(replicaConfig.GetMemoryBytes()))

	clientDebugService.SetState(stateManager)
	reverseProxyHTTP.SetState(stateManager)

	return &Client{
		logger:             logger.WithField("Name", "Client"),
		configChan:         make(chan config.AgentConfiguration),
		stateManager:       stateManager,
		replicaConfig:      replicaConfig,
		rpHTTP:             reverseProxyHTTP,
		clientDebugService: clientDebugService,
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

	// TODO: move this outside and perhaps add to ClientServices

	logger.Infof("Starting and waiting for Reverse Proxy to be ready")
	err = c.rpHTTP.Start()
	if err != nil {
		return err
	}
	logFailure = func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("HTTP reverse proxy not ready")
	}
	err = backoff.RetryNotify(c.rpHTTP.Ready, backoff.NewExponentialBackOff(), logFailure)
	if err != nil {
		return err
	}

	// TODO: move this outside and perhaps add to ClientServices
	logger.Infof("Starting client debug service")
	err = c.clientDebugService.Start()
	if err != nil {
		return err
	}

	return nil
}

func getConnection(host string, port int) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
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
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytes(),
	}, grpc_retry.WithMax(100)) //TODO make configurable
	if err != nil {
		return err
	}
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

	c.stateManager.modelLoadLockCreate(modelName)
	defer c.stateManager.modelLoadUnlock(modelName)

	logger.Infof("Load model %s:%d", modelName, modelVersion)

	// Get Rclone configuration
	config, err := c.getArtifactConfig(request)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	// Copy model artifact
	chosenVersionPath, err := c.ModelRepository.DownloadModelVersion(
		modelName, modelVersion, request.GetModelVersion().GetModel().GetModelSpec().ArtifactVersion,
		request.GetModelVersion().GetModel().GetModelSpec().Uri, config)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	logger.Infof("Chose path %s for model %s:%d", *chosenVersionPath, modelName, modelVersion)

	err = c.stateManager.LoadModelVersion(request.GetModelVersion())
	if err != nil {
		c.stateManager.modelVersions.removeModelVersion(request.GetModelVersion())
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

	c.stateManager.modelLoadLockCreate(modelName)
	defer c.stateManager.modelLoadUnlock(modelName)

	logger.Infof("Unload model %s:%d", modelName, modelVersion)

	_, err := c.ModelRepository.RemoveModelVersion(modelName, modelVersion)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}

	if err := c.stateManager.UnloadModelVersion(request.ModelVersion); err != nil {
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
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytes(),
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
		AvailableMemoryBytes: c.stateManager.GetAvailableMemoryBytes(),
	})
	return err
}
