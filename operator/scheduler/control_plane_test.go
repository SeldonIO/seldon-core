/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"fmt"
	"testing"
	"time"

	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"

	. "github.com/onsi/gomega"
)

func TestSendWithTimeout(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		sleepTime time.Duration
		err       error
		isErr     bool
		isExpired bool
	}

	fn := func(err error) error {
		time.Sleep(5 * time.Millisecond)
		return err
	}

	tests := []test{
		{
			name:      "simple",
			sleepTime: 10 * time.Millisecond,
			err:       nil,
			isErr:     false,
			isExpired: false,
		},
		{
			name:      "timeout",
			sleepTime: 1 * time.Millisecond,
			err:       nil,
			isErr:     true,
			isExpired: true,
		},
		{
			name:      "error",
			sleepTime: 10 * time.Millisecond,
			err:       fmt.Errorf("error"),
			isErr:     true,
			isExpired: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hasExpired, err := execWithTimeout(func() error {
				return fn(test.err)
			}, test.sleepTime)
			g.Expect(hasExpired).To(Equal(test.isExpired))
			if test.isErr {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestControlPlaneEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                          string
		existing_resources            []client.Object
		expected_requests_pipelines   []*scheduler.LoadPipelineRequest
		expected_requests_models      []*scheduler.LoadModelRequest
		expected_requests_servers     []*scheduler.ServerNotify
		expected_requests_experiments []*scheduler.StartExperimentRequest
	}
	//now := metav1.Now()

	// note expected state is derived in the test, maybe we should be explictl about it in the future
	tests := []test{
		{
			name: "with no deleting resources",
			existing_resources: []client.Object{
				&mlopsv1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.PipelineSpec{},
				},
				&mlopsv1alpha1.Model{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ModelSpec{},
				},
				&mlopsv1alpha1.Experiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ExperimentSpec{},
				},
				&mlopsv1alpha1.Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: mlopsv1alpha1.ServerSpec{},
				},
			},
			expected_requests_pipelines: []*scheduler.LoadPipelineRequest{
				{
					Pipeline: &scheduler.Pipeline{
						KubernetesMeta: &scheduler.KubernetesMeta{
							Namespace:  "default",
							Generation: 1,
						},
						Name: "foo",
					},
				},
			},
			expected_requests_models: []*scheduler.LoadModelRequest{
				{
					Model: &scheduler.Model{
						Meta: &scheduler.MetaData{
							Name: "foo",
							KubernetesMeta: &scheduler.KubernetesMeta{
								Namespace:  "default",
								Generation: 1,
							},
						},
						ModelSpec: &scheduler.ModelSpec{},
						DeploymentSpec: &scheduler.DeploymentSpec{
							Replicas: 1,
						},
					},
				},
			},
			expected_requests_experiments: []*scheduler.StartExperimentRequest{
				{
					Experiment: &scheduler.Experiment{
						KubernetesMeta: &scheduler.KubernetesMeta{
							Namespace:  "default",
							Generation: 1,
						},
						Name: "foo",
					},
				},
			},
			expected_requests_servers: []*scheduler.ServerNotify{
				{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
					ExpectedReplicas: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerGrpcClient{}

			controller := newMockControllerClient(test.existing_resources...)

			err := controller.SubscribeControlPlaneEvents(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())

			// check state is correct for each resource
			for _, r := range test.expected_requests_pipelines {
				g.Expect(grpcClient.requests_pipelines).To(ContainElement(r))
			}
			for _, r := range test.expected_requests_experiments {
				g.Expect(grpcClient.requests_experiments).To(ContainElement(r))
			}
			for _, r := range test.expected_requests_models {
				g.Expect(grpcClient.requests_models).To(ContainElement(r))
			}
			for _, r := range test.expected_requests_servers {
				g.Expect(grpcClient.requests_servers).To(ContainElement(r))
			}

		})
	}
}
