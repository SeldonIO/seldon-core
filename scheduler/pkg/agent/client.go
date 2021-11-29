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
	mu             sync.RWMutex
	schedulerHost  string
	schedulerPort  int
	serverName     string
	replicaIdx     uint32
	conn           *grpc.ClientConn
	callOptions    []grpc.CallOption
	logger         log.FieldLogger
	RCloneClient   *RCloneClient
	V2Client       *V2Client
	replicaConfig  *agent.ReplicaConfig
	loadedModels   map[string]*pbs.ModelDetails
	secretsHandler *k8s.SecretHandler
	namespace      string
	configHandler  *AgentConfigHandler
	configChan     chan string
}

func ParseReplicConfig(json string) (*agent.ReplicaConfig, error) {
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
	namespace string,
	configHandler *AgentConfigHandler) (*Client, error) {
	replicaConfig.InferenceSvc = inferenceSvcName
	replicaConfig.AvailableMemoryBytes = replicaConfig.MemoryBytes

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &Client{
		schedulerHost: schedulerHost,
		schedulerPort: schedulerPort,
		serverName:    serverName,
		replicaIdx:    replicaIdx,
		callOptions:   opts,
		logger:        logger.WithField("Name", "Client"),
		RCloneClient:  rcloneClient,
		V2Client:      v2Client,
		replicaConfig: replicaConfig,
		loadedModels:  make(map[string]*pbs.ModelDetails),
		namespace:     namespace,
		configHandler: configHandler,
		configChan:    make(chan string),
	}, nil
}

func (c *Client) Start() error {
	err := c.waitReady()
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to create connection")
		return err
	}
	if c.conn == nil {
		err = c.createConnection()
		if err != nil {
			c.logger.WithError(err).Errorf("Failed to create connection")
			return err
		}
	}
	// Start config listener
	go c.listenForConfigUpdates()
	// Add ourself as listener on channel and handle initial config
	err = c.loadRcloneDefaults(c.configHandler.AddListener(c.configChan))
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
	for range c.configChan {
		c.logger.Info("Received config update")
		go func() {
			err := c.loadRcloneDefaults(c.configHandler.getConfiguration())
			if err != nil {
				logger.WithError(err).Error("Failed to load rclone defaults")
			}
		}()
	}
}

func (c *Client) loadRcloneDefaults(rcloneConfig *AgentConfiguration) error {
	logger := c.logger.WithField("func", "loadRcloneDefaults")
	var rcloneNamesAdded []string
	if rcloneConfig != nil {
		// Load any secrets that have Rclone config
		if len(rcloneConfig.Rclone.ConfigSecrets) > 0 {
			secretClientSet, err := k8s.CreateSecretsClientset()
			if err != nil {
				return err
			}
			secretsHandler := k8s.NewSecretsHandler(secretClientSet, c.namespace)
			for _, secret := range rcloneConfig.Rclone.ConfigSecrets {
				logger.Infof("Loading rclone secret %s", secret)
				config, err := secretsHandler.GetSecretConfig(secret)
				if err != nil {
					return err
				}
				name, err := c.RCloneClient.Config(config)
				if err != nil {
					return err
				}
				rcloneNamesAdded = append(rcloneNamesAdded, name)
			}
		}
		// Load any raw Rclone configs
		if len(rcloneConfig.Rclone.Config) > 0 {
			for _, config := range rcloneConfig.Rclone.Config {
				logger.Infof("Loading rclone config %s", config)
				name, err := c.RCloneClient.Config([]byte(config))
				if err != nil {
					return err
				}
				rcloneNamesAdded = append(rcloneNamesAdded, name)
			}
		}

		// Delete any existing remotes not in defaults
		exsitingNames, err := c.RCloneClient.ListRemotes()
		if err != nil {
			return err
		}
		for _, existingName := range exsitingNames {
			found := false
			for _, addedName := range rcloneNamesAdded {
				if existingName == addedName {
					found = true
					break
				}
			}
			if !found {
				logger.Warnf("Delete remote %s as not in new list of defaults", existingName)
				err := c.RCloneClient.DeleteRemote(existingName)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Client) createConnection() error {
	c.logger.Infof("Creating connection to %s:%d", c.schedulerHost, c.schedulerPort)
	conn, err := getConnection(c.schedulerHost, c.schedulerPort)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) waitReady() error {
	logFailure := func(err error, delay time.Duration) {
		c.logger.WithError(err).Errorf("Rclone not ready")
	}
	err := backoff.RetryNotify(c.RCloneClient.Ready, backoff.NewExponentialBackOff(), logFailure)
	if err != nil {
		return err
	}
	logFailure = func(err error, delay time.Duration) {
		c.logger.WithError(err).Errorf("Server not ready")
	}
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
			secretClientSet, err := k8s.CreateSecretsClientset()
			if err != nil {
				return nil, err
			}
			c.secretsHandler = k8s.NewSecretsHandler(secretClientSet, c.namespace)
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
