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

func TestSubscribeModelEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		existing_resources []client.Object
		results            []*scheduler.ModelStatusResponse
		noSchedulerState   bool
	}
	now := metav1.Now()

	// note expected state is derived in the test, maybe we should be explicitly about it in the future
	tests := []test{
		{
			name: "model available",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Status: mlopsv1alpha1.ModelStatus{
						Replicas: 1,
					},
				},
			},
			results: []*scheduler.ModelStatusResponse{
				{
					ModelName: "foo",
					Versions: []*scheduler.ModelVersionStatus{
						{
							KubernetesMeta: &scheduler.KubernetesMeta{
								Namespace:  "default",
								Generation: 1,
							},
							State: &scheduler.ModelStatus{
								State:             scheduler.ModelStatus_ModelAvailable,
								AvailableReplicas: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "model not available",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Status: mlopsv1alpha1.ModelStatus{
						Replicas: 1,
					},
				},
			},
			results: []*scheduler.ModelStatusResponse{
				{
					ModelName: "foo",
					Versions: []*scheduler.ModelVersionStatus{
						{
							KubernetesMeta: &scheduler.KubernetesMeta{
								Namespace:  "default",
								Generation: 1,
							},
							State: &scheduler.ModelStatus{
								State:               scheduler.ModelStatus_ModelProgressing,
								AvailableReplicas:   0,
								UnavailableReplicas: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "model being removed",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "foo",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ModelFinalizerName},
					},
					Status: mlopsv1alpha1.ModelStatus{
						Replicas: 1,
					},
				},
			},
			results: []*scheduler.ModelStatusResponse{
				{
					ModelName: "foo",
					Versions: []*scheduler.ModelVersionStatus{
						{
							KubernetesMeta: &scheduler.KubernetesMeta{
								Namespace:  "default",
								Generation: 1,
							},
							State: &scheduler.ModelStatus{
								State: scheduler.ModelStatus_ModelTerminated,
							},
						},
					},
				},
			},
		},
		{
			name: "model not removed",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "foo",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
						Finalizers:        []string{constants.ModelFinalizerName},
					},
					Status: mlopsv1alpha1.ModelStatus{
						Replicas: 1,
					},
				},
			},
			results: []*scheduler.ModelStatusResponse{
				{
					ModelName: "foo",
					Versions: []*scheduler.ModelVersionStatus{
						{
							KubernetesMeta: &scheduler.KubernetesMeta{
								Namespace:  "default",
								Generation: 1,
							},
							State: &scheduler.ModelStatus{
								State:             scheduler.ModelStatus_ModelTerminating,
								AvailableReplicas: 1,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// note that if responses_pipelines is nil -> scheduler state is not existing
			var grpcClient mockSchedulerGrpcClient
			if !test.noSchedulerState {
				grpcClient = mockSchedulerGrpcClient{
					responses_subscribe_models: test.results,
					responses_models:           test.results,
				}
			} else {
				grpcClient = mockSchedulerGrpcClient{
					responses_subscribe_models: test.results,
				}
			}
			controller := newMockControllerClient(test.existing_resources...)
			err := controller.handleModels(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())
			err = controller.SubscribeModelEvents(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())

			isBeingDeleted := map[string]bool{}
			for _, req := range test.existing_resources {
				if !req.GetDeletionTimestamp().IsZero() {
					isBeingDeleted[req.GetName()] = true
				} else {
					isBeingDeleted[req.GetName()] = false
				}
			}

			// for model resources that are not deleted we reload them
			// this is not necessary check but added for sanity
			activeResources := 0
			for idx, req := range test.existing_resources {
				if req.GetDeletionTimestamp().IsZero() {
					g.Expect(req.GetName()).To(Equal(grpcClient.requests_models[idx].Model.GetMeta().GetName()))
					activeResources++
				}
			}
			g.Expect(len(grpcClient.requests_models)).To(Equal(activeResources))

			// check state is correct for each model
			for _, r := range test.results {
				if r.Versions[0].State.GetState() != scheduler.ModelStatus_ModelTerminated {
					model := &mlopsv1alpha1.Model{}
					err := controller.Get(
						context.Background(),
						client.ObjectKey{
							Name:      r.GetModelName(),
							Namespace: r.Versions[0].KubernetesMeta.Namespace,
						},
						model,
					)
					// we check if the model is not in k8s (existing_resources) then we should not act on it
					if _, ok := isBeingDeleted[r.GetModelName()]; !ok {
						g.Expect(err).ToNot(BeNil())
					} else {
						g.Expect(err).To(BeNil())
					}
					if r.Versions[0].State.GetState() == scheduler.ModelStatus_ModelAvailable {
						g.Expect(model.Status.IsReady()).To(BeTrueBecause("Model state is ModelAvailable"))
					} else {
						g.Expect(model.Status.IsReady()).To(BeFalseBecause("Model state is not ModelAvailable"))
					}

					g.Expect(uint32(model.Status.Replicas)).To(Equal(r.Versions[0].State.GetAvailableReplicas() + r.Versions[0].State.GetUnavailableReplicas()))
					g.Expect(model.Status.Selector).To(Equal("server=" + r.Versions[0].ServerName))
				} else {
					model := &mlopsv1alpha1.Model{}
					err := controller.Get(
						context.Background(),
						client.ObjectKey{
							Name:      r.GetModelName(),
							Namespace: r.Versions[0].KubernetesMeta.Namespace,
						},
						model,
					)
					g.Expect(err).ToNot(BeNil())

				}
			}

		})
	}
}
