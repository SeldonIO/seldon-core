package agent

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	k8s "github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type Client struct {
	mu         sync.RWMutex
	logger     log.FieldLogger
	configChan chan AgentConfiguration
	ClientState
	ClientServices
	SchedulerGrpcClientOptions
	KubernetesOptions
}

type ClientState struct {
	replicaConfig *agent.ReplicaConfig
	loadedModels  map[string]*pbs.ModelDetails
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
	RCloneClient *RCloneClient
	V2Client     *V2Client
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
	rcloneClient *RCloneClient,
	v2Client *V2Client,
	replicaConfig *agent.ReplicaConfig,
	inferenceSvcName string,
	namespace string) *Client {

	replicaConfig.InferenceSvc = inferenceSvcName
	replicaConfig.AvailableMemoryBytes = replicaConfig.MemoryBytes

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &Client{
		logger:     logger.WithField("Name", "Client"),
		configChan: make(chan AgentConfiguration),
		ClientState: ClientState{
			replicaConfig: replicaConfig,
			loadedModels:  make(map[string]*pbs.ModelDetails),
		},
		ClientServices: ClientServices{
			RCloneClient: rcloneClient,
			V2Client:     v2Client,
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

func (c *Client) Start(configHandler *AgentConfigHandler) error {
	logger := c.logger.WithField("func", "Start")
	if configHandler == nil {
		return fmt.Errorf("configHandler is nil. Can't start client grpc server.")
	}
	err := c.waitReady()
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to wait for all agent dependent services to be ready")
		return err
	}
	if c.conn == nil {
		err = c.createConnection()
		if err != nil {
			c.logger.WithError(err).Errorf("Failed to create connection to scheduler")
			return err
		}
	}
	// Start config listener
	go c.listenForConfigUpdates()
	// Add ourself as listener on channel and handle initial config
	logger.Info("Loadining initial rclone configuration")
	err = c.loadRcloneConfiguration(configHandler.AddListener(c.configChan))
	if err != nil {
		c.logger.WithError(err).Fatal("Failed to load rclone defaults")
		return err
	}
	logFailure := func(err error, delay time.Duration) {
		c.logger.WithError(err).Errorf("Scheduler not ready")
	}
	err = backoff.RetryNotify(c.StartService, backoff.NewExponentialBackOff(), logFailure)
	if err != nil {
		c.logger.WithError(err).Fatal("Failed to start client")
		return err
	}
	return nil
}

func (c *Client) listenForConfigUpdates() {
	logger := c.logger.WithField("func", "listenForConfigUpdates")
	for config := range c.configChan {
		c.logger.Info("Received config update")
		config := config
		go func() {
			err := c.loadRcloneConfiguration(&config)
			if err != nil {
				logger.WithError(err).Error("Failed to load rclone defaults")
			}
		}()
	}
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

func (c *Client) waitReady() error {
	logger := c.logger.WithField("func", "waitReady")
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Rclone not ready")
	}
	logger.Infof("Waiting for Rclone server to be ready")
	err := backoff.RetryNotify(c.RCloneClient.Ready, backoff.NewExponentialBackOff(), logFailure)
	if err != nil {
		return err
	}
	logFailure = func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Server not ready")
	}
	logger.Infof("Waiting for inference server to be ready")
	err = backoff.RetryNotify(c.V2Client.Ready, backoff.NewExponentialBackOff(), logFailure)
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
	var loadedModels []*pbs.ModelDetails
	for _, v := range c.loadedModels {
		loadedModels = append(loadedModels, v)
	}
	stream, err := grpcClient.Subscribe(context.Background(), &agent.AgentSubscribeRequest{
		ServerName:    c.serverName,
		ReplicaIdx:    c.replicaIdx,
		ReplicaConfig: c.replicaConfig,
		LoadedModels:  loadedModels,
		Shared:        true,
	}, grpc_retry.WithMax(100))
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
			err := c.LoadModel(operation)
			if err != nil {
				c.logger.WithError(err).Errorf("Failed to handle load model")
			}
		case agent.ModelOperationMessage_UNLOAD_MODEL:
			c.logger.Infof("calling unload model")
			err := c.UnloadModel(operation)
			if err != nil {
				c.logger.WithError(err).Errorf("Failed to handle unload model")
			}
		}
	}
	return nil
}

func (c *Client) sendModelEventError(modelName string, modelVersion string, event agent.ModelEventMessage_Event, err error) {
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         modelVersion,
		Event:                event,
		Message:              err.Error(),
		AvailableMemoryBytes: c.replicaConfig.AvailableMemoryBytes,
	})
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to send error back on load model")
	}
}

func (c *Client) getArtifactConfig(request *agent.ModelOperationMessage) ([]byte, error) {
	if request.Details.StorageConfig == nil {
		return nil, nil
	}
	logger := c.logger.WithField("func", "setupArtifactConfig")
	logger.Infof("Getting Rclone configuration")
	switch x := request.Details.StorageConfig.Config.(type) {
	case *pbs.StorageConfig_StorageRcloneConfig:
		return []byte(x.StorageRcloneConfig), nil
	case *pbs.StorageConfig_StorageSecretName:
		if c.secretsHandler == nil {
			secretClientSet, err := k8s.CreateClientset()
			if err != nil {
				return nil, err
			}
			c.KubernetesOptions.secretsHandler = k8s.NewSecretsHandler(secretClientSet, c.namespace)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	if request == nil || request.Details == nil {
		return fmt.Errorf("Empty request received for load model")
	}
	modelName := request.Details.Name
	if request.Details.GetMemoryBytes() > c.replicaConfig.AvailableMemoryBytes {
		err := fmt.Errorf("Not enough memory on replica for model %s available %d requested %d", modelName, c.replicaConfig.AvailableMemoryBytes, request.Details.GetMemoryBytes())
		c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}

	logger.Infof("Load model %s", modelName)

	// Get Rclone configuration
	config, err := c.getArtifactConfig(request)
	if err != nil {
		c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	// Copy model artifact using RClone
	err = c.RCloneClient.Copy(request.Details.Name, request.Details.Uri, config)
	if err != nil {
		c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	err = c.V2Client.LoadModel(modelName)
	if err != nil {
		c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	logger.Infof("Load model %s success", modelName)
	loadedModel, ok := c.loadedModels[modelName]
	if ok {
		c.replicaConfig.AvailableMemoryBytes = c.replicaConfig.AvailableMemoryBytes + loadedModel.GetMemoryBytes()
	}
	c.loadedModels[modelName] = request.Details
	c.replicaConfig.AvailableMemoryBytes = c.replicaConfig.AvailableMemoryBytes - request.Details.GetMemoryBytes()
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         request.Details.GetVersion(),
		Event:                agent.ModelEventMessage_LOADED,
		AvailableMemoryBytes: c.replicaConfig.AvailableMemoryBytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UnloadModel(request *agent.ModelOperationMessage) error {
	logger := c.logger.WithField("func", "UnloadModel")
	c.mu.Lock()
	defer c.mu.Unlock()
	if request == nil || request.Details == nil {
		return fmt.Errorf("Empty request received for load model")
	}
	modelName := request.Details.Name
	logger.Infof("Unload model %s", modelName)
	err := c.V2Client.UnloadModel(modelName)
	if err != nil {
		c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}
	loadedModel, ok := c.loadedModels[modelName]
	if !ok {
		err := fmt.Errorf("Unknown model with name %s", modelName)
		c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}
	logger.Infof("Unload model %s success", modelName)
	delete(c.loadedModels, modelName)
	c.replicaConfig.AvailableMemoryBytes = c.replicaConfig.AvailableMemoryBytes + loadedModel.GetMemoryBytes()
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ModelName:            modelName,
		ModelVersion:         loadedModel.GetVersion(),
		Event:                agent.ModelEventMessage_UNLOADED,
		AvailableMemoryBytes: c.replicaConfig.AvailableMemoryBytes,
	})
	if err != nil {
		return err
	}
	return nil
}
