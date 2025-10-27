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
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-logr/logr"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/operator/v2/internal"
)

const (
	// schedulerConstantBackoff time to backoff between retries. We use a constant backoff as the k8s controller
	// reconciller will handle exponential backoff and there is a risk of using max retries with too high a value
	// and we end up backoff retrying forever.
	schedulerConstantBackoff = 2 * time.Second
	// schedulerConnectMaxRetries maximum amount of retries to attempt with failed gRPC requests to scheduler
	schedulerConnectMaxRetries = 2
)

const (
	// these keep alive settings need to match the scheduler counterpart in scheduler/pkg/util/constants.go
	clientKeepAliveTime    = 60 * time.Second
	clientKeepAliveTimeout = 2 * time.Second
	clientKeepAlivePermit  = false
	// backoff
	backoffMaxElapsedTime  = 0 // Never stop due to large time between calls
	backOffMaxInterval     = time.Second * 15
	backOffInitialInterval = time.Second
)

type SchedulerClient struct {
	client.Client
	logger                   logr.Logger
	callOptions              []grpc.CallOption
	recorder                 record.EventRecorder
	seldonRuntimes           map[string]*grpc.ClientConn // map of namespace to grpc connection
	tlsOptions               tls.TLSOptions
	useDeploymentsForServers bool
	mu                       sync.Mutex
}

//	connect on demand by add getConnection(namespace) which if not existing calls connect to scheduler.
//
// For this will need to know ports (hardwire for now to 9004 and 9044 - ssl comes fom envvar - so always
// the same for all schedulers
func NewSchedulerClient(logger logr.Logger, client client.Client, recorder record.EventRecorder, tlsOptions tls.TLSOptions, useDeploymentsForServers bool) *SchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &SchedulerClient{
		Client:                   client,
		logger:                   logger.WithName("SchedulerClient"),
		callOptions:              opts,
		recorder:                 recorder,
		seldonRuntimes:           make(map[string]*grpc.ClientConn),
		tlsOptions:               tlsOptions,
		useDeploymentsForServers: useDeploymentsForServers,
	}
}

func getSchedulerHost(namespace string) string {
	return fmt.Sprintf("seldon-scheduler.%s", namespace)
}

// startEventHanders starts the grpc stream connections to the scheduler for the different resources we care about
// we also add a retry mechanism to reconnect if the connection is lost, this can happen if the scheduler is restarted
// or if the network connection is lost. We use an exponential backoff to retry the connection.
// note that when the scheduler is completely dead we will be not be able to reconnect and these go routines will retry forever
// on reconnect we send the state of the different resources to the scheduler, this is to make sure that the scheduler has the correct state
// TODO: add a max retry count and report back to the caller.
// TODO add done for graceful shutdown otherwise these go routines will run forever
// TODO tidy up ctx from the different handlers, currently they are all context.Background()
func (s *SchedulerClient) startEventHanders(namespace string, conn *grpc.ClientConn) {
	s.logger.Info("Starting event handling", "namespace", namespace)

	// Subscribe the event streams from scheduler
	go func() {
		for {
			err := retryFnExpBackoff(s.SubscribeModelEvents, conn, namespace, s.logger.WithName("SubscribeModelEvents"))
			if err != nil {
				s.logger.Error(err, "Subscribe ended for model events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for model events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFnExpBackoff(s.SubscribeServerEvents, conn, namespace, s.logger.WithName("SubscribeServerEvents"))
			if err != nil {
				s.logger.Error(err, "Subscribe ended for server events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for server events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFnExpBackoff(s.SubscribePipelineEvents, conn, namespace, s.logger.WithName("SubscribePipelineEvents"))
			if err != nil {
				s.logger.Error(err, "Subscribe ended for pipeline events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for pipeline events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFnExpBackoff(s.SubscribeExperimentEvents, conn, namespace, s.logger.WithName("SubscribeExperimentEvents"))
			if err != nil {
				s.logger.Error(err, "Subscribe ended for experiment events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for experiment events", "namespace", namespace)
			}
		}
	}()
	go func() {
		for {
			err := retryFnExpBackoff(s.SubscribeControlPlaneEvents, conn, namespace, s.logger.WithName("SubscribeControlPlaneEvents"))
			if err != nil {
				s.logger.Error(err, "Subscribe ended for control plane events", "namespace", namespace)
			} else {
				s.logger.Info("Subscribe ended for control plane events", "namespace", namespace)
			}
		}
	}()
}

func (s *SchedulerClient) handleControlPlaneEvent(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string, operation scheduler.ControlPlaneResponse_Event) error {
	switch operation {
	case scheduler.ControlPlaneResponse_SEND_SERVERS:
		// on new reconnects we send a list of servers to the schedule
		err := s.handleRegisteredServers(ctx, grpcClient, namespace)
		if err != nil {
			if errors.Is(err, internal.ErrScalingSpec) {
				s.logger.Error(err, "All servers have scaling spec errors, CRs needs correcting")
				// we don't retry scaling spec errors as they can't be fixed by retrying, customer has to manually fix,
				// which will trigger a reconcile and notify scheduler
				return nil
			}
			s.logger.Error(err, "Failed to send registered server to scheduler")
			return err
		}
		return nil
	case scheduler.ControlPlaneResponse_SEND_EXPERIMENTS:
		err := s.handleExperiments(ctx, grpcClient, namespace)
		if err != nil {
			return fmt.Errorf("failed sending experiments: %w", err)
		}
		return nil
	case scheduler.ControlPlaneResponse_SEND_PIPELINES:
		err := s.handlePipelines(ctx, grpcClient, namespace)
		if err != nil {
			return fmt.Errorf("failed sending pipelines: %w", err)
		}
		return nil
	case scheduler.ControlPlaneResponse_SEND_MODELS:
		err := s.handleModels(ctx, grpcClient, namespace)
		if err != nil {
			return fmt.Errorf("failed sending models: %w", err)
		}
		return nil
	default:
		s.logger.Error(errors.New("unknown operation"), "Unsupported operation", "operation", operation)
		// we don't want to retry unknow operations
		return nil
	}
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
	grpcClient := scheduler.NewSchedulerClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := grpcClient.SubscribeModelStatus(ctx, &scheduler.ModelSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(1))
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

	if s.tlsOptions.Cert != nil {
		s.logger.Info("Connecting to scheduler", "protocol", "tls")
	} else {
		s.logger.Info("Connecting to scheduler", "protocol", "plaintext")
	}

	kacp := keepalive.ClientParameters{
		Time:                clientKeepAliveTime,
		Timeout:             clientKeepAliveTimeout,
		PermitWithoutStream: clientKeepAlivePermit,
	}

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{}

	var port int
	if s.tlsOptions.Cert != nil {
		port = tlsPort
		opts = append(opts, grpc.WithTransportCredentials(s.tlsOptions.Cert.CreateClientTransportCredentials()))
		s.logger.Info("Running scheduler client in TLS mode", "port", port)
	} else {
		port = plainTxtPort
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		s.logger.Info("Running scheduler client in plain text mode", "port", port)
	}
	// we dont have backoff retry on the grpc streams as we handle this in the event handlers
	opts = append(opts, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	opts = append(opts, grpc.WithKeepaliveParams(kacp))
	s.logger.Info("Dialing scheduler", "host", host, "port", port)
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", host, port), opts...)
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
