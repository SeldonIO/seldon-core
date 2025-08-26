/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package k8s

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/pkg/ptr"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s/mocks"
	mocks2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s/mocks/watch"
)

func TestHasPublishedIP_ImmediateResults(t *testing.T) {
	t.Parallel()

	createEndpointWithAddresses := func(name, resourceVersion string) *v1.EndpointSliceList {
		ep := &v1.EndpointSliceList{
			Items: []v1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Endpoints: []v1.Endpoint{
						{
							Addresses: []string{"1.2.3.4"},
						},
					},
				},
			},
		}
		ep.ResourceVersion = resourceVersion
		return ep
	}

	createEndpointWithoutAddresses := func(name, resourceVersion string) *v1.EndpointSliceList {
		ep := &v1.EndpointSliceList{
			Items: []v1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Endpoints: []v1.Endpoint{
						{},
					},
				},
			},
		}
		ep.ResourceVersion = resourceVersion
		return ep
	}

	tests := []struct {
		name          string
		serviceID     string
		endpoint      *v1.EndpointSliceList
		getError      error
		expectedError string
		shouldWatch   bool
		watchChan     func() chan watch.Event
	}{
		{
			name:        "success - immediate success from GET - endpoint has addresses",
			serviceID:   "test-service",
			endpoint:    createEndpointWithAddresses("test-service", "123"),
			getError:    nil,
			shouldWatch: false,
		},
		{
			name:          "failure - get endpoint error",
			serviceID:     "missing-service",
			endpoint:      nil,
			getError:      errors.New("endpoint not found"),
			expectedError: "failed to list endpointslice for missing-service: endpoint not found",
			shouldWatch:   false,
		},
		{
			name:        "success - endpoint without addresses - success on watch",
			serviceID:   "pending-service",
			endpoint:    createEndpointWithoutAddresses("pending-service", "123"),
			getError:    nil,
			shouldWatch: true,
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 1)
				event := watch.Event{
					Type: watch.Added,
					Object: &v1.EndpointSlice{
						Endpoints: []v1.Endpoint{
							{
								Addresses: []string{"1.2.3.4"},
							},
						},
					},
				}

				event.Object.(*v1.EndpointSlice).Labels = map[string]string{
					"kubernetes.io/service-name": "pending-service",
				}
				watchChan <- event
				return watchChan
			},
		},
		{
			name:          "failure - watch returned wrong event",
			serviceID:     "pending-service",
			endpoint:      createEndpointWithoutAddresses("pending-service", "123"),
			getError:      nil,
			shouldWatch:   true,
			expectedError: "event labels from channel <map[kubernetes.io/service-name:some-wrong-name]> does not match requested <kubernetes.io/service-name=pending-service>",
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 1)
				event := watch.Event{
					Type: watch.Added,
					Object: &v1.EndpointSlice{
						Endpoints: []v1.Endpoint{
							{
								Addresses: []string{"1.2.3.4"},
							},
						},
					},
				}

				event.Object.(*v1.EndpointSlice).Labels = map[string]string{
					"kubernetes.io/service-name": "some-wrong-name",
				}
				watchChan <- event
				return watchChan
			},
		},
		{
			name:        "success - watch - finds IP on second receive from channel",
			serviceID:   "pending-service",
			endpoint:    createEndpointWithoutAddresses("pending-service", "123"),
			getError:    nil,
			shouldWatch: true,
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 2)
				// 1st msg has IP but not ready
				event := watch.Event{
					Type: watch.Added,
					Object: &v1.EndpointSlice{
						Endpoints: []v1.Endpoint{
							{
								Addresses: []string{"1.2.3.4"},
								Conditions: v1.EndpointConditions{
									Ready: ptr.Bool(false),
								},
							},
						},
					},
				}

				event.Object.(*v1.EndpointSlice).Labels = map[string]string{
					"kubernetes.io/service-name": "pending-service",
				}
				watchChan <- event

				// 2nd msg has IP and is ready
				event2 := watch.Event{
					Type: watch.Added,
					Object: &v1.EndpointSlice{
						Endpoints: []v1.Endpoint{
							{
								Addresses: []string{"1.2.3.4"},
								Conditions: v1.EndpointConditions{
									Ready: ptr.Bool(true),
								},
							},
						},
					},
				}

				event2.Object.(*v1.EndpointSlice).Labels = map[string]string{
					"kubernetes.io/service-name": "pending-service",
				}
				watchChan <- event2
				return watchChan
			},
		},
		{
			name:        "success - watch - finds IP on second receive from channel, first event is error type",
			serviceID:   "pending-service",
			endpoint:    createEndpointWithoutAddresses("pending-service", "123"),
			getError:    nil,
			shouldWatch: true,
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 2)
				// 1st msg is Error so ignored
				event := watch.Event{
					Type: watch.Error,
					Object: &v1.EndpointSlice{
						Endpoints: []v1.Endpoint{
							{
								Addresses: []string{"1.2.3.4"},
								Conditions: v1.EndpointConditions{
									Ready: ptr.Bool(true),
								},
							},
						},
					},
				}

				event.Object.(*v1.EndpointSlice).Labels = map[string]string{
					"kubernetes.io/service-name": "pending-service",
				}
				watchChan <- event

				// 2nd msg has IP and is ready
				event2 := watch.Event{
					Type: watch.Added,
					Object: &v1.EndpointSlice{
						Endpoints: []v1.Endpoint{
							{
								Addresses: []string{"1.2.3.4"},
								Conditions: v1.EndpointConditions{
									Ready: ptr.Bool(true),
								},
							},
						},
					},
				}

				event2.Object.(*v1.EndpointSlice).Labels = map[string]string{
					"kubernetes.io/service-name": "pending-service",
				}
				watchChan <- event2
				return watchChan
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockK8sClient := mocks.NewMockInterface(ctrl)
			mockDiscoveryV1 := mocks.NewMockDiscoveryV1Interface(ctrl)
			mockEndpoints := mocks.NewMockEndpointSliceInterface(ctrl)

			mockK8sClient.EXPECT().DiscoveryV1().Return(mockDiscoveryV1)
			mockDiscoveryV1.EXPECT().EndpointSlices("test-ns").Return(mockEndpoints)
			mockEndpoints.EXPECT().List(gomock.Any(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", "kubernetes.io/service-name", tt.serviceID),
				Limit:         1,
			}).Return(tt.endpoint, tt.getError)

			if tt.shouldWatch {
				mockK8sClient.EXPECT().DiscoveryV1().Return(mockDiscoveryV1)
				mockDiscoveryV1.EXPECT().EndpointSlices("test-ns").Return(mockEndpoints)

				watcher := mocks2.NewMockInterface(ctrl)
				mockEndpoints.EXPECT().Watch(gomock.Any(), metav1.ListOptions{
					LabelSelector:   fmt.Sprintf("%s=%s", "kubernetes.io/service-name", tt.serviceID),
					ResourceVersion: tt.endpoint.ResourceVersion,
					Watch:           true,
				}).Return(watcher, nil)

				watcher.EXPECT().ResultChan().Return(tt.watchChan()).MinTimes(1)
				watcher.EXPECT().Stop()
			}

			client := NewExtendedClient("test-ns", mockK8sClient, logrus.New())

			var ctx context.Context
			var cancel context.CancelFunc
			if tt.shouldWatch {
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
			} else {
				ctx = context.Background()
			}

			err := client.HasPublishedIP(ctx, tt.serviceID, "")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}
