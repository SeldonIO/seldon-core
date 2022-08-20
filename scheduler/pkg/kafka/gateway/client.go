package gateway

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	seldontls "github.com/seldonio/seldon-core-v2/components/tls/pkg/tls"
	"k8s.io/client-go/kubernetes"

	"github.com/cenkalti/backoff/v4"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	SubscriberName        = "seldon-modelgateway"
	envSchedulerTLSPrefix = "SCHEDULER"
)

type KafkaSchedulerClient struct {
	logger           logrus.FieldLogger
	conn             *grpc.ClientConn
	callOptions      []grpc.CallOption
	consumerManager  *ConsumerManager
	certificateStore *seldontls.CertificateStore
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
	}
}

func (kc *KafkaSchedulerClient) ConnectToScheduler(host string, plainTxtPort int, tlsPort int, clientset kubernetes.Interface) error {
	logger := kc.logger.WithField("func", "ConnectToScheduler")
	var err error
	kc.certificateStore, err = seldontls.NewCertificateStore(envSchedulerTLSPrefix, clientset)
	if err != nil {
		return err
	}
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
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
	kc.logger.Info("Connected to scheduler")
	return nil
}

func (kc *KafkaSchedulerClient) Start() error {
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
	return nil
}

func (kc *KafkaSchedulerClient) SubscribeModelEvents() error {
	logger := kc.logger.WithField("func", "SubscribeModelEvents")
	grpcClient := scheduler.NewSchedulerClient(kc.conn)
	logger.Info("Subscribing to model status events")
	stream, err := grpcClient.SubscribeModelStatus(context.Background(), &scheduler.ModelSubscriptionRequest{SubscriberName: SubscriberName}, grpc_retry.WithMax(100))
	if err != nil {
		return err
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error(err, "event recv failed")
			return err
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
	return nil
}
