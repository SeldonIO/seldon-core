/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"context"
	"io"
	"testing"

	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
)

// Experiment mock grpc client

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

// Pipeline subscribe mock grpc client

type mockSchedulerPipelineSubscribeGrpcClient struct {
	counter int
	results []*scheduler.PipelineStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_SubscribePipelineStatusClient = (*mockSchedulerPipelineSubscribeGrpcClient)(nil)

func newMockSchedulerPipelineSubscribeGrpcClient(results []*scheduler.PipelineStatusResponse) *mockSchedulerPipelineSubscribeGrpcClient {
	return &mockSchedulerPipelineSubscribeGrpcClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerPipelineSubscribeGrpcClient) Recv() (*scheduler.PipelineStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

// Server subscribe mock grpc client

type mockSchedulerServerSubscribeGrpcClient struct {
	counter int
	results []*scheduler.ServerStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_SubscribeServerStatusClient = (*mockSchedulerServerSubscribeGrpcClient)(nil)

func newMockSchedulerServerSubscribeGrpcClient(results []*scheduler.ServerStatusResponse) *mockSchedulerServerSubscribeGrpcClient {
	return &mockSchedulerServerSubscribeGrpcClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerServerSubscribeGrpcClient) Recv() (*scheduler.ServerStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

// Control Plane subscribe mock grpc client

type mockControlPlaneSubscribeGrpcClient struct {
	sent int
	grpc.ClientStream
}

var _ scheduler.Scheduler_SubscribeControlPlaneClient = (*mockControlPlaneSubscribeGrpcClient)(nil)

func newMockControlPlaneSubscribeGrpcClient() *mockControlPlaneSubscribeGrpcClient {
	return &mockControlPlaneSubscribeGrpcClient{}
}

func (s *mockControlPlaneSubscribeGrpcClient) Recv() (*scheduler.ControlPlaneResponse, error) {
	if s.sent == 0 {
		s.sent++
		return &scheduler.ControlPlaneResponse{Event: scheduler.ControlPlaneResponse_SEND_SERVERS}, nil
	}
	if s.sent == 1 {
		s.sent++
		return &scheduler.ControlPlaneResponse{Event: scheduler.ControlPlaneResponse_SEND_RESOURCES}, nil
	}
	return nil, io.EOF
}

// Pipeline mock grpc client

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

// Experiment subscribe mock grpc client

type mockSchedulerExperimentSubscribeGrpcClient struct {
	counter int
	results []*scheduler.ExperimentStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_SubscribeExperimentStatusClient = (*mockSchedulerExperimentSubscribeGrpcClient)(nil)

func newMockSchedulerExperimentSubscribeGrpcClient(results []*scheduler.ExperimentStatusResponse) *mockSchedulerExperimentSubscribeGrpcClient {
	return &mockSchedulerExperimentSubscribeGrpcClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerExperimentSubscribeGrpcClient) Recv() (*scheduler.ExperimentStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

// Model subscribe mock grpc client

type mockSchedulerModelSubscribeGrpcClient struct {
	counter int
	results []*scheduler.ModelStatusResponse
	grpc.ClientStream
}

var _ scheduler.Scheduler_SubscribeModelStatusClient = (*mockSchedulerModelSubscribeGrpcClient)(nil)

func newMockSchedulerModelSubscribeGrpcClient(results []*scheduler.ModelStatusResponse) *mockSchedulerModelSubscribeGrpcClient {
	return &mockSchedulerModelSubscribeGrpcClient{
		results: results,
		counter: 0,
	}
}

func (s *mockSchedulerModelSubscribeGrpcClient) Recv() (*scheduler.ModelStatusResponse, error) {
	if s.counter < len(s.results) {
		s.counter++
		return s.results[s.counter-1], nil
	}
	return nil, io.EOF
}

// Scheduler mock grpc client

type mockSchedulerGrpcClient struct {
	responses_experiments           []*scheduler.ExperimentStatusResponse
	responses_subscribe_experiments []*scheduler.ExperimentStatusResponse
	responses_pipelines             []*scheduler.PipelineStatusResponse
	responses_subscribe_pipelines   []*scheduler.PipelineStatusResponse
	responses_servers               []*scheduler.ServerStatusResponse
	responses_subscribe_servers     []*scheduler.ServerStatusResponse
	responses_models                []*scheduler.ModelStatusResponse
	responses_subscribe_models      []*scheduler.ModelStatusResponse
	requests_experiments            []*scheduler.StartExperimentRequest
	requests_experiments_unload     []*scheduler.StopExperimentRequest
	requests_pipelines              []*scheduler.LoadPipelineRequest
	requests_pipelines_unload       []*scheduler.UnloadPipelineRequest
	requests_models                 []*scheduler.LoadModelRequest
	requests_models_unload          []*scheduler.UnloadModelRequest
	requests_servers                []*scheduler.ServerNotify
	errors                          map[string]error
}

var _ scheduler.SchedulerClient = (*mockSchedulerGrpcClient)(nil)

func (s *mockSchedulerGrpcClient) ExperimentStatus(ctx context.Context, in *scheduler.ExperimentStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ExperimentStatusClient, error) {
	return newMockSchedulerExperimentGrpcClient(s.responses_experiments), nil
}

func (s *mockSchedulerGrpcClient) ServerNotify(ctx context.Context, in *scheduler.ServerNotifyRequest, opts ...grpc.CallOption) (*scheduler.ServerNotifyResponse, error) {
	s.requests_servers = append(s.requests_servers, in.Servers...)
	return nil, nil
}

func (s *mockSchedulerGrpcClient) LoadModel(ctx context.Context, in *scheduler.LoadModelRequest, opts ...grpc.CallOption) (*scheduler.LoadModelResponse, error) {
	s.requests_models = append(s.requests_models, in)
	return nil, nil
}
func (s *mockSchedulerGrpcClient) UnloadModel(ctx context.Context, in *scheduler.UnloadModelRequest, opts ...grpc.CallOption) (*scheduler.UnloadModelResponse, error) {
	err, ok := s.errors["UnloadModel"]
	if ok {
		return nil, err
	} else {
		s.requests_models_unload = append(s.requests_models_unload, in)
		return nil, nil
	}
}
func (s *mockSchedulerGrpcClient) LoadPipeline(ctx context.Context, in *scheduler.LoadPipelineRequest, opts ...grpc.CallOption) (*scheduler.LoadPipelineResponse, error) {
	s.requests_pipelines = append(s.requests_pipelines, in)
	return nil, nil
}
func (s *mockSchedulerGrpcClient) UnloadPipeline(ctx context.Context, in *scheduler.UnloadPipelineRequest, opts ...grpc.CallOption) (*scheduler.UnloadPipelineResponse, error) {
	s.requests_pipelines_unload = append(s.requests_pipelines_unload, in)
	return nil, nil
}
func (s *mockSchedulerGrpcClient) StartExperiment(ctx context.Context, in *scheduler.StartExperimentRequest, opts ...grpc.CallOption) (*scheduler.StartExperimentResponse, error) {
	s.requests_experiments = append(s.requests_experiments, in)
	return nil, nil
}
func (s *mockSchedulerGrpcClient) StopExperiment(ctx context.Context, in *scheduler.StopExperimentRequest, opts ...grpc.CallOption) (*scheduler.StopExperimentResponse, error) {
	s.requests_experiments_unload = append(s.requests_experiments_unload, in)
	return nil, nil
}
func (s *mockSchedulerGrpcClient) ServerStatus(ctx context.Context, in *scheduler.ServerStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ServerStatusClient, error) {
	// only used for seldon cli, not used in controller
	return nil, nil

}
func (s *mockSchedulerGrpcClient) ModelStatus(ctx context.Context, in *scheduler.ModelStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_ModelStatusClient, error) {
	// only used for seldon cli, not used in controller
	return nil, nil
}
func (s *mockSchedulerGrpcClient) PipelineStatus(ctx context.Context, in *scheduler.PipelineStatusRequest, opts ...grpc.CallOption) (scheduler.Scheduler_PipelineStatusClient, error) {
	return newMockSchedulerPipelineGrpcClient(s.responses_pipelines), nil
}
func (s *mockSchedulerGrpcClient) SchedulerStatus(ctx context.Context, in *scheduler.SchedulerStatusRequest, opts ...grpc.CallOption) (*scheduler.SchedulerStatusResponse, error) {
	return nil, nil
}
func (s *mockSchedulerGrpcClient) SubscribeServerStatus(ctx context.Context, in *scheduler.ServerSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeServerStatusClient, error) {
	return newMockSchedulerServerSubscribeGrpcClient(s.responses_subscribe_servers), nil
}
func (s *mockSchedulerGrpcClient) SubscribeModelStatus(ctx context.Context, in *scheduler.ModelSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeModelStatusClient, error) {
	return newMockSchedulerModelSubscribeGrpcClient(s.responses_subscribe_models), nil
}
func (s *mockSchedulerGrpcClient) SubscribeExperimentStatus(ctx context.Context, in *scheduler.ExperimentSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeExperimentStatusClient, error) {
	return newMockSchedulerExperimentSubscribeGrpcClient(s.responses_subscribe_experiments), nil
}
func (s *mockSchedulerGrpcClient) SubscribePipelineStatus(ctx context.Context, in *scheduler.PipelineSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribePipelineStatusClient, error) {
	return newMockSchedulerPipelineSubscribeGrpcClient(s.responses_subscribe_pipelines), nil
}
func (s *mockSchedulerGrpcClient) SubscribeControlPlane(ctx context.Context, in *scheduler.ControlPlaneSubscriptionRequest, opts ...grpc.CallOption) (scheduler.Scheduler_SubscribeControlPlaneClient, error) {
	return newMockControlPlaneSubscribeGrpcClient(), nil
}

// new mockSchedulerClient (not grpc)
func newMockControllerClient(objs ...client.Object) *SchedulerClient {
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
			grpcClient := mockSchedulerGrpcClient{}
			client := newMockControllerClient(test.resources...)
			err := client.handleLoadedExperiments(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())
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
			grpcClient := mockSchedulerGrpcClient{}
			client := newMockControllerClient(test.resources...)
			err := client.handleLoadedModels(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())
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
			grpcClient := mockSchedulerGrpcClient{}
			client := newMockControllerClient(test.resources...)
			err := client.handleLoadedPipelines(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())
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
			s := newMockControllerClient(test.resources...)
			err := s.handlePendingDeleteExperiments(context.Background(), "")
			g.Expect(err).To(BeNil())

			actualResourcesList := &mlopsv1alpha1.ExperimentList{}
			// Get all experiments in the namespace
			err = s.List(
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
			s := newMockControllerClient(test.resources...)
			err := s.handlePendingDeletePipelines(context.Background(), "")
			g.Expect(err).To(BeNil())

			actualResourcesList := &mlopsv1alpha1.PipelineList{}
			// Get all pipelines in the namespace
			err = s.List(
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
			grpcClient := mockSchedulerGrpcClient{
				errors: map[string]error{
					"UnloadModel": status.Error(codes.FailedPrecondition, "no models"),
				},
			}
			s := newMockControllerClient(test.resources...)
			err := s.handlePendingDeleteModels(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())

			actualResourcesList := &mlopsv1alpha1.ModelList{}
			// Get all models in the namespace
			err = s.List(
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

func TestHandleRegisteredServers(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		resources []client.Object
		expected  []*scheduler.ServerNotify
	}

	tests := []test{
		{
			name: "with 1 server",
			resources: []client.Object{
				&mlopsv1alpha1.Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			expected: []*scheduler.ServerNotify{
				{
					Name:             "foo",
					ExpectedReplicas: 1,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
		},
		{
			name: "with multiple servers",
			resources: []client.Object{
				&mlopsv1alpha1.Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
				},
				&mlopsv1alpha1.Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "bar",
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			expected: []*scheduler.ServerNotify{
				{
					Name:             "bar",
					ExpectedReplicas: 1,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name:             "foo",
					ExpectedReplicas: 1,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
		},
		{
			name:      "no servers",
			resources: []client.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerGrpcClient{}
			client := newMockControllerClient(test.resources...)
			err := client.handleRegisteredServers(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())
			g.Expect(grpcClient.requests_servers).To(Equal(test.expected))
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
			client := mockSchedulerGrpcClient{responses_experiments: test.results}
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
			name:    "no pipelines",
			results: []*scheduler.PipelineStatusResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mockSchedulerGrpcClient{responses_pipelines: test.results}
			num, err := getNumPipelinesFromScheduler(context.Background(), &client)

			g.Expect(err).To(BeNil())
			g.Expect(num).To(Equal(len(test.results)))
		})
	}
}
