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
	v2 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s/mocks"
	mocks2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s/mocks/watch"
)

func TestHasPublishedIP_ImmediateResults(t *testing.T) {
	t.Parallel()

	createEndpointWithAddresses := func(name, resourceVersion string) *v2.Endpoints {
		return &v2.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				ResourceVersion: resourceVersion,
			},
			Subsets: []v2.EndpointSubset{
				{
					Addresses: []v2.EndpointAddress{
						{IP: "10.0.0.1"},
					},
				},
			},
		}
	}

	createEndpointWithoutAddresses := func(name, resourceVersion string) *v2.Endpoints {
		return &v2.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				ResourceVersion: resourceVersion,
			},
			Subsets: []v2.EndpointSubset{},
		}
	}

	tests := []struct {
		name          string
		serviceName   string
		endpoint      *v2.Endpoints
		getError      error
		expectedError string
		shouldWatch   bool
		watchChan     func() chan watch.Event
	}{
		{
			name:        "success - immediate success from GET - endpoint has addresses",
			serviceName: "test-service",
			endpoint:    createEndpointWithAddresses("test-service", "123"),
			getError:    nil,
			shouldWatch: false,
		},
		{
			name:          "failure - get endpoint error",
			serviceName:   "missing-service",
			endpoint:      nil,
			getError:      errors.New("endpoint not found"),
			expectedError: "failed to get endpoint for missing-service",
			shouldWatch:   false,
		},
		{
			name:        "success - endpoint without addresses - should watch",
			serviceName: "pending-service",
			endpoint:    createEndpointWithoutAddresses("pending-service", "123"),
			getError:    nil,
			shouldWatch: true,
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 1)
				watchChan <- watch.Event{
					Type: watch.Added,
					Object: &v2.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pending-service",
						},
						Subsets: []v2.EndpointSubset{
							{Addresses: []v2.EndpointAddress{
								{IP: "10.0.0.1"},
							}},
						},
					},
				}

				return watchChan
			},
		},
		{
			name:          "failure - watch returned wrong event",
			serviceName:   "pending-service",
			endpoint:      createEndpointWithoutAddresses("pending-service", "123"),
			getError:      nil,
			shouldWatch:   true,
			expectedError: "event from channel <some-wrong-name> does not match requested <pending-service>",
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 1)
				watchChan <- watch.Event{
					Type: watch.Added,
					Object: &v2.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-wrong-name",
						},
						Subsets: []v2.EndpointSubset{
							{Addresses: []v2.EndpointAddress{
								{IP: "10.0.0.1"},
							}},
						},
					},
				}
				return watchChan
			},
		},
		{
			name:        "success - watch - finds IP on second receive from channel",
			serviceName: "pending-service",
			endpoint:    createEndpointWithoutAddresses("pending-service", "123"),
			getError:    nil,
			shouldWatch: true,
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 2)
				// 1st msg has no IP
				watchChan <- watch.Event{
					Type: watch.Added,
					Object: &v2.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pending-service",
						},
					},
				}

				// 2nd msg has IP
				watchChan <- watch.Event{
					Type: watch.Added,
					Object: &v2.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pending-service",
						},
						Subsets: []v2.EndpointSubset{
							{Addresses: []v2.EndpointAddress{
								{IP: "10.0.0.1"},
							}},
						},
					},
				}
				return watchChan
			},
		},
		{
			name:        "success - watch - finds IP on second receive from channel, first event is error type",
			serviceName: "pending-service",
			endpoint:    createEndpointWithoutAddresses("pending-service", "123"),
			getError:    nil,
			shouldWatch: true,
			watchChan: func() chan watch.Event {
				watchChan := make(chan watch.Event, 2)
				// 1st msg has no IP
				watchChan <- watch.Event{
					Type: watch.Error,
					Object: &v2.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pending-service",
						},
					},
				}

				// 2nd msg has IP
				watchChan <- watch.Event{
					Type: watch.Added,
					Object: &v2.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pending-service",
						},
						Subsets: []v2.EndpointSubset{
							{Addresses: []v2.EndpointAddress{
								{IP: "10.0.0.1"},
							}},
						},
					},
				}
				return watchChan
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockK8sClient := mocks.NewMockInterface(ctrl)
			mockCoreV1 := mocks.NewMockCoreV1Interface(ctrl)
			mockEndpoints := mocks.NewMockEndpointsInterface(ctrl)

			mockK8sClient.EXPECT().CoreV1().Return(mockCoreV1)
			mockCoreV1.EXPECT().Endpoints("test-ns").Return(mockEndpoints)
			mockEndpoints.EXPECT().Get(gomock.Any(), tt.serviceName, metav1.GetOptions{}).Return(tt.endpoint, tt.getError)

			if tt.shouldWatch {
				mockK8sClient.EXPECT().CoreV1().Return(mockCoreV1)
				mockCoreV1.EXPECT().Endpoints("test-ns").Return(mockEndpoints)

				watcher := mocks2.NewMockInterface(ctrl)
				mockEndpoints.EXPECT().Watch(gomock.Any(), metav1.ListOptions{
					FieldSelector:   fmt.Sprintf("metadata.name=%s", tt.serviceName),
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
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
			} else {
				ctx = context.Background()
			}

			err := client.HasPublishedIP(ctx, tt.serviceName)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}
