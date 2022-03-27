package proxy

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/chainer"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// TODO - support these via gRPC call/CLI args
const (
	chainerInputTopic1  = "seldon.some-namespace.some-model-1.outputs"
	chainerOutputTopic1 = "seldon.some-namespace.some-model-2.inputs"
	chainerInputTopic2  = "seldon.some-namespace.some-model-3.outputs.tensor1"
	chainerOutputTopic2 = "seldon.some-namespace.some-model-4.inputs"
)

type ProxyChainer struct {
	chainer.UnimplementedChainerServer
	logger log.FieldLogger
}

func New(logger log.FieldLogger) *ProxyChainer {
	return &ProxyChainer{
		logger: logger.WithField("source", "ProxyChainer"),
	}
}

func (pc *ProxyChainer) Start(port uint) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		pc.logger.Errorf("unable to start gRPC chainer server on port %d", port)
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	chainer.RegisterChainerServer(grpcServer, pc)

	pc.logger.Infof("starting gRPC chainer server on port %d", port)
	return grpcServer.Serve(l)
}

func (pc *ProxyChainer) SubscribePipelineUpdates(
	request *chainer.PipelineSubscriptionRequest,
	subscription chainer.Chainer_SubscribePipelineUpdatesServer,
) error {
	logger := pc.logger.WithField("func", "SubscribePipelineUpdates")
	logger.Infof("received subscription from %s", request.GetName())

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		// TODO - support an actual stream of requests
		err := subscription.Send(
			&chainer.PipelineUpdateMessage{
				Op:       chainer.PipelineUpdateMessage_Create,
				Pipeline: "some-pipeline",
				Version:  1,
				Uid:      "1234",
				Updates: []*chainer.PipelineStepUpdate{
					&chainer.PipelineStepUpdate{
						Sources: []string{chainerInputTopic1},
						Sink:    chainerOutputTopic1,
						Ty:      chainer.PipelineStepUpdate_Inner,
					},
					&chainer.PipelineStepUpdate{
						Sources: []string{chainerInputTopic2},
						Sink:    chainerOutputTopic2,
						Ty:      chainer.PipelineStepUpdate_Inner,
					},
				},
			},
		)
		if err != nil {
			return err
		}
	}

	time.Sleep(30 * time.Second)
	return nil
}

func (pc *ProxyChainer) PipelineUpdateEvent(
	_ context.Context,
	msg *chainer.PipelineUpdateStatusMessage,
) (*chainer.PipelineUpdateStatusResponse, error) {
	logger := pc.logger.WithField("func", "PipelineUpdateEvent")
	logger.Infof(
		"received update for pipeline %s (op: %s -- succeeded: %t -- reason: %s)",
		msg.Update.Pipeline, msg.Update.Op, msg.Success, msg.Reason,
	)

	// TODO - handle these better
	return &chainer.PipelineUpdateStatusResponse{}, nil
}
