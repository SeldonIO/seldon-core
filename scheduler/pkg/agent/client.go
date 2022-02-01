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

type Client struct {
	logger     log.FieldLogger
	configChan chan config.AgentConfiguration
	*ClientState
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
	V2Client        *V2Client
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
	namespace string) *Client {

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &Client{
		logger:      logger.WithField("Name", "Client"),
		configChan:  make(chan config.AgentConfiguration),
		ClientState: NewClientState(replicaConfig),
		ClientServices: ClientServices{
			ModelRepository: modelRepository,
			V2Client:        v2Client,
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
	err := c.waitReady()
	if err != nil {
		logger.WithError(err).Errorf("Failed to wait for all agent dependent services to be ready")
		return err
	}
	if c.conn == nil {
		err = c.createConnection()
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
	err = backoff.RetryNotify(c.StartService, backOffExp, logFailure)
	if err != nil {
		c.logger.WithError(err).Fatal("Failed to start client")
		return err
	}
	return nil
}

func (c *Client) createConnection() error {
	logger := c.logger.WithField("func", "createConnection")
	logger.Infof("Creating connection to %s:%d", c.schedulerHost, c.schedulerPort)
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
		logger.WithError(err).Errorf("Model repository not ready")
	}
	logger.Infof("Waiting for Model Repository to be ready")
	//TODO make retry configurable
	err := backoff.RetryNotify(c.ModelRepository.Ready, backoff.NewExponentialBackOff(), logFailure)
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
	logger := c.logger.WithField("func", "StartService")
	logger.Infof("Call subscribe to scheduler")
	grpcClient := agent.NewAgentServiceClient(c.conn)
	var loadedModels []*agent.ModelVersion
	for _, mv := range c.loadedModels {
		for _, md := range mv.versions {
			loadedModels = append(loadedModels, md)
		}

	}
	stream, err := grpcClient.Subscribe(context.Background(), &agent.AgentSubscribeRequest{
		ServerName:           c.serverName,
		ReplicaIdx:           c.replicaIdx,
		ReplicaConfig:        c.replicaConfig,
		LoadedModels:         loadedModels,
		Shared:               true,
		AvailableMemoryBytes: c.availableMemoryBytes,
	}, grpc_retry.WithMax(100)) //TODO make configurable
	if err != nil {
		return err
	}
	logger.Infof("Subscribed to scheduler. Listening for events...")
	for {
		operation, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		logger.Infof("Received operation")
		switch operation.Operation {
		case agent.ModelOperationMessage_LOAD_MODEL:
			logger.Infof("calling load model")
			go func() {
				err := c.LoadModel(operation)
				if err != nil {
					c.logger.WithError(err).Errorf("Failed to handle load model")
				}
			}()

		case agent.ModelOperationMessage_UNLOAD_MODEL:
			logger.Infof("calling unload model")
			go func() {
				err := c.UnloadModel(operation)
				if err != nil {
					logger.WithError(err).Errorf("Failed to handle unload model")
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
	logger.Infof("Load model %s:%d", modelName, modelVersion)

	err := c.addModelVersion(request.GetModelVersion())
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}

	// Get Rclone configuration
	config, err := c.getArtifactConfig(request)
	if err != nil {
		c.removeModelVersion(request.GetModelVersion())
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	// Copy model artifact
	chosenVersionPath, err := c.ModelRepository.DownloadModelVersion(modelName, modelVersion, request.GetModelVersion().GetModel().GetModelSpec().ArtifactVersion, request.GetModelVersion().GetModel().GetModelSpec().Uri, config)
	if err != nil {
		c.removeModelVersion(request.GetModelVersion())
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_LOAD_FAILED, err)
		return err
	}
	logger.Infof("Chose path %s for model %s:%d", *chosenVersionPath, modelName, modelVersion)
	err = c.V2Client.LoadModel(modelName)
	if err != nil {
		c.removeModelVersion(request.GetModelVersion())
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
	logger.Infof("Unload model %s:%d", modelName, modelVersion)

	remainingVersions, err := c.ModelRepository.RemoveModelVersion(modelName, modelVersion)
	if err != nil {
		c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
		return err
	}
	if remainingVersions == 0 {
		err := c.V2Client.UnloadModel(modelName)
		if err != nil {
			c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
			return err
		}
	} else {
		err := c.V2Client.LoadModel(modelName) // Force a reload to resync server with available versions on disk
		if err != nil {
			c.sendModelEventError(modelName, modelVersion, agent.ModelEventMessage_UNLOAD_FAILED, err)
			return err
		}
	}

	removedModel := c.removeModelVersion(request.GetModelVersion())
	noRemainingVersions := remainingVersions == 0
	if noRemainingVersions != removedModel {
		c.logger.Warnf("Mismatch in state. Removed all versions from state is [%v] but model repo says remaining versions [%d] for %s:%d", removedModel, remainingVersions, modelName, modelVersion)
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
		AvailableMemoryBytes: c.availableMemoryBytes,
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
		AvailableMemoryBytes: c.availableMemoryBytes,
	})
	return err
}
