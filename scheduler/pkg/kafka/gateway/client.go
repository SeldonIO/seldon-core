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
	// note: retry is done in the caller
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transCreds),
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
	logger.Infof("Closing connection to scheduler")
	defer func() {
		_ = stream.CloseSend()
	}()
	return nil
}
