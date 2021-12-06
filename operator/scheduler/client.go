package scheduler

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/go-logr/logr"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler/api"
	"google.golang.org/grpc"
)

type SchedulerClient struct {
	logger      logr.Logger
	conn        *grpc.ClientConn
	callOptions []grpc.CallOption
}

func NewSchedulerClient(logger logr.Logger) *SchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &SchedulerClient{
		logger:      logger.WithName("schedulerClient"),
		callOptions: opts,
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

func (s *SchedulerClient) LoadModel(ctx context.Context, model *v1alpha1.Model) error {
	logger := s.logger.WithName("LoadModel")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	md, err := model.AsModelDetails()
	if err != nil {
		return err
	}
	loadModelRequest := scheduler.LoadModelRequest{
		Model: md,
	}

	logger.Info("Load", "model name", model.Name)
	_, err = grcpClient.LoadModel(ctx, &loadModelRequest, grpc_retry.WithMax(2))
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) UnloadModel(ctx context.Context, model *v1alpha1.Model) error {
	logger := s.logger.WithName("UnloadModel")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	modelRef := &scheduler.ModelReference{
		Name: model.Name,
	}
	logger.Info("Unload", "model name", model.Name)
	_, err := grcpClient.UnloadModel(ctx, modelRef, grpc_retry.WithMax(2))
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) SubscribeEvents(ctx context.Context) error {
	logger := s.logger.WithName("SubscribeEvent")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	stream, err := grcpClient.SubscribeModelEvents(ctx, &scheduler.ModelSubscriptionRequest{Name: "seldon manager"}, grpc_retry.WithMax(10))
	if err != nil {
		return err
	}
	for {
		operation, err := stream.Recv()
		logger.Info("Received event %v", operation.Event)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch operation.Event {
		case scheduler.ModelEventMessage_REPLICAS_LOADED:
		}
	}
	return nil
}
