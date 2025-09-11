/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

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
	"google.golang.org/protobuf/encoding/protojson"

	chainer "github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/status"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	SubscriberName = "seldon-pipelinegateway"
)

type PipelineSchedulerClient struct {
	logger                logrus.FieldLogger
	conn                  *grpc.ClientConn
	callOptions           []grpc.CallOption
	pipelineStatusUpdater status.PipelineStatusUpdater
	pipelineInferer       PipelineInferer
	stop                  atomic.Bool
	ready                 atomic.Bool
	tlsOptions            *tls.TLSOptions
}

func NewPipelineSchedulerClient(logger logrus.FieldLogger, pipelineStatusUpdater status.PipelineStatusUpdater, pipelineInferer PipelineInferer, tlsOptions *tls.TLSOptions) *PipelineSchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &PipelineSchedulerClient{
		logger:                logger.WithField("source", "PipelineSchedulerClient"),
		callOptions:           opts,
		pipelineStatusUpdater: pipelineStatusUpdater,
		pipelineInferer:       pipelineInferer,
		tlsOptions:            tlsOptions,
	}
}

func (pc *PipelineSchedulerClient) IsConnected() bool {
	return pc.ready.Load()
}

func (pc *PipelineSchedulerClient) connectToScheduler(host string, plainTxtPort int, tlsPort int) error {
	logger := pc.logger.WithField("func", "ConnectToScheduler")
	var err error
	if pc.conn != nil {
		err = pc.conn.Close()
		if err != nil {
			logger.WithError(err).Error("Failed to close previous grpc connection to scheduler")
		}
	}

	var transCreds credentials.TransportCredentials
	var port int
	if pc.tlsOptions.Cert == nil {
		logger.Info("Starting plaintxt client to scheduler")
		transCreds = insecure.NewCredentials()
		port = plainTxtPort
	} else {
		logger.Info("Starting TLS client to scheduler")
		transCreds = pc.tlsOptions.Cert.CreateClientTransportCredentials()
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
	pc.conn = conn
	return nil
}

func (pc *PipelineSchedulerClient) Stop() {
	pc.stop.Store(true)
	if pc.conn != nil {
		_ = pc.conn.Close()
	}
}

func (pc *PipelineSchedulerClient) Start(host string, plainTxtPort int, tlsPort int) error {
	logger := pc.logger.WithField("func", "Start")
	for {
		if pc.stop.Load() {
			logger.Info("Stopping")
			return nil
		}
		err := pc.connectToScheduler(host, plainTxtPort, tlsPort)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to connect to scheduler")
		}
		logger := pc.logger.WithField("func", "Start")
		logFailure := func(err error, delay time.Duration) {
			logger.WithError(err).Errorf("Scheduler not ready")
		}
		backOffExp := util.GetClientExponentialBackoff()
		err = backoff.RetryNotify(pc.SubscribePipelineEvents, backOffExp, logFailure)
		if err != nil {
			logger.WithError(err).Fatal("Failed to start pipeline gateway client")
			return err
		}
		logger.Info("Subscribe ended")
	}
}

func getSubscriberName() string {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		return SubscriberName
	}
	return podName
}

func getSubsriberIp() (string, error) {
	podIp := os.Getenv("POD_IP")
	if podIp == "" {
		return "", fmt.Errorf("POD_IP environment variable is not set")
	}
	return podIp, nil
}

func (pc *PipelineSchedulerClient) SubscribePipelineEvents() error {
	logger := pc.logger.WithField("func", "SubscribePipelineEvents")

	stream, processor, err := pc.setupSubscription(logger)
	if err != nil {
		return err
	}

	defer pc.cleanup(stream)
	return pc.processEventStream(stream, processor, logger)
}

func (pc *PipelineSchedulerClient) setupSubscription(logger *logrus.Entry) (scheduler.Scheduler_SubscribePipelineStatusClient, *EventProcessor, error) {
	grpcClient := scheduler.NewSchedulerClient(pc.conn)

	subscriberName := getSubscriberName()
	subscriberIp, err := getSubsriberIp()
	if err != nil {
		return nil, nil, err
	}

	logger.Infof("Subscriber (%s, %s) subscribing to pipeline status events", subscriberName, subscriberIp)
	stream, err := grpcClient.SubscribePipelineStatus(
		context.Background(),
		&scheduler.PipelineSubscriptionRequest{
			SubscriberName:    subscriberName,
			SubscriberIp:      subscriberIp,
			IsPipelineGateway: true,
		},
		grpc_retry.WithMax(util.MaxGRPCRetriesOnStream),
	)

	if err != nil {
		return nil, nil, err
	}

	processor := &EventProcessor{
		client:         pc,
		grpcClient:     grpcClient,
		subscriberName: subscriberName,
		logger:         logger,
	}

	pc.ready.Store(true)
	return stream, processor, nil
}

func (pc *PipelineSchedulerClient) processEventStream(stream scheduler.Scheduler_SubscribePipelineStatusClient, processor *EventProcessor, logger *logrus.Entry) error {
	for {
		if pc.stop.Load() {
			logger.Info("Stopping")
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

func (pc *PipelineSchedulerClient) cleanup(stream scheduler.Scheduler_SubscribePipelineStatusClient) {
	pc.ready.Store(false)
	pc.logger.Info("Closing connection to scheduler")
	if stream != nil {
		_ = stream.CloseSend()
	}
}

func (ep *EventProcessor) handleEvent(event *scheduler.PipelineStatusResponse) {
	switch len(event.Versions) {
	case 0:
		ep.handleDeletePipeline(event)
	case 1:
		ep.handleCreateOrUpdatePipeline(event)
	default:
		ep.handleInvalidVersionCount(event)
	}
}

type EventProcessor struct {
	client         *PipelineSchedulerClient
	grpcClient     scheduler.SchedulerClient
	subscriberName string
	logger         *logrus.Entry
}

func (ep *EventProcessor) handleDeletePipeline(event *scheduler.PipelineStatusResponse) {
	psm := ep.client.pipelineStatusUpdater.(*status.PipelineStatusManager)
	pv := psm.Get(event.PipelineName)
	if pv == nil {
		ep.reportFailure(
			chainer.PipelineUpdateMessage_Delete,
			nil,
			fmt.Sprintf("No existing pipeline %s to delete", event.PipelineName),
			event.Timestamp,
			nil,
		)
		return
	}

	err := ep.client.pipelineInferer.DeletePipeline(event.PipelineName, false)
	if err != nil {
		ep.reportFailure(
			chainer.PipelineUpdateMessage_Delete,
			pv,
			fmt.Sprintf("Failed to delete pipeline %s", event.PipelineName),
			event.Timestamp,
			err,
		)
		return
	}

	message := fmt.Sprintf("Pipeline %s deleted", event.PipelineName)
	ep.reportSuccess(chainer.PipelineUpdateMessage_Delete, pv, message, event.Timestamp)
}

func (ep *EventProcessor) handleCreateOrUpdatePipeline(event *scheduler.PipelineStatusResponse) {
	pv, err := pipeline.CreatePipelineVersionWithStateFromProto(event.Versions[0])
	if err != nil {
		ep.reportFailure(
			chainer.PipelineUpdateMessage_Create,
			nil,
			fmt.Sprintf("Failed to create pipeline version for pipeline %s with %s", event.PipelineName, protojson.Format(event)),
			event.Timestamp,
			err,
		)
		return
	}

	ep.logger.Debugf("Processing pipeline %s version %d with state %s", pv.Name, pv.Version, pv.State.Status.String())
	ep.client.pipelineStatusUpdater.Update(pv)

	_, err = ep.client.pipelineInferer.LoadOrStorePipeline(pv.Name, false, false)
	if err != nil {
		ep.reportFailure(
			chainer.PipelineUpdateMessage_Create,
			pv,
			fmt.Sprintf("Failed to load/store pipeline %s", pv.Name),
			event.Timestamp,
			err,
		)
		return
	}

	message := fmt.Sprintf("Pipeline %s loaded", event.PipelineName)
	ep.reportSuccess(chainer.PipelineUpdateMessage_Create, pv, message, event.Timestamp)
}

func (ep *EventProcessor) handleInvalidVersionCount(event *scheduler.PipelineStatusResponse) {
	message := fmt.Sprint("Expected at most a single model version", "numVersions", len(event.Versions), "name", event.GetPipelineName())
	ep.reportFailure(
		chainer.PipelineUpdateMessage_Create,
		nil,
		message,
		event.Timestamp,
		fmt.Errorf("invalid version count"),
	)
}

func (ep *EventProcessor) reportSuccess(op chainer.PipelineUpdateMessage_PipelineOperation, pv *pipeline.PipelineVersion, message string, timestamp uint64) {
	ep.logger.Info(message)
	ep.sendPipelineStatusEvent(op, pv, true, message, timestamp)
}

func (ep *EventProcessor) reportFailure(op chainer.PipelineUpdateMessage_PipelineOperation, pv *pipeline.PipelineVersion, message string, timestamp uint64, err error) {
	if err != nil {
		ep.logger.WithError(err).Error(message)
	} else {
		ep.logger.Error(message)
	}
	ep.sendPipelineStatusEvent(op, pv, false, message, timestamp)
}

func (ep *EventProcessor) sendPipelineStatusEvent(
	op chainer.PipelineUpdateMessage_PipelineOperation,
	pv *pipeline.PipelineVersion,
	success bool,
	reason string,
	timestamp uint64,
) {
	_, err := ep.grpcClient.PipelineStatusEvent(
		context.Background(),
		&chainer.PipelineUpdateStatusMessage{
			Update: &chainer.PipelineUpdateMessage{
				Op:        op,
				Pipeline:  pv.Name,
				Version:   pv.Version,
				Uid:       pv.UID,
				Stream:    ep.subscriberName,
				Timestamp: timestamp,
			},
			Success: success,
			Reason:  reason,
		},
	)
	if err != nil {
		ep.logger.WithError(err).Errorf("Failed to send pipeline status event for pipeline %s", pv.Name)
	}
}
