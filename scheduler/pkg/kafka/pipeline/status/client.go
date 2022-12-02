/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package status

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"

	"github.com/cenkalti/backoff/v4"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GrpcRetryBackoffMillisecs * time.Millisecond)),
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
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
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
