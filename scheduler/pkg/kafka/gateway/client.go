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

package gateway

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/util"

	"github.com/cenkalti/backoff/v4"
	seldontls "github.com/seldonio/seldon-core-v2/components/tls/pkg/tls"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"sync/atomic"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	SubscriberName = "seldon-modelgateway"
)

type KafkaSchedulerClient struct {
	logger           logrus.FieldLogger
	conn             *grpc.ClientConn
	callOptions      []grpc.CallOption
	consumerManager  *ConsumerManager
	certificateStore *seldontls.CertificateStore
	stop             atomic.Bool
}

func NewKafkaSchedulerClient(logger logrus.FieldLogger, consumerManager *ConsumerManager) *KafkaSchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &KafkaSchedulerClient{
		logger:          logger.WithField("source", "KafkaSchedulerClient"),
		callOptions:     opts,
		consumerManager: consumerManager,
		stop:            atomic.Bool{},
	}
}

func (kc *KafkaSchedulerClient) ConnectToScheduler(host string, plainTxtPort int, tlsPort int) error {
	logger := kc.logger.WithField("func", "ConnectToScheduler")
	var err error
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixControlPlane)
	if protocol == seldontls.SecurityProtocolSSL {
		kc.certificateStore, err = seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneClient),
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
	if kc.certificateStore == nil {
		logger.Info("Starting plaintxt client to scheduler")
		transCreds = insecure.NewCredentials()
		port = plainTxtPort
	} else {
		logger.Info("Starting TLS client to scheduler")
		transCreds = kc.certificateStore.CreateClientTransportCredentials()
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
	kc.conn = conn
	logger.Info("Connected to scheduler")
	return nil
}

func (kc *KafkaSchedulerClient) Stop() {
	kc.stop.Store(true)
}

func (kc *KafkaSchedulerClient) Start() error {
	logger := kc.logger.WithField("func", "Start")
	// We keep trying to connect to scheduler.
	// If SubscribeModelEvents returns we try to start connecing again.
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
		backOffExp := backoff.NewExponentialBackOff()
		// Set some reasonable settings for trying to reconnect to scheduler
		backOffExp.MaxElapsedTime = 0 // Never stop due to large time between calls
		backOffExp.MaxInterval = time.Second * 15
		backOffExp.InitialInterval = time.Second
		err := backoff.RetryNotify(kc.SubscribeModelEvents, backOffExp, logFailure)
		if err != nil {
			kc.logger.WithError(err).Fatal("Failed to start modelgateway client")
			return err
		}
		logger.Info("Subscribe ended")
	}
}

func (kc *KafkaSchedulerClient) SubscribeModelEvents() error {
	logger := kc.logger.WithField("func", "SubscribeModelEvents")
	grpcClient := scheduler.NewSchedulerClient(kc.conn)
	logger.Info("Subscribing to model status events")
	stream, errSub := grpcClient.SubscribeModelStatus(context.Background(), &scheduler.ModelSubscriptionRequest{SubscriberName: SubscriberName}, grpc_retry.WithMax(100))
	if errSub != nil {
		return errSub
	}
	for {
		if kc.stop.Load() {
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
			logger.Info("Expected a single model version", "numVersions", len(event.Versions), "name", event.GetModelName())
			continue
		}
		latestVersionStatus := event.Versions[0]

		logger.Infof("Received event name %s version %d state %s", event.ModelName, latestVersionStatus.Version, latestVersionStatus.State.State.String())

		switch latestVersionStatus.State.State {
		case scheduler.ModelStatus_ModelAvailable:
			logger.Infof("Adding model %s", event.ModelName)
			err := kc.consumerManager.AddModel(event.ModelName)
			if err != nil {
				kc.logger.WithError(err).Errorf("Failed to add model %s", event.ModelName)
			}
		default:
			logger.Infof("Removing model %s", event.ModelName)
			err := kc.consumerManager.RemoveModel(event.ModelName)
			if err != nil {
				kc.logger.WithError(err).Errorf("Failed to remove model %s", event.ModelName)
			}
		}

	}

	defer func() {
		_ = stream.CloseSend()
	}()

	logger.Info("Exiting")
	return nil
}
