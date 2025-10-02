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
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	pmock "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/mocks"
)

type MockGrpcClient struct {
	pmock.MockSchedulerClient
	stream chan *pb.ModelUpdateStatusMessage
}

func (mgrpc *MockGrpcClient) ModelStatusEvent(ctx context.Context, msg *pb.ModelUpdateStatusMessage, opts ...grpc.CallOption) (*pb.ModelUpdateStatusResponse, error) {
	mgrpc.stream <- msg
	return &pb.ModelUpdateStatusResponse{}, nil
}

func (mgrpc *MockGrpcClient) Recv() (*pb.ModelUpdateStatusMessage, error) {
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

func newMockGrpcClient() *MockGrpcClient {
	return &MockGrpcClient{
		MockSchedulerClient: *pmock.NewMockSchedulerClient(nil),
		stream:              make(chan *pb.ModelUpdateStatusMessage, 100),
	}
}

type MockConsumerManager struct{}

func (mcm *MockConsumerManager) AddModel(modelName string) error {
	return nil
}

func (mcm *MockConsumerManager) RemoveModel(modelName string, cleanTopicsOnDeletion bool, keepTopics bool) error {
	return nil
}

func (mcm *MockConsumerManager) Exists(modelName string) bool {
	return false
}

func (mcm *MockConsumerManager) GetNumModels() int {
	return 0
}

func (mcm *MockConsumerManager) Healthy() error {
	return nil
}

func (mcm *MockConsumerManager) Stop() {}

func createClient() *KafkaSchedulerClient {
	logger := logrus.New()
	testLogger := logger.WithField("test", "PipelineSchedulerClientTest")
	return NewKafkaSchedulerClient(
		testLogger,
		&MockConsumerManager{},
		&tls.TLSOptions{},
	)
}

func TestCreateAndDeleteConfirmationMessages(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name    string
		message string
		event   *pb.ModelStatusResponse
	}

	tests := []test{
		{
			name:    "delete",
			message: "Model dummy-model removed",
			event: &pb.ModelStatusResponse{
				Operation: pb.ModelStatusResponse_ModelDelete,
				ModelName: "dummy-model",
				Versions: []*pb.ModelVersionStatus{
					&pb.ModelVersionStatus{
						Version: 1,
						State: &pb.ModelStatus{
							ModelGwState: pb.ModelStatus_ModelAvailable,
							Reason:       "",
						},
					},
				},
			},
		},
		{
			name:    "create",
			message: "Model dummy-model added",
			event: &pb.ModelStatusResponse{
				Operation: pb.ModelStatusResponse_ModelCreate,
				ModelName: "dummy-model",
				Versions: []*pb.ModelVersionStatus{
					&pb.ModelVersionStatus{
						Version: 1,
						State: &pb.ModelStatus{
							ModelGwState: pb.ModelStatus_ModelCreate,
							Reason:       "",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create pipeline-gw client
			client := createClient()

			// create mock grpc client
			grpcClient := newMockGrpcClient()

			processor := &EventProcessor{
				client:         client,
				grpcClient:     grpcClient,
				subscriberName: "dummy-subscriber",
				logger:         client.logger.WithField("component", "EventProcessor"),
			}

			processor.handleEvent(test.event)
			msg, err := grpcClient.Recv()
			g.Expect(err).To(BeNil())
			g.Expect(msg.Success).To(BeTrue())
			g.Expect(msg.Reason).To(Equal(test.message))
		})
	}
}
