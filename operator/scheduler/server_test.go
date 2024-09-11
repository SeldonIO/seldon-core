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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

func TestServerNotify(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		server         *v1alpha1.Server
		expectedProtos []*scheduler.ServerNotify
	}
	getIntPtr := func(val int32) *int32 { return &val }
	now := metav1.Now()
	tests := []test{
		{
			name: "good server - replicas set",
			server: &v1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: v1alpha1.ServerSpec{
					ScalingSpec: v1alpha1.ScalingSpec{
						Replicas: getIntPtr(2),
					},
				},
			},
			expectedProtos: []*scheduler.ServerNotify{
				{
					Name:             "foo",
					ExpectedReplicas: 2,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
		},
		{
			name: "good server - replicas not set",
			server: &v1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
			},
			expectedProtos: []*scheduler.ServerNotify{
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
			name: "deleted server",
			server: &v1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
					DeletionTimestamp: &now,
				},
				Spec: v1alpha1.ServerSpec{
					ScalingSpec: v1alpha1.ScalingSpec{
						Replicas: getIntPtr(2),
					},
				},
			},
			expectedProtos: []*scheduler.ServerNotify{
				{
					Name:             "foo",
					ExpectedReplicas: 0,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// note that responses_experiments is nill -> scheduler state is not existing
			grpcClient := mockSchedulerGrpcClient{
				requests_servers: []*scheduler.ServerNotifyRequest{},
			}
			controller := newMockControllerClient()
			err := controller.ServerNotify(context.Background(), &grpcClient, test.server)
			g.Expect(err).To(BeNil())

			g.Expect(len(grpcClient.requests_servers)).To(Equal(1))
			g.Expect(grpcClient.requests_servers[0].Servers).To(Equal(test.expectedProtos))

		})
	}
}
