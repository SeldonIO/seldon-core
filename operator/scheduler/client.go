package scheduler

import (
	"fmt"
	"math"
	"time"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
)

type SchedulerClient struct {
	client.Client
	logger      logr.Logger
	conn        *grpc.ClientConn
	callOptions []grpc.CallOption
	recorder    record.EventRecorder
}

func NewSchedulerClient(logger logr.Logger, client client.Client, recorder record.EventRecorder) *SchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &SchedulerClient{
		Client:      client,
		logger:      logger.WithName("schedulerClient"),
		callOptions: opts,
		recorder:    recorder,
	}
}

func (s *SchedulerClient) ConnectToScheduler(host string, port int) error {
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
		return err
	}
	s.conn = conn
	s.logger.Info("Connected to scheduler")
	return nil
}
