/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/go-logr/logr"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

const (
	// these 2 constants in combination with the backoff exponential function will give us a max backoff of 13.5 minutes
	SchedulerConnectMaxRetries    = 12
	SchedulerConnectBackoffScalar = 200 * time.Millisecond
)

type SchedulerClient struct {
	client.Client
	logger           logr.Logger
	callOptions      []grpc.CallOption
	recorder         record.EventRecorder
	certificateStore *tls.CertificateStore
	seldonRuntimes   map[string]*grpc.ClientConn // map of namespace to grpc connection
	mu               sync.Mutex
}

//  connect on demand by add getConnection(namespace) which if not existing calls connect to scheduler.
// For this will need to know ports (hardwire for now to 9004 and 9044 - ssl comes fom envvar - so always
// the same for all schedulers

func NewSchedulerClient(logger logr.Logger, client client.Client, recorder record.EventRecorder) *SchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &SchedulerClient{
		Client:         client,
		logger:         logger.WithName("schedulerClient"),
		callOptions:    opts,
		recorder:       recorder,
		seldonRuntimes: make(map[string]*grpc.ClientConn),
	}
}

func getSchedulerHost(namespace string) string {
	return fmt.Sprintf("seldon-scheduler.%s", namespace)
}

// startEventHanders starts the grpc stream connections to the scheduler for the different resources we care about
// we also add a retry mechanism to reconnect if the connection is lost, this can happen if the scheduler is restarted
// or if the network connection is lost. We use an exponential backoff to retry the connection.
// note that when the scheduler is completely dead we will be not be able to reconnect and these go routines will retry forever
// TODO: add a max retry count and report back to the caller.
func (s *SchedulerClient) startEventHanders(namespace string, conn *grpc.ClientConn) {
	// Subscribe the event streams from scheduler
	go func() {
		for {
			err := retryFn(s.SubscribeModelEvents, conn, namespace, s.logger)
			if err != nil {
				s.logger.Error(err, "Subscribe ended for model events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for model events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFn(s.SubscribeServerEvents, conn, namespace, s.logger)
			if err != nil {
				s.logger.Error(err, "Subscribe ended for server events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for server events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFn(s.SubscribePipelineEvents, conn, namespace, s.logger)
			if err != nil {
				s.logger.Error(err, "Subscribe ended for pipeline events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for pipeline events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFn(s.SubscribeExperimentEvents, conn, namespace, s.logger)
			if err != nil {
				s.logger.Error(err, "Subscribe ended for experiment events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for experiment events", "namespace", namespace)
			}
		}
	}()
}

func (s *SchedulerClient) RemoveConnection(namespace string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if conn, ok := s.seldonRuntimes[namespace]; ok {
		delete(s.seldonRuntimes, namespace)
		err := conn.Close()
		if err != nil {
			s.logger.Error(err, "Failed to close grpc connection to scheduler", "namespace", namespace)
		}
	}
}

// A smoke test allows us to quickly check if we actually have a functional grpc connection to the scheduler
func (s *SchedulerClient) smokeTestConnection(conn *grpc.ClientConn) error {
	grcpClient := scheduler.NewSchedulerClient(conn)

	stream, err := grcpClient.SubscribeModelStatus(context.TODO(), &scheduler.ModelSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(1))
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) getConnection(namespace string) (*grpc.ClientConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if conn, ok := s.seldonRuntimes[namespace]; !ok {
		var err error
		conn, err = s.connectToScheduler(getSchedulerHost(namespace), namespace, 9004, 9044)
		if err != nil {
			return nil, err
		}
		err = s.smokeTestConnection(conn)
		if err != nil {
			s.logger.Info("Failed smoke test on scheduler", "namespace", namespace)
			return nil, err
		}
		s.startEventHanders(namespace, conn)
		s.seldonRuntimes[namespace] = conn
		return conn, nil
	} else {
		return conn, nil
	}
}

func (s *SchedulerClient) connectToScheduler(host string, namespace string, plainTxtPort int, tlsPort int) (*grpc.ClientConn, error) {
	var err error
	protocol := tls.GetSecurityProtocolFromEnv(tls.EnvSecurityPrefixControlPlane)
	s.logger.Info("connect to scheduler", "protocol", protocol)
	if protocol == tls.SecurityProtocolSSL {
		s.certificateStore, err = tls.NewCertificateStore(tls.Prefix(tls.EnvSecurityPrefixControlPlaneClient),
			tls.ValidationPrefix(tls.EnvSecurityPrefixControlPlaneServer),
			tls.Namespace(namespace))
		if err != nil {
			return nil, err
		}
	}
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{}

	var port int
	if s.certificateStore != nil {
		port = tlsPort
		opts = append(opts, grpc.WithTransportCredentials(s.certificateStore.CreateClientTransportCredentials()))
		s.logger.Info("Running scheduler client in TLS mode", "port", port)
	} else {
		port = plainTxtPort
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		s.logger.Info("Running scheduler client in plain text mode", "port", port)
	}
	opts = append(opts, grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)))
	opts = append(opts, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	s.logger.Info("Dialing scheduler", "host", host, "port", port)
	// Not using DialContext with context timeout and withBlocking as this seems to ignore errors such as TLS certificate
	// issues and not return any error resulting in uninformative context timeouts only.
	// See https://github.com/grpc/grpc-go/issues/622
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		s.logger.Error(err, "Failed to connect to scheduler")
		return nil, err
	}
	s.logger.Info("Connected to scheduler", "host", host, "port", port)
	return conn, nil
}

func (s *SchedulerClient) checkErrorRetryable(resource string, resourceName string, err error) bool {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			s.logger.Info(
				"Got grpc status code",
				"err", err.Error(),
				"code", st.Code(),
				"resource", resource,
				"resourceName", resourceName,
			)
			switch st.Code() {
			case codes.FailedPrecondition,
				codes.Unimplemented:
				s.logger.Info(
					"Non retryable error",
					"code", st.Code(),
					"resource", resource,
					"resourceName", resourceName,
				)
				return false
			default:
				s.logger.Info(
					"retryable error",
					"code", st.Code(),
					"resource", resource,
					"resourceName", resourceName,
				)
				return true
			}
		} else {
			s.logger.Info(
				"Got non grpc error",
				"error", err.Error(),
				"resource", resource,
				"resourceName", resourceName,
			)
			return true
		}
	} else {
		return false
	}
}

func retryFn(
	fn func(context context.Context, conn *grpc.ClientConn, namespace string) error,
	conn *grpc.ClientConn, namespace string, logger logr.Logger,
) error {
	logger.Info("RetryFn", "namespace", namespace)
	logFailure := func(err error, delay time.Duration) {
		logger.Error(err, "Scheduler not ready")
	}
	backOffExp := backoff.NewExponentialBackOff()
	fnWithArgs := func() error {
		return fn(context.Background(), conn, namespace)
	}
	err := backoff.RetryNotify(fnWithArgs, backOffExp, logFailure)
	if err != nil {
		logger.Error(err, "Failed to connect to scheduler", "namespace", namespace)
		return err
	}
	return nil
}
