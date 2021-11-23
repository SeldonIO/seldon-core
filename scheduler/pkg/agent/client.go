package agent

import (
	"context"
	"fmt"
	backoff "github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	k8s "github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"math"
	"sync"
	"time"
)

type Client struct {
	mu sync.RWMutex
	schedulerHost string
	schedulerPort int
	serverName string
	replicaIdx uint32
	conn *grpc.ClientConn
	callOptions []grpc.CallOption
	logger log.FieldLogger
	RCloneClient *RCloneClient
	V2Client *V2Client
	replicaConfig *agent.ReplicaConfig
	loadedModels map[string]*pbs.ModelDetails
	secretsHandler *k8s.SecretHandler
	namespace string
}

func ParseReplicConfig(json string) (*agent.ReplicaConfig, error) {
	config := agent.ReplicaConfig{}
	err := protojson.Unmarshal([]byte(json),&config)
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
	namespace string)  (*Client, error) {
	replicaConfig.InferenceSvc = inferenceSvcName
	replicaConfig.AvailableMemoryBytes = replicaConfig.MemoryBytes

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &Client{
		schedulerHost: schedulerHost,
		schedulerPort: schedulerPort,
		serverName: serverName,
		replicaIdx: replicaIdx,
		callOptions: opts,
		logger: logger.WithField("Name","Client"),
		RCloneClient: rcloneClient,
		V2Client: v2Client,
		replicaConfig: replicaConfig,
		loadedModels: make(map[string]*pbs.ModelDetails),
		namespace: namespace,
	}, nil
}

func (c *Client) CreateConnection() error {
	c.logger.Infof("Creating connection to %s:%d",c.schedulerHost,c.schedulerPort)
	conn, err := getConnection(c.schedulerHost, c.schedulerPort)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func ( c *Client) WaitReady() error {
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


func (c *Client) Start() error {
	c.logger.Infof("Call subscribe to scheduler")
	grpcClient := agent.NewAgentServiceClient(c.conn)
	var loadedModels []*pbs.ModelDetails
	for _,v := range c.loadedModels {
		loadedModels = append(loadedModels,v)
	}
	stream, err := grpcClient.Subscribe(context.Background(), &agent.AgentSubscribeRequest{
		ServerName: c.serverName,
		ReplicaIdx: c.replicaIdx,
		ReplicaConfig: c.replicaConfig,
		LoadedModels: loadedModels,
		Shared: true,
	},grpc_retry.WithMax(100))
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
			err  := c.LoadModel(operation)
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

func (c *Client) sendModelEventError(modelName string, modelVersion string, event agent.ModelEventMessage_Event, err error) error {
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName: c.serverName,
		ReplicaIdx: c.replicaIdx,
		ModelName: modelName,
		ModelVersion: modelVersion,
		Event: event,
		Message: err.Error(),
		AvailableMemoryBytes: c.replicaConfig.AvailableMemoryBytes,
	})
	return err
}

func (c *Client) LoadModel(request *agent.ModelOperationMessage) error  {
	c.mu.Lock()
	defer c.mu.Unlock()
	if request == nil || request.Details == nil {
		return fmt.Errorf("Empty request received for load model")
	}
	modelName := request.Details.Name
	if request.Details.GetMemoryBytes() > c.replicaConfig.AvailableMemoryBytes {
		err := fmt.Errorf("Not enough memory on replica for model %s available %d requested %d",modelName,c.replicaConfig.AvailableMemoryBytes, request.Details.GetMemoryBytes())
		err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		if err2 != nil {
			c.logger.WithError(err2).Errorf("Failed to send error back on load model")
		}
		return err
	}

	c.logger.Infof("Load model %s", modelName)

	// Load model storage configuration before copying if needed
	if request.Details.StorageRCloneConfig != nil { // Load rclone config from model details
		err := c.RCloneClient.Config(request.Details.GetName(), request.Details.GetVersion(), []byte(request.Details.GetStorageRCloneConfig()))
		if err != nil {
			err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
			if err2 != nil {
				c.logger.WithError(err2).Errorf("Failed to send error back on load model")
			}
			return err
		}
	} else if request.Details.StorageSecretName != nil { // Load rclone config from k8s secret
		if c.secretsHandler == nil {
			secretClientSet, err := k8s.CreateSecretsClientset()
			if err != nil {
				err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
				if err2 != nil {
					c.logger.WithError(err2).Errorf("Failed to send error back on load model")
				}
				return err
			}
			c.secretsHandler = k8s.NewSecretsHandler(secretClientSet, c.namespace)
		}
		config, err := c.secretsHandler.GetSecretConfig(request.Details.GetStorageSecretName())
		if err != nil {
			err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
			if err2 != nil {
				c.logger.WithError(err2).Errorf("Failed to send error back on load model")
			}
			return err
		}
		err = c.RCloneClient.Config(request.Details.GetName(), request.Details.GetVersion(), config)
		if err != nil {
			err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
			if err2 != nil {
				c.logger.WithError(err2).Errorf("Failed to send error back on load model")
			}
			return err
		}
	}

	// Copy model artifact using RClone
	err := c.RCloneClient.Copy(request.Details.Name, request.Details.GetVersion(), request.Details.Uri)
	if err != nil {
		err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		if err2 != nil {
			c.logger.WithError(err2).Errorf("Failed to send error back on load model")
		}
		return err
	}
	err = c.V2Client.LoadModel(modelName)
	if err != nil {
		err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_LOAD_FAILED, err)
		if err2 != nil {
			c.logger.WithError(err2).Errorf("Failed to send error back on load model")
		}
		return err
	}
	c.logger.Infof("Load model %s success", modelName)
	loadedModel, ok := c.loadedModels[modelName]
	if ok {
		c.replicaConfig.AvailableMemoryBytes = c.replicaConfig.AvailableMemoryBytes + loadedModel.GetMemoryBytes()
	}
	c.loadedModels[modelName] = request.Details
	c.replicaConfig.AvailableMemoryBytes = c.replicaConfig.AvailableMemoryBytes - request.Details.GetMemoryBytes()
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName: c.serverName,
		ReplicaIdx: c.replicaIdx,
		ModelName: modelName,
		ModelVersion: request.Details.GetVersion(),
		Event: agent.ModelEventMessage_LOADED,
		AvailableMemoryBytes: c.replicaConfig.AvailableMemoryBytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UnloadModel(request *agent.ModelOperationMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if request == nil || request.Details == nil {
		return fmt.Errorf("Empty request received for load model")
	}
	modelName := request.Details.Name
	c.logger.Infof("Unload model %s", modelName)
	err := c.V2Client.UnloadModel(modelName)
	if err != nil {
		err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_UNLOAD_FAILED, err)
		if err2 != nil {
			c.logger.WithError(err2).Errorf("Failed to send error back on unload model")
		}
		return err
	}
	loadedModel, ok := c.loadedModels[modelName]
	if !ok {
		err := fmt.Errorf("Unknown model with name %s", modelName)
		err2 := c.sendModelEventError(modelName, request.Details.GetVersion(), agent.ModelEventMessage_UNLOAD_FAILED, err)
		if err2 != nil {
			c.logger.WithError(err2).Errorf("Failed to send error back on unload model")
		}
		return err
	}
	c.logger.Infof("Unload model %s success", modelName)
	delete(c.loadedModels, modelName)
	c.replicaConfig.AvailableMemoryBytes = c.replicaConfig.AvailableMemoryBytes + loadedModel.GetMemoryBytes()
	grpcClient := agent.NewAgentServiceClient(c.conn)
	_, err = grpcClient.AgentEvent(context.Background(), &agent.ModelEventMessage{
		ServerName: c.serverName,
		ReplicaIdx: c.replicaIdx,
		ModelName: modelName,
		ModelVersion: loadedModel.GetVersion(),
		Event: agent.ModelEventMessage_UNLOADED,
		AvailableMemoryBytes: c.replicaConfig.AvailableMemoryBytes,
	})
	if err != nil {
		return err
	}
	return nil
}