/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package hodometer

import (
	"context"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/version"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestUpdatePipelineMetrics(t *testing.T) {
	type test struct {
		name     string
		statuses []*scheduler.PipelineStatusResponse
		expected pipelineMetrics
	}

	tests := []test{
		{
			name:     "no metrics",
			statuses: []*scheduler.PipelineStatusResponse{},
			expected: pipelineMetrics{},
		},
		{
			name: "to-create, creating, and ready pipelines count",
			statuses: []*scheduler.PipelineStatusResponse{
				{
					Versions: []*scheduler.PipelineWithState{
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineCreate},
						},
					},
				},
				{
					Versions: []*scheduler.PipelineWithState{
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineCreating},
						},
					},
				},
				{
					Versions: []*scheduler.PipelineWithState{
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineReady},
						},
					},
				},
			},
			expected: pipelineMetrics{count: 3},
		},
		{
			name: "other statuses do not count",
			statuses: []*scheduler.PipelineStatusResponse{
				{
					Versions: []*scheduler.PipelineWithState{
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineStatusUnknown},
						},
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineFailed},
						},
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineTerminate},
						},
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineFailed},
						},
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineTerminate},
						},
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineTerminating},
						},
						{
							State: &scheduler.PipelineVersionState{Status: scheduler.PipelineVersionState_PipelineTerminated},
						},
					},
				},
			},
			expected: pipelineMetrics{count: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := pipelineMetrics{}
			for _, s := range tt.statuses {
				updatePipelineMetrics(&metrics, s)
			}
			require.Equal(t, tt.expected, metrics)
		})
	}
}

func TestUpdateServerMetrics(t *testing.T) {
	type test struct {
		name     string
		statuses []*scheduler.ServerStatusResponse
		expected serverMetrics
	}

	tests := []test{
		{
			name:     "no metrics",
			statuses: []*scheduler.ServerStatusResponse{},
			expected: serverMetrics{},
		},
		{
			name: "expected replicas count",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas:  5,
					AvailableReplicas: 0,
				},
			},
			expected: serverMetrics{
				count:        1,
				replicaCount: 5,
			},
		},
		{
			name: "available replica counts even when expected unknown",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas:  -1,
					AvailableReplicas: 5,
				},
			},
			expected: serverMetrics{
				count:        1,
				replicaCount: 5,
			},
		},
		{
			name: "expected or available count",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas:  3,
					AvailableReplicas: 1,
				},
			},
			expected: serverMetrics{
				count:        1,
				replicaCount: 3,
			},
		},
		{
			name: "many servers",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas:  1,
					AvailableReplicas: 0,
				},
				{
					ExpectedReplicas:  -1,
					AvailableReplicas: 1,
				},
			},
			expected: serverMetrics{
				count:        2,
				replicaCount: 2,
			},
		},
		{
			name: "overcommit enabled but no models counts",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*scheduler.ServerReplicaResources{
						{
							OverCommitPercentage: 20,
						},
					},
				},
			},
			expected: serverMetrics{
				count:             1,
				replicaCount:      1,
				multimodelEnabled: 1,
				overcommitEnabled: 1,
			},
		},
		{
			name: "overcommit enabled with allocated models",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*scheduler.ServerReplicaResources{
						{
							OverCommitPercentage: 20,
							NumLoadedModels:      1,
						},
					},
				},
			},
			expected: serverMetrics{
				count:             1,
				replicaCount:      1,
				multimodelEnabled: 1,
				overcommitEnabled: 1,
			},
		},
		{
			name: "multi-model without overcommit",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*scheduler.ServerReplicaResources{
						{
							NumLoadedModels: 10,
						},
					},
				},
			},
			expected: serverMetrics{
				count:             1,
				replicaCount:      1,
				multimodelEnabled: 1,
			},
		},
		{
			name: "memory consumption is totalled correctly",
			statuses: []*scheduler.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*scheduler.ServerReplicaResources{
						{
							TotalMemoryBytes: 1,
						},
					},
				},
				{
					ExpectedReplicas: 1,
					Resources: []*scheduler.ServerReplicaResources{
						{
							TotalMemoryBytes: 2,
						},
					},
				},
			},
			expected: serverMetrics{
				count:        2,
				replicaCount: 2,
				memoryBytes:  uint(3),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := serverMetrics{}
			for _, s := range tt.statuses {
				updateServerMetrics(&metrics, s)
			}
			require.Equal(t, tt.expected, metrics)
		})
	}
}

func TestCollectKubernetes(t *testing.T) {
	type test struct {
		name          string
		hasKubeConfig bool
		expected      *kubernetesMetrics
	}

	tests := []test{
		// Unable to easily test case of server version being unavailable,
		// due to how the FakeDiscovery object in the fake client is set up.
		{
			name:          "Only able to retrieve server version",
			hasKubeConfig: true,
			expected:      &kubernetesMetrics{version: "v1.23.4"},
		},
		{
			name:          "should not return values when no kube config available",
			hasKubeConfig: false,
			expected:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			logger := logrus.New()
			logger.Out = io.Discard
			ctx := context.Background()
			var client *fake.Clientset

			// When
			if tt.hasKubeConfig {
				client = fake.NewSimpleClientset()

				d := client.Discovery().(*fakeDiscovery.FakeDiscovery)
				d.FakedServerVersion = &version.Info{GitVersion: "v1.23.4"}
			}

			// Then
			scc := SeldonCoreCollector{
				k8sClient: client,
				logger:    logger,
			}
			metrics := scc.collectKubernetes(ctx)
			require.Equal(t, tt.expected, metrics)
		})
	}
}
