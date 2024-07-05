package scheduler

import (
	"context"
	"io"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"google.golang.org/grpc"
)

type mockSchedulerExperimentClient struct {
	counter int
	results []*scheduler.ExperimentStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_ExperimentStatusClient = (*mockSchedulerExperimentClient)(nil)

func newMockSchedulerExperimentClient(results []*scheduler.ExperimentStatusResponse) *mockSchedulerExperimentClient {
	return &mockSchedulerExperimentClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerExperimentClient) Recv() (*scheduler.ExperimentStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

type mockSchedulerClient struct {
	results []*scheduler.ExperimentStatusResponse
}

var _ scheduler.SchedulerClient = (*mockSchedulerClient)(nil)

func (s *mockSchedulerClient) ExperimentStatus(ctx context.Context, in *scheduler.ExperimentStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ExperimentStatusClient, error) {
	return newMockSchedulerExperimentClient(s.results), nil
}

// the below functions are not implemented
func (s *mockSchedulerClient) ServerNotify(ctx context.Context, in *scheduler.ServerNotifyRequest, opts ...grpc.CallOption) (*scheduler.ServerNotifyResponse, error) {
	return nil, nil
}

func (s *mockSchedulerClient) LoadModel(ctx context.Context, in *scheduler.LoadModelRequest, opts ...grpc.CallOption) (*scheduler.LoadModelResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) UnloadModel(ctx context.Context, in *scheduler.UnloadModelRequest, opts ...grpc.CallOption) (*scheduler.UnloadModelResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) LoadPipeline(ctx context.Context, in *scheduler.LoadPipelineRequest, opts ...grpc.CallOption) (*scheduler.LoadPipelineResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) UnloadPipeline(ctx context.Context, in *scheduler.UnloadPipelineRequest, opts ...grpc.CallOption) (*scheduler.UnloadPipelineResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) StartExperiment(ctx context.Context, in *scheduler.StartExperimentRequest, opts ...grpc.CallOption) (*scheduler.StartExperimentResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) StopExperiment(ctx context.Context, in *scheduler.StopExperimentRequest, opts ...grpc.CallOption) (*scheduler.StopExperimentResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) ServerStatus(ctx context.Context, in *scheduler.ServerStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ServerStatusClient, error) {
	return nil, nil

}
func (s *mockSchedulerClient) ModelStatus(ctx context.Context, in *scheduler.ModelStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ModelStatusClient, error) {
	return nil, nil
}
func (s *mockSchedulerClient) PipelineStatus(ctx context.Context, in *scheduler.PipelineStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_PipelineStatusClient, error) {
	return nil, nil
}
func (s *mockSchedulerClient) SchedulerStatus(ctx context.Context, in *scheduler.SchedulerStatusRequest, opts ...grpc.CallOption) (*scheduler.SchedulerStatusResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) SubscribeServerStatus(ctx context.Context, in *scheduler.ServerSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeServerStatusClient, error) {
	return nil, nil
}
func (s *mockSchedulerClient) SubscribeModelStatus(ctx context.Context, in *scheduler.ModelSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeModelStatusClient, error) {
	return nil, nil
}
func (s *mockSchedulerClient) SubscribeExperimentStatus(ctx context.Context, in *scheduler.ExperimentSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeExperimentStatusClient, error) {
	return nil, nil
}
func (s *mockSchedulerClient) SubscribePipelineStatus(ctx context.Context, in *scheduler.PipelineSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribePipelineStatusClient, error) {
	return nil, nil
}

func TestGetNumExperiments(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		results []*scheduler.ExperimentStatusResponse
	}

	tests := []test{
		{
			name: "experiment ok",
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName: "foo",
				},
				{
					ExperimentName: "bar",
				},
			},
		},
		{
			name:    "experiment ok",
			results: []*scheduler.ExperimentStatusResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mockSchedulerClient{results: test.results}
			num, err := getNumExperimentsFromScheduler(context.Background(), &client)

			g.Expect(err).To(BeNil())
			g.Expect(num).To(Equal(len(test.results)))

		})
	}
}
