/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package status

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

const (
	SubscriberName = "seldon-pipelinegateway"
)

type PipelineSchedulerClient struct {
	logger                logrus.FieldLogger
	conn                  *grpc.ClientConn
	callOptions           []grpc.CallOption
	pipelineStatusUpdater PipelineStatusUpdater
	certificateStore      *seldontls.CertificateStore
	stop                  atomic.Bool
}

func NewPipelineSchedulerClient(logger logrus.FieldLogger, pipelineStatusUpdater PipelineStatusUpdater) *PipelineSchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &PipelineSchedulerClient{
		logger:                logger.WithField("source", "PipelineSchedulerClient"),
		callOptions:           opts,
		pipelineStatusUpdater: pipelineStatusUpdater,
	}
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
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixControlPlane)
	if protocol == seldontls.SecurityProtocolSSL {
		pc.certificateStore, err = seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixControlPlaneServer))
		if err != nil {
			return err
		}
	}
	var transCreds credentials.TransportCredentials
	var port int
	if pc.certificateStore == nil {
		logger.Info("Starting plaintxt client to scheduler")
		transCreds = insecure.NewCredentials()
		port = plainTxtPort
	} else {
		logger.Info("Starting TLS client to scheduler")
		transCreds = pc.certificateStore.CreateClientTransportCredentials()
		port = tlsPort
	}
	// note: retry is done in the caller
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
	}
	logger.Infof("Connecting to scheduler at %s:%d", host, port)
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
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
		backOffExp := backoff.NewExponentialBackOff()
		// Set some reasonable settings for trying to reconnect to scheduler
		backOffExp.MaxElapsedTime = 0 // Never stop due to large time between calls
		backOffExp.MaxInterval = time.Second * 15
		backOffExp.InitialInterval = time.Second
		err = backoff.RetryNotify(pc.SubscribePipelineEvents, backOffExp, logFailure)
		if err != nil {
			logger.WithError(err).Fatal("Failed to start pipeline gateway client")
			return err
		}
		logger.Info("Subscribe ended")
	}
}

func (pc *PipelineSchedulerClient) SubscribePipelineEvents() error {
	logger := pc.logger.WithField("func", "SubscribePipelineEvents")
	grpcClient := scheduler.NewSchedulerClient(pc.conn)
	logger.Info("Subscribing to pipeline status events")
	stream, errSub := grpcClient.SubscribePipelineStatus(context.Background(), &scheduler.PipelineSubscriptionRequest{SubscriberName: SubscriberName}, grpc_retry.WithMax(100))
	if errSub != nil {
		return errSub
	}
	for {
		if pc.stop.Load() {
			logger.Info("Stopping")
			break
		}
		event, err := stream.Recv()
		if err != nil {
			logger.WithError(err).Error("event recv failed")
			break
		}
		// The expected contract is just the latest version will be sent to us
		if len(event.Versions) != 1 {
			logger.Info("Expected a single model version", "numVersions", len(event.Versions), "name", event.GetPipelineName())
			continue
		}

		pv, err := pipeline.CreatePipelineVersionWithStateFromProto(event.Versions[0])
		if err != nil {
			logger.Warningf("Failed to create pipeline state for pipeline %s with %s", event.PipelineName, protojson.Format(event))
			continue
		}

		logger.Debugf("Processing pipeline %s version %d with state %s", pv.Name, pv.Version, pv.State.Status.String())

		pc.pipelineStatusUpdater.Update(pv)
	}
	logger.Infof("Closing connection to scheduler")
	defer func() {
		_ = stream.CloseSend()
	}()
	return nil
}
