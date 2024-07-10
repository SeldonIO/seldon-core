package scheduler

import (
	"context"
	"io"
	"testing"

	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockSchedulerExperimentGrpcClient struct {
	counter int
	results []*scheduler.ExperimentStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_ExperimentStatusClient = (*mockSchedulerExperimentGrpcClient)(nil)

func newMockSchedulerExperimentGrpcClient(results []*scheduler.ExperimentStatusResponse) *mockSchedulerExperimentGrpcClient {
	return &mockSchedulerExperimentGrpcClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerExperimentGrpcClient) Recv() (*scheduler.ExperimentStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

type mockSchedulerPipelineGrpcClient struct {
	counter int
	results []*scheduler.PipelineStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_PipelineStatusClient = (*mockSchedulerPipelineGrpcClient)(nil)

func newMockSchedulerPipelineGrpcClient(results []*scheduler.PipelineStatusResponse) *mockSchedulerPipelineGrpcClient {
	return &mockSchedulerPipelineGrpcClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerPipelineGrpcClient) Recv() (*scheduler.PipelineStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

type mockSchedulerClient struct {
	responses_experiments []*scheduler.ExperimentStatusResponse
	responses_pipelines   []*scheduler.PipelineStatusResponse
	requests_experiments  []*scheduler.StartExperimentRequest
	requests_pipelines    []*scheduler.LoadPipelineRequest
	requests_models       []*scheduler.LoadModelRequest
	errors                map[string]error
}

var _ scheduler.SchedulerClient = (*mockSchedulerClient)(nil)

func (s *mockSchedulerClient) ExperimentStatus(ctx context.Context, in *scheduler.ExperimentStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ExperimentStatusClient, error) {
	return newMockSchedulerExperimentGrpcClient(s.responses_experiments), nil
}

func (s *mockSchedulerClient) ServerNotify(ctx context.Context, in *scheduler.ServerNotifyRequest, opts ...grpc.CallOption) (*scheduler.ServerNotifyResponse, error) {
	return nil, nil
}

func (s *mockSchedulerClient) LoadModel(ctx context.Context, in *scheduler.LoadModelRequest, opts ...grpc.CallOption) (*scheduler.LoadModelResponse, error) {
	s.requests_models = append(s.requests_models, in)
	return nil, nil
}
func (s *mockSchedulerClient) UnloadModel(ctx context.Context, in *scheduler.UnloadModelRequest, opts ...grpc.CallOption) (*scheduler.UnloadModelResponse, error) {
	err, ok := s.errors["UnloadModel"]
	if ok {
		return nil, err
	} else {
		return nil, nil
	}
}
func (s *mockSchedulerClient) LoadPipeline(ctx context.Context, in *scheduler.LoadPipelineRequest, opts ...grpc.CallOption) (*scheduler.LoadPipelineResponse, error) {
	s.requests_pipelines = append(s.requests_pipelines, in)
	return nil, nil
}
func (s *mockSchedulerClient) UnloadPipeline(ctx context.Context, in *scheduler.UnloadPipelineRequest, opts ...grpc.CallOption) (*scheduler.UnloadPipelineResponse, error) {
	return nil, nil
}
func (s *mockSchedulerClient) StartExperiment(ctx context.Context, in *scheduler.StartExperimentRequest, opts ...grpc.CallOption) (*scheduler.StartExperimentResponse, error) {
	s.requests_experiments = append(s.requests_experiments, in)
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
	return newMockSchedulerPipelineGrpcClient(s.responses_pipelines), nil
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

// new mockSchedulerClient (not grpc)
func newMockSchedulerClient(objs ...client.Object) *SchedulerClient {
	logger := zap.New()
	fakeRecorder := record.NewFakeRecorder(3)
	scheme := runtime.NewScheme()
	_ = mlopsv1alpha1.AddToScheme(scheme)
	fakeClient := testing2.NewFakeClient(scheme, objs...)
	return NewSchedulerClient(
		logger,
		fakeClient,
		fakeRecorder,
	)
}

func TestHandleLoadedExperiments(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "with experiments",
			resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
		},
		{
			name: "with deleted experiments",
			resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bar",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ExperimentFinalizerName},
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
		},
		{
			name:      "no experiments",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerClient{}
			client := newMockSchedulerClient(test.resources...)
			handleLoadedExperiments(context.Background(), "", client, &grpcClient)
			activeResources := 0
			// TODO check the entire object
			for idx, req := range test.resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(grpcClient.requests_experiments[idx].Experiment.GetName()))
					activeResources++
				}
			}
			g.Expect(len(grpcClient.requests_experiments)).To(Equal(activeResources))
		})
	}
}

func TestHandleLoadedModels(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "with models",
			resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
			},
		},
		{
			name: "with deleted models",
			resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bar",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ExperimentFinalizerName},
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
			},
		},
		{
			name:      "no models",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerClient{}
			client := newMockSchedulerClient(test.resources...)
			handleLoadedModels(context.Background(), "", client, &grpcClient)
			activeResources := 0
			// TODO check the entire object
			for idx, req := range test.resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(grpcClient.requests_models[idx].Model.Meta.GetName()))
					activeResources++
				}
			}
			g.Expect(len(grpcClient.requests_models)).To(Equal(activeResources))
		})
	}
}

func TestHandleLoadedPipelines(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "with pipelines",
			resources: []client.Object{
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
			},
		},
		{
			name: "with deleted pipelines",
			resources: []client.Object{
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bar",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.PipelineFinalizerName},
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
			},
		},
		{
			name:      "no pipelines",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerClient{}
			client := newMockSchedulerClient(test.resources...)
			handleLoadedPipelines(context.Background(), "", client, &grpcClient)
			activeResources := 0
			// TODO check the entire object
			for idx, req := range test.resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(grpcClient.requests_pipelines[idx].Pipeline.GetName()))
					activeResources++
				}
			}
			g.Expect(len(grpcClient.requests_pipelines)).To(Equal(activeResources))
		})
	}
}

func TestHandleDeletedExperiments(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "with experiments",
			resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
		},
		{
			name: "with deleted experiments",
			resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bar",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ExperimentFinalizerName},
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
		},
		{
			name:      "no experiments",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newMockSchedulerClient(test.resources...)
			handlePendingDeleteExperiments(context.Background(), "", s)

			actualResourcesList := &mlopsv1alpha1.ExperimentList{}
			// Get all experiments in the namespace
			err := s.List(
				context.Background(),
				actualResourcesList,
				client.InNamespace(""),
			)
			g.Expect(err).To(BeNil())

			activeResources := 0
			for idx, req := range test.resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(actualResourcesList.Items[idx].GetName()))
					activeResources++
				}
			}
			g.Expect(len(actualResourcesList.Items)).To(Equal(activeResources))
		})
	}
}

func TestHandleDeletedPipelines(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "with pipelines",
			resources: []client.Object{
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
			},
		},
		{
			name: "with deleted pipelines",
			resources: []client.Object{
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bar",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.PipelineFinalizerName},
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
			},
		},
		{
			name:      "no experiments",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newMockSchedulerClient(test.resources...)
			handlePendingDeletePipelines(context.Background(), "", s)

			actualResourcesList := &mlopsv1alpha1.PipelineList{}
			// Get all pipelines in the namespace
			err := s.List(
				context.Background(),
				actualResourcesList,
				client.InNamespace(""),
			)
			g.Expect(err).To(BeNil())

			activeResources := 0
			for idx, req := range test.resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(actualResourcesList.Items[idx].GetName()))
					activeResources++
				}
			}
			g.Expect(len(actualResourcesList.Items)).To(Equal(activeResources))
		})
	}
}

func TestHandleDeletedModels(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "with models",
			resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
			},
		},
		{
			name: "with deleted models",
			resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bar",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ModelFinalizerName},
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
			},
		},
		{
			name:      "no models",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerClient{
				errors: map[string]error{
					"UnloadModel": status.Error(codes.FailedPrecondition, "no models"),
				},
			}
			s := newMockSchedulerClient(test.resources...)
			handlePendingDeleteModels(context.Background(), "", s, &grpcClient)

			actualResourcesList := &mlopsv1alpha1.ModelList{}
			// Get all models in the namespace
			err := s.List(
				context.Background(),
				actualResourcesList,
				client.InNamespace(""),
			)
			g.Expect(err).To(BeNil())

			activeResources := 0
			for idx, req := range test.resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(actualResourcesList.Items[idx].GetName()))
					activeResources++
				}
			}
			g.Expect(len(actualResourcesList.Items)).To(Equal(activeResources))
		})
	}
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
			client := mockSchedulerClient{responses_experiments: test.results}
			num, err := getNumExperimentsFromScheduler(context.Background(), &client)

			g.Expect(err).To(BeNil())
			g.Expect(num).To(Equal(len(test.results)))
		})
	}
}

func TestGetNumPipelines(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		results []*scheduler.PipelineStatusResponse
	}

	tests := []test{
		{
			name: "pipeline ok",
			results: []*scheduler.PipelineStatusResponse{
				{
					PipelineName: "foo",
				},
				{
					PipelineName: "bar",
				},
			},
		},
		{
			name:    "experiment ok",
			results: []*scheduler.PipelineStatusResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mockSchedulerClient{responses_pipelines: test.results}
			num, err := getNumPipelinesFromScheduler(context.Background(), &client)

			g.Expect(err).To(BeNil())
			g.Expect(num).To(Equal(len(test.results)))
		})
	}
}
