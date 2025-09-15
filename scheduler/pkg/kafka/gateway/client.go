/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	SubscriberName   = "seldon-modelgateway"
	SubscriberEnvVar = "POD_NAME"
)

type KafkaSchedulerClient struct {
	logger          logrus.FieldLogger
	conn            *grpc.ClientConn
	callOptions     []grpc.CallOption
	consumerManager *ConsumerManager
	stop            atomic.Bool
	ready           atomic.Bool
	subscriberName  string
	tlsOptions      *seldontls.TLSOptions
}

func NewKafkaSchedulerClient(logger logrus.FieldLogger, consumerManager *ConsumerManager, tlsOptions *seldontls.TLSOptions) *KafkaSchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	subscriberName := os.Getenv(SubscriberEnvVar)
	if subscriberName == "" {
		subscriberName = SubscriberName
	}

	return &KafkaSchedulerClient{
		logger:          logger.WithField("source", "KafkaSchedulerClient"),
		callOptions:     opts,
		consumerManager: consumerManager,
		stop:            atomic.Bool{},
		subscriberName:  subscriberName,
		tlsOptions:      tlsOptions,
	}
}

func (kc *KafkaSchedulerClient) ConnectToScheduler(host string, plainTxtPort int, tlsPort int) error {
	logger := kc.logger.WithField("func", "ConnectToScheduler")

	var transCreds credentials.TransportCredentials
	var port int
	if kc.tlsOptions.Cert == nil {
		logger.Info("Starting plaintxt client to scheduler")
		transCreds = insecure.NewCredentials()
		port = plainTxtPort
	} else {
		logger.Info("Starting TLS client to scheduler")
		transCreds = kc.tlsOptions.Cert.CreateClientTransportCredentials()
		port = tlsPort
	}

	kacp := util.GetClientKeepAliveParameters()

	// note: retry is done in the caller
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
		grpc.WithKeepaliveParams(kacp),
	}
	logger.Infof("Connecting to scheduler at %s:%d", host, port)
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return err
	}
	kc.conn = conn
	return nil
}

func (kc *KafkaSchedulerClient) Stop() {
	kc.stop.Store(true)
	if kc.conn != nil {
		_ = kc.conn.Close()
	}
}

func (kc *KafkaSchedulerClient) Start() error {
	logger := kc.logger.WithField("func", "Start")
	// We keep trying to connect to scheduler.
	// If SubscribeModelEvents returns we try to start connecting again.
	// Only stop if asked to via the keepRunning flag.
	// We allow the subscribeModelEvents to return nil on EOF to allow a new Exponential backoff to be started
	// rather than return an error and continue the current Exponential backoff with might have reached large intervals
	for {
		if kc.stop.Load() {
			logger.Info("Stopping")
			return nil
		}
		logFailure := func(err error, delay time.Duration) {
			kc.logger.WithError(err).Errorf("Scheduler not ready")
		}
		backOffExp := util.GetClientExponentialBackoff()
		err := backoff.RetryNotify(kc.SubscribeModelEvents, backOffExp, logFailure)
		if err != nil {
			kc.logger.WithError(err).Fatal("Failed to start modelgateway client")
			return err
		}
		logger.Info("Subscribe ended")
	}
}

func (kc *KafkaSchedulerClient) IsConnected() bool {
	return kc.ready.Load()
}

func (kc *KafkaSchedulerClient) setupSubscription() (*EventProcessor, scheduler.Scheduler_SubscribeModelStatusClient, error) {
	grpcClient := scheduler.NewSchedulerClient(kc.conn)
	kc.logger.Info("Subscribing to model status events")
	stream, err := grpcClient.SubscribeModelStatus(
		context.Background(),
		&scheduler.ModelSubscriptionRequest{SubscriberName: kc.subscriberName, IsModelGateway: true},
		grpc_retry.WithMax(util.MaxGRPCRetriesOnStream),
	)
	if err != nil {
		return nil, nil, err
	}

	processor := &EventProcessor{
		client:         kc,
		grpcClient:     grpcClient,
		subscriberName: kc.subscriberName,
		logger:         kc.logger.WithField("source", "EventProcessor"),
	}
	kc.ready.Store(true)
	return processor, stream, nil
}

func (kc *KafkaSchedulerClient) cleanup(stream scheduler.Scheduler_SubscribeModelStatusClient) {
	kc.logger.Infof("Closing connection to scheduler")
	kc.ready.Store(false)
	if stream != nil {
		_ = stream.CloseSend()
	}
}

func (kc *KafkaSchedulerClient) processEventsStream(
	stream scheduler.Scheduler_SubscribeModelStatusClient, processor *EventProcessor, logger *logrus.Entry,
) error {
	for {
		if kc.stop.Load() {
			kc.logger.Info("Stopping")
			return nil
		}

		event, err := stream.Recv()
		if err != nil {
			logger.WithError(err).Error("event recv failed")
			return err
		}

		processor.handleEvent(event)
	}
}

func (kc *KafkaSchedulerClient) SubscribeModelEvents() error {
	logger := kc.logger.WithField("func", "SubscribeModelEvents")

	processor, stream, err := kc.setupSubscription()
	if err != nil {
		return err
	}

	defer kc.cleanup(stream)
	return kc.processEventsStream(stream, processor, logger)
}

type EventProcessor struct {
	client         *KafkaSchedulerClient
	grpcClient     scheduler.SchedulerClient
	subscriberName string
	logger         *logrus.Entry
}

func (ep *EventProcessor) handleEvent(event *scheduler.ModelStatusResponse) {
	// The expected contract is just the latest version will be sent to us
	if len(event.Versions) != 1 {
		ep.logger.Info("Expected a single model version", "numVersions", len(event.Versions), "name", event.GetModelName())
		return
	}

	// get latest version status
	versionStatus := event.Versions[0]
	ep.logger.Infof("Received event name %s version %d state %s", event.ModelName, versionStatus.Version, versionStatus.State.State.String())

	// if the model is in a failed state and the consumer exists then we skip the removal
	// this is to prevent the consumer from being removed during transient failures of the control plane
	// in this way data plane can potentially continue to serve requests
	if versionStatus.GetState().GetState() == scheduler.ModelStatus_ScheduleFailed || versionStatus.GetState().GetState() == scheduler.ModelStatus_ModelProgressing {
		if ep.client.consumerManager.Exists(event.ModelName) {
			ep.logger.Warnf("Model %s schedule failed / progressing and consumer exists, skipping from removal", event.ModelName)
			return
		}
	}

	switch versionStatus.GetState().GetAvailableReplicas() {
	case 0:
		ep.handleDeleteModel(event, versionStatus)
	default:
		ep.handleCreateModel(event, versionStatus)
	}

}

func (ep *EventProcessor) handleCreateModel(event *scheduler.ModelStatusResponse, versionStatus *scheduler.ModelVersionStatus) {
	ep.logger.Infof("Removing model %s", event.ModelName)
	keepTopics := event.GetKeepTopics()
	cleanTopicsOnDeletion := versionStatus.GetModelDefn().GetDataflowSpec().GetCleanTopicsOnDelete()
	err := ep.client.consumerManager.RemoveModel(event.ModelName, cleanTopicsOnDeletion, keepTopics)
	if err != nil {
		ep.logger.WithError(err).Errorf("Failed to remove model %s", event.ModelName)
	}
}

func (ep *EventProcessor) handleDeleteModel(event *scheduler.ModelStatusResponse, versionStatus *scheduler.ModelVersionStatus) {
	// if there are available replicas then we add the consumer for the model
	// note that this will also get triggered if the model is already added but there is a status change (e.g. due to scale up)
	// and in the case then it is a no-op
	// note in the future we might want to check that available replicas > min replicas
	if versionStatus.GetState().GetState() != scheduler.ModelStatus_ModelAvailable {
		ep.logger.Warnf("Model %s state is: %s", event.ModelName, versionStatus.GetState().GetState().String())
	}

	if ep.client.consumerManager.Exists(event.ModelName) {
		ep.logger.Debugf("Model consumer %s already exists", event.ModelName)
		return
	}

	ep.logger.Infof("Adding model %s", event.ModelName)
	err := ep.client.consumerManager.AddModel(event.ModelName)
	if err != nil {
		ep.client.logger.WithError(err).Errorf("Failed to add model %s", event.ModelName)
	}
}

func (kc *KafkaSchedulerClient) sendModelStatusEvent(
	grpcClient scheduler.SchedulerClient,
	op scheduler.ModelUpdateMessage_ModelOperation,
	success bool,
	reason string,
	timestamp uint64,
) {
	_, err := grpcClient.ModelStatusEvent(
		context.Background(),
		&scheduler.ModelUpdateStatusMessage{
			Update: &scheduler.ModelUpdateMessage{
				Model:     "model",
				Version:   0,
				Uid:       "model-uid",
				Timestamp: timestamp,
				Stream:    "stream",
			},
			Success: success,
			Reason:  reason,
		},
	)
	if err != nil {
		kc.logger.WithError(err).Errorf("Failed to send model status event %s", op.String())
	}
}
