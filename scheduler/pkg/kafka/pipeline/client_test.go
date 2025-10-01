/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	chainer "github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	pipeline "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline"
	pmock "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/mocks"
	smock "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/status/mocks"
)

type MockGrpcClient struct {
	pmock.MockSchedulerClient
	stream chan *chainer.PipelineUpdateStatusMessage
}

func (mgrpc *MockGrpcClient) PipelineStatusEvent(arg0 context.Context, arg1 *chainer.PipelineUpdateStatusMessage, arg2 ...grpc.CallOption) (*chainer.PipelineUpdateStatusResponse, error) {
	mgrpc.stream <- arg1
	return &chainer.PipelineUpdateStatusResponse{}, nil
}

func (mgrpc *MockGrpcClient) Recv() (*chainer.PipelineUpdateStatusMessage, error) {
	msg, ok := <-mgrpc.stream
	if !ok {
		return nil, context.Canceled
	}
	return msg, nil
}

func (mgrpc *MockGrpcClient) Close() error {
	close(mgrpc.stream)
	return nil
}

func newMockGrpcClient(ctrl *gomock.Controller) *MockGrpcClient {
	return &MockGrpcClient{
		MockSchedulerClient: *pmock.NewMockSchedulerClient(ctrl),
		stream:              make(chan *chainer.PipelineUpdateStatusMessage, 100),
	}
}

func createClient(
	ctrl *gomock.Controller,
	statusUpdater *smock.MockPipelineStatusUpdater,
	inferer *pmock.MockPipelineInferer,
) *pipeline.PipelineSchedulerClient {
	logger := logrus.New()
	testLogger := logger.WithField("test", "PipelineSchedulerClientTest")
	return pipeline.NewPipelineSchedulerClient(
		testLogger,
		statusUpdater,
		inferer,
		&tls.TLSOptions{},
	)
}

func TestCreateAndDeleteConfirmationMessages(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name    string
		message string
		event   *pb.PipelineStatusResponse
	}

	tests := []test{
		{
			name:    "delete",
			message: "Pipeline dummy-pipeline deleted",
			event: &pb.PipelineStatusResponse{
				Operation:    pb.PipelineStatusResponse_PipelineDelete,
				PipelineName: "dummy-pipeline",
				Versions: []*pb.PipelineWithState{
					&pb.PipelineWithState{
						Pipeline: &pb.Pipeline{
							Name:    "dummy-pipeline",
							Version: 1,
							Uid:     "x",
							Steps: []*pb.PipelineStep{
								{
									Name: "a",
								},
								{
									Name:   "b",
									Inputs: []string{"a.outputs"},
								},
							},
						},
						State: &pb.PipelineVersionState{
							PipelineVersion:     1,
							Status:              pb.PipelineVersionState_PipelineReady,
							Reason:              "",
							LastChangeTimestamp: &timestamppb.Timestamp{},
							ModelsReady:         true,
							PipelineGwStatus:    pb.PipelineVersionState_PipelineReady,
							PipelineGwReason:    "",
						},
					},
				},
			},
		},
		{
			name:    "create",
			message: "Pipeline dummy-pipeline loaded",
			event: &pb.PipelineStatusResponse{
				Operation:    pb.PipelineStatusResponse_PipelineCreate,
				PipelineName: "dummy-pipeline",
				Versions: []*pb.PipelineWithState{
					&pb.PipelineWithState{
						Pipeline: &pb.Pipeline{
							Name:    "dummy-pipeline",
							Version: 1,
							Uid:     "x",
							Steps: []*pb.PipelineStep{
								{
									Name: "a",
								},
								{
									Name:   "b",
									Inputs: []string{"a.outputs"},
								},
							},
						},
						State: &pb.PipelineVersionState{
							PipelineVersion:     1,
							Status:              pb.PipelineVersionState_PipelineCreate,
							Reason:              "",
							LastChangeTimestamp: &timestamppb.Timestamp{},
							ModelsReady:         true,
							PipelineGwStatus:    pb.PipelineVersionState_PipelineCreate,
							PipelineGwReason:    "",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)
			ctrl := gomock.NewController(t)
			mockStatusUpdater := smock.NewMockPipelineStatusUpdater(ctrl)
			mockInferer := pmock.NewMockPipelineInferer(ctrl)

			// Add these lines to set up expectations:
			mockInferer.EXPECT().
				DeletePipeline(gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()

			mockInferer.EXPECT().
				LoadOrStorePipeline(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()

			mockStatusUpdater.EXPECT().
				Update(gomock.Any()).
				AnyTimes()

			// create pipeline-gw client
			client := createClient(ctrl, mockStatusUpdater, mockInferer)

			// create mock grpc client
			grpcClient := newMockGrpcClient(ctrl)
			processor := pipeline.NewEventProcessor(
				client, grpcClient, "dummy-subscriber",
				logrus.New().WithField("test", "EventProcessorTest"),
			)

			processor.HandleEvent(test.event)
			msg, err := grpcClient.Recv()
			g.Expect(err).To(BeNil())
			g.Expect(msg.Success).To(BeTrue())
			g.Expect(msg.Reason).To(Equal(test.message))
		})
	}
}
