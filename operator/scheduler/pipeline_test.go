package scheduler

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSubscribePipelineEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		resources        []client.Object
		results          []*scheduler.PipelineStatusResponse
		noSchedulerState bool
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "model and pipeline ready",
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
			results: []*scheduler.PipelineStatusResponse{
				{
					PipelineName: "foo",
					Versions: []*scheduler.PipelineWithState{
						{
							Pipeline: &scheduler.Pipeline{
								Name: "foo",
								KubernetesMeta: &scheduler.KubernetesMeta{
									Namespace:  "default",
									Generation: 1,
								},
							},
							State: &scheduler.PipelineVersionState{
								Status:      scheduler.PipelineVersionState_PipelineReady,
								ModelsReady: true,
							},
						},
					},
				},
			},
			noSchedulerState: true,
		},
		{
			name: "model and pipeline ready - with scheduler state",
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
			results: []*scheduler.PipelineStatusResponse{
				{
					PipelineName: "foo",
					Versions: []*scheduler.PipelineWithState{
						{
							Pipeline: &scheduler.Pipeline{
								Name: "foo",
								KubernetesMeta: &scheduler.KubernetesMeta{
									Namespace:  "default",
									Generation: 1,
								},
							},
							State: &scheduler.PipelineVersionState{
								Status:      scheduler.PipelineVersionState_PipelineReady,
								ModelsReady: true,
							},
						},
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name: "model not ready and pipeline ready",
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
			results: []*scheduler.PipelineStatusResponse{
				{
					PipelineName: "foo",
					Versions: []*scheduler.PipelineWithState{
						{
							Pipeline: &scheduler.Pipeline{
								Name: "foo",
								KubernetesMeta: &scheduler.KubernetesMeta{
									Namespace:  "default",
									Generation: 1,
								},
							},
							State: &scheduler.PipelineVersionState{
								Status:      scheduler.PipelineVersionState_PipelineReady,
								ModelsReady: false,
							},
						},
					},
				},
			},
			noSchedulerState: true,
		},
		{
			name: "model and pipeline not ready",
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
			results: []*scheduler.PipelineStatusResponse{
				{
					PipelineName: "foo",
					Versions: []*scheduler.PipelineWithState{
						{
							Pipeline: &scheduler.Pipeline{
								Name: "foo",
								KubernetesMeta: &scheduler.KubernetesMeta{
									Namespace:  "default",
									Generation: 1,
								},
							},
							State: &scheduler.PipelineVersionState{
								Status:      scheduler.PipelineVersionState_PipelineFailed,
								ModelsReady: false,
							},
						},
					},
				},
			},
			noSchedulerState: true,
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
			noSchedulerState: true,
		},
		{
			name:             "no pipelines",
			resources:        []client.Object{},
			noSchedulerState: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// note that responses_pipelines is nill -> scheduler state is not existing
			var grpcClient mockSchedulerGrpcClient
			if !test.noSchedulerState {
				grpcClient = mockSchedulerGrpcClient{
					responses_subscribe_pipelines: test.results,
					responses_pipelines:           test.results,
				}
			} else {
				grpcClient = mockSchedulerGrpcClient{
					responses_subscribe_pipelines: test.results,
				}
			}
			controller := newMockControllerClient(test.resources...)
			controller.SubscribePipelineEvents(context.Background(), &grpcClient, "")

			// check that we have reloaded the correct resources if the stata of the scheduler is not correct
			if test.noSchedulerState {
				activeResources := 0
				for idx, req := range test.resources {
					if req.GetDeletionTimestamp().IsZero() {
						g.Expect(req.GetName()).To(Equal(grpcClient.requests_pipelines[idx].Pipeline.GetName()))
						activeResources++
					}
				}
				g.Expect(len(grpcClient.requests_pipelines)).To(Equal(activeResources))
			} else {
				g.Expect(len(grpcClient.requests_pipelines)).To(Equal(0))
			}

			// check state is correct for each pipeline
			for _, r := range test.results {
				if r.Versions[0].State.Status != scheduler.PipelineVersionState_PipelineTerminated {
					pipeline := &mlopsv1alpha1.Pipeline{}
					err := controller.Get(
						context.Background(),
						client.ObjectKey{
							Name:      r.PipelineName,
							Namespace: r.Versions[0].Pipeline.KubernetesMeta.Namespace,
						},
						pipeline,
					)
					g.Expect(err).To(BeNil())
					if r.Versions[0].State.Status == scheduler.PipelineVersionState_PipelineReady && r.Versions[0].State.ModelsReady {
						g.Expect(pipeline.Status.IsReady()).To(BeTrue())
					} else {
						g.Expect(pipeline.Status.IsReady()).To(BeFalse())
					}
					if r.Versions[0].State.ModelsReady {
						g.Expect(pipeline.Status.IsConditionReady(mlopsv1alpha1.ModelsReady)).To(BeTrue())
					} else {
						g.Expect(pipeline.Status.IsConditionReady(mlopsv1alpha1.ModelsReady)).To(BeFalse())
					}

				} else {
					pipeline := &mlopsv1alpha1.Pipeline{}
					err := controller.Get(
						context.Background(),
						client.ObjectKey{
							Name:      r.PipelineName,
							Namespace: r.Versions[0].Pipeline.KubernetesMeta.Namespace,
						},
						pipeline,
					)
					g.Expect(err).ToNot(BeNil())

				}
			}

		})
	}
}
