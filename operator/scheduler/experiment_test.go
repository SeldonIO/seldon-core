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
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
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

	// note expected state is derived in the test, maybe we should be explictl about it in the future
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
		{
			name: "experiment does not exist in k8s",
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
					ExperimentName:  "foo2",
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
			err := controller.SubscribeExperimentEvents(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())

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

			// check state is correct for each experiment
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

					// we check if the experiement is not in k8s (existing_resources) then we should not act on it
					if _, ok := isBeingDeleted[r.ExperimentName]; !ok {
						g.Expect(err).ToNot(BeNil())
					} else {
						g.Expect(err).To(BeNil())
					}

					if r.CandidatesReady && r.Active && r.MirrorReady {
						g.Expect(experiment.Status.IsReady()).To(BeTrueBecause("All CandidatesReady, Active and MirrorReady are true"))
					} else {
						g.Expect(experiment.Status.IsReady()).To(BeFalseBecause("Either CandidatesReady, Active and MirrorReady are false"))
					}

					if r.Active {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.ExperimentReady)).To(BeTrueBecause("Active is true"))
					} else {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.ExperimentReady)).To(BeFalseBecause("Active is false"))
					}

					if r.MirrorReady {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.MirrorReady)).To(BeTrueBecause("MirrorReady is true"))
					} else {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.MirrorReady)).To(BeFalseBecause("MirrorReady is false"))
					}

					if r.CandidatesReady {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.CandidatesReady)).To(BeTrueBecause("CandidatesReady is true"))
					} else {
						g.Expect(experiment.Status.IsConditionReady(mlopsv1alpha1.CandidatesReady)).To(BeFalseBecause("CandidatesReady is false"))
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
						g.Expect(r.Active).To(BeFalseBecause("Experiment is not in k8s"))
					} else {
						g.Expect(r.Active).To(BeTrueBecause("Experiment is in k8s"))
					}
				}
			}

		})
	}
}
