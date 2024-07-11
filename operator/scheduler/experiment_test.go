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

func TestSubscribeExperimentsEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		existing_resources []client.Object
		results            []*scheduler.ExperimentStatusResponse
		noSchedulerState   bool
	}
	now := metav1.Now()

	tests := []test{
		{
			name: "experiment ready",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "foo",
					Active:          true,
					CandidatesReady: true,
					MirrorReady:     true,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: true,
		},
		{
			name: "experiment ready - with scheduler state",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "foo",
					Active:          true,
					CandidatesReady: true,
					MirrorReady:     true,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name: "experiment terminated - with scheduler state",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "foo",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ExperimentFinalizerName},
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName: "foo",
					Active:         false,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name: "candidate not ready and expriment ready",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "foo",
					Active:          true,
					CandidatesReady: false,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name: "candiadate and experiment not ready",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
			},
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "foo",
					Active:          false,
					CandidatesReady: false,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name: "with deleted experiments",
			existing_resources: []client.Object{
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
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "bar",
					Active:          false,
					CandidatesReady: false,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: true,
		},
		{
			name: "with deleted experiments - with scheduler state",
			existing_resources: []client.Object{
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
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "bar",
					Active:          false,
					CandidatesReady: false,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name: "with deleted experiments - no remove - with scheduler state",
			existing_resources: []client.Object{
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
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName:  "bar",
					Active:          true,
					CandidatesReady: true,
					MirrorReady:     true,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
			noSchedulerState: false,
		},
		{
			name:               "no experiments",
			existing_resources: []client.Object{},
			noSchedulerState:   true,
		},
		{
			name:               "no experiments - with scheduler state",
			existing_resources: []client.Object{},
			noSchedulerState:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// note that responses_experiments is nill -> scheduler state is not existing
			var grpcClient mockSchedulerGrpcClient
			if !test.noSchedulerState {
				grpcClient = mockSchedulerGrpcClient{
					responses_subscribe_experiments: test.results,
					responses_experiments:           test.results,
				}
			} else {
				grpcClient = mockSchedulerGrpcClient{
					responses_subscribe_experiments: test.results,
				}
			}
			controller := newMockControllerClient(test.existing_resources...)
			controller.SubscribeExperimentEvents(context.Background(), &grpcClient, "")

			isBeingDeleted := map[string]bool{}
			for _, req := range test.existing_resources {
				if !req.GetDeletionTimestamp().IsZero() {
					isBeingDeleted[req.GetName()] = true
				} else {
					isBeingDeleted[req.GetName()] = false
				}
			}
			// check that we have reloaded the correct resources if the stata of the scheduler is not correct
			if test.noSchedulerState {
				activeResources := 0
				for idx, req := range test.existing_resources {
					if req.GetDeletionTimestamp().IsZero() {
						g.Expect(req.GetName()).To(Equal(grpcClient.requests_experiments[idx].Experiment.GetName()))
						activeResources++
					}
				}
				g.Expect(len(grpcClient.requests_experiments)).To(Equal(activeResources))
			} else {
				g.Expect(len(grpcClient.requests_experiments)).To(Equal(0))
			}

			// check state is correct for each pipeline
			for _, r := range test.results {
				if !isBeingDeleted[r.ExperimentName] {
					experiment := &mlopsv1alpha1.Experiment{}
					err := controller.Get(
						context.Background(),
						client.ObjectKey{
							Name:      r.ExperimentName,
							Namespace: r.GetKubernetesMeta().Namespace,
						},
						experiment,
					)
					g.Expect(err).To(BeNil())
					if r.CandidatesReady && r.Active && r.MirrorReady {
						g.Expect(experiment.Status.IsReady()).To(BeTrue())
					} else {
						g.Expect(experiment.Status.IsReady()).To(BeFalse())
					}

					if r.Active {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.ExperimentReady)).To(BeTrue())
					} else {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.ExperimentReady)).To(BeFalse())
					}

					if r.MirrorReady {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.MirrorReady)).To(BeTrue())
					} else {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.MirrorReady)).To(BeFalse())
					}

					if r.CandidatesReady {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.CandidatesReady)).To(BeTrue())
					} else {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.CandidatesReady)).To(BeFalse())
					}

				} else {
					experiment := &mlopsv1alpha1.Experiment{}
					err := controller.Get(
						context.Background(),
						client.ObjectKey{
							Name:      r.ExperimentName,
							Namespace: r.KubernetesMeta.Namespace,
						},
						experiment,
					)
					if err != nil { // in case the experiment is remove from k8s we should get an error and active is false
						g.Expect(r.Active).To(BeFalse())
					} else {
						g.Expect(r.Active).To(BeTrue())
					}
				}
			}

		})
	}
}
