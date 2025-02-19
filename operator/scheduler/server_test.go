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

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

var getIntPtr = func(val int32) *int32 { return &val }

func TestServerNotify(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		servers        []v1alpha1.Server
		expectedProtos []*scheduler.ServerNotify
	}
	now := metav1.Now()
	tests := []test{
		{
			name: "good server - replicas set",
			servers: []v1alpha1.Server{
				{
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
			servers: []v1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
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
			servers: []v1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "foo",
						Namespace:         "default",
						Generation:        1,
						DeletionTimestamp: &now,
					},
					Spec: v1alpha1.ServerSpec{
						ScalingSpec: v1alpha1.ScalingSpec{
							Replicas: getIntPtr(2),
						},
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
		{
			name:           "nil servers",
			expectedProtos: []*scheduler.ServerNotify{},
		},
		{
			name: "list of servers",
			servers: []v1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "foo",
						Namespace:  "default",
						Generation: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "bar",
						Namespace:  "default",
						Generation: 2,
					},
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
				{
					Name:             "bar",
					ExpectedReplicas: 1,
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerGrpcClient{
				requests_servers: []*scheduler.ServerNotify{},
			}
			controller := newMockControllerClient()
			err := controller.ServerNotify(context.Background(), &grpcClient, test.servers, false)
			g.Expect(err).To(BeNil())

			if len(test.servers) != 0 {
				g.Expect(len(grpcClient.requests_servers)).To(Equal(len(test.expectedProtos)))
				g.Expect(grpcClient.requests_servers).To(Equal(test.expectedProtos))
			} else {
				g.Expect(len(grpcClient.requests_servers)).To(Equal(0))
			}
		})
	}
}

func TestSubscribeServerEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		existingServer mlopsv1alpha1.Server
		response       *scheduler.ServerStatusResponse
	}

	// note expected state is derived in the test, maybe we should be explicit about it in the future
	tests := []test{
		{
			name: "server - with a valid scaling request",
			existingServer: mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "seldon",
					Generation: 1,
				},
				Spec: v1alpha1.ServerSpec{
					ScalingSpec: v1alpha1.ScalingSpec{
						Replicas:    getIntPtr(1),
						MinReplicas: getIntPtr(1),
						MaxReplicas: getIntPtr(5),
					},
				},
			},
			response: &scheduler.ServerStatusResponse{
				Type:       scheduler.ServerStatusResponse_ScalingRequest,
				ServerName: "foo",
				Resources: []*scheduler.ServerReplicaResources{
					{
						ReplicaIdx: 0,
					},
				},
				AvailableReplicas:      1,
				ExpectedReplicas:       2,
				NumLoadedModelReplicas: 0,
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "seldon",
					Generation: 1,
				},
			},
		},
		{
			name: "server - with a valid scaling request - different generation",
			existingServer: mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "seldon",
					Generation: 2,
				},
				Spec: v1alpha1.ServerSpec{
					ScalingSpec: v1alpha1.ScalingSpec{
						Replicas:    getIntPtr(1),
						MinReplicas: getIntPtr(1),
						MaxReplicas: getIntPtr(5),
					},
				},
			},
			response: &scheduler.ServerStatusResponse{
				Type:       scheduler.ServerStatusResponse_ScalingRequest,
				ServerName: "foo",
				Resources: []*scheduler.ServerReplicaResources{
					{
						ReplicaIdx: 0,
					},
				},
				AvailableReplicas:      1,
				ExpectedReplicas:       2,
				NumLoadedModelReplicas: 0,
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "seldon",
					Generation: 1, // older generation is still allowed for scaling requests
				},
			},
		},
		{
			name: "server - with an invalid scaling request",
			existingServer: mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "seldon",
					Generation: 1,
				},
				Spec: v1alpha1.ServerSpec{
					ScalingSpec: v1alpha1.ScalingSpec{
						Replicas:    getIntPtr(1),
						MinReplicas: getIntPtr(1),
						MaxReplicas: getIntPtr(5),
					},
				},
			},
			response: &scheduler.ServerStatusResponse{
				Type:       scheduler.ServerStatusResponse_ScalingRequest,
				ServerName: "foo",
				Resources: []*scheduler.ServerReplicaResources{
					{
						ReplicaIdx: 0,
					},
				},
				AvailableReplicas:      1,
				ExpectedReplicas:       6,
				NumLoadedModelReplicas: 0,
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "seldon",
					Generation: 1,
				},
			},
		},
		{
			name: "server - with no scaling spec",
			existingServer: mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "seldon",
					Generation: 1,
				},
			},
			response: &scheduler.ServerStatusResponse{
				Type:       scheduler.ServerStatusResponse_ScalingRequest,
				ServerName: "foo",
				Resources: []*scheduler.ServerReplicaResources{
					{
						ReplicaIdx: 0,
					},
				},
				AvailableReplicas:      3,
				ExpectedReplicas:       6,
				NumLoadedModelReplicas: 0,
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "seldon",
					Generation: 1,
				},
			},
		},
		{
			// no scheduler state means lost servers metadata
			name: "servers - no scheduler state",
			existingServer: mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
			},
			response: &scheduler.ServerStatusResponse{
				Type:       scheduler.ServerStatusResponse_NonAuthoritativeReplicaInfo,
				ServerName: "foo",
				Resources: []*scheduler.ServerReplicaResources{
					{
						ReplicaIdx: 0,
					},
				},
				AvailableReplicas:      1,
				ExpectedReplicas:       1,
				NumLoadedModelReplicas: 0, // no update
			},
		},
		{
			name: "server - with scheduler state",
			existingServer: mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "seldon",
					Generation: 1,
				},
			},
			response: &scheduler.ServerStatusResponse{
				Type:       scheduler.ServerStatusResponse_StatusUpdate,
				ServerName: "foo",
				Resources: []*scheduler.ServerReplicaResources{
					{
						ReplicaIdx: 0,
					},
				},
				AvailableReplicas:      1,
				ExpectedReplicas:       1,
				NumLoadedModelReplicas: 2,
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "seldon",
					Generation: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grpcClient := mockSchedulerGrpcClient{
				responses_subscribe_servers: []*scheduler.ServerStatusResponse{test.response},
			}

			existing_resources := []client.Object{&test.existingServer}
			controller := newMockControllerClient(existing_resources...)
			err := controller.SubscribeServerEvents(context.Background(), &grpcClient, "")
			g.Expect(err).To(BeNil())

			// check state is correct for each server

			namespace := "default"
			if test.response.KubernetesMeta != nil {
				namespace = test.response.KubernetesMeta.Namespace
			}
			server := &mlopsv1alpha1.Server{}
			err = controller.Get(
				context.Background(),
				client.ObjectKey{
					Name:      test.response.ServerName,
					Namespace: namespace,
				},
				server,
			)
			g.Expect(err).To(BeNil())
			g.Expect(server.Status.LoadedModelReplicas).To(Equal(test.response.NumLoadedModelReplicas))
			isValidEvent, _, err := isValidScalingEvent(&test.existingServer, test.response)
			g.Expect(err).To(BeNil())
			if isValidEvent && test.existingServer.Spec.Replicas != nil {
				g.Expect(*server.Spec.Replicas).To(Equal(test.response.ExpectedReplicas))
			} else {
				g.Expect(server.Spec.Replicas).To(Equal(test.existingServer.Spec.Replicas))
			}
		})
	}
}
