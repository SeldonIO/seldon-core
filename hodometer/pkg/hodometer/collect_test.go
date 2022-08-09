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

	pb "github.com/seldonio/seldon-core/hodometer/apis"
)

func TestUpdatePipelineMetrics(t *testing.T) {
	type test struct {
		name     string
		statuses []*pb.PipelineStatusResponse
		expected pipelineMetrics
	}

	tests := []test{
		{
			name:     "no metrics",
			statuses: []*pb.PipelineStatusResponse{},
			expected: pipelineMetrics{},
		},
		{
			name: "to-create, creating, and ready pipelines count",
			statuses: []*pb.PipelineStatusResponse{
				{
					Versions: []*pb.PipelineWithState{
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineCreate},
						},
					},
				},
				{
					Versions: []*pb.PipelineWithState{
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineCreating},
						},
					},
				},
				{
					Versions: []*pb.PipelineWithState{
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineReady},
						},
					},
				},
			},
			expected: pipelineMetrics{count: 3},
		},
		{
			name: "other statuses do not count",
			statuses: []*pb.PipelineStatusResponse{
				{
					Versions: []*pb.PipelineWithState{
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineStatusUnknown},
						},
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineFailed},
						},
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineTerminate},
						},
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineFailed},
						},
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineTerminate},
						},
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineTerminating},
						},
						{
							State: &pb.PipelineVersionState{Status: pb.PipelineVersionState_PipelineTerminated},
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
		statuses []*pb.ServerStatusResponse
		expected serverMetrics
	}

	tests := []test{
		{
			name:     "no metrics",
			statuses: []*pb.ServerStatusResponse{},
			expected: serverMetrics{},
		},
		{
			name: "expected replicas count",
			statuses: []*pb.ServerStatusResponse{
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
			statuses: []*pb.ServerStatusResponse{
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
			statuses: []*pb.ServerStatusResponse{
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
			statuses: []*pb.ServerStatusResponse{
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
			statuses: []*pb.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*pb.ServerReplicaResources{
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
			statuses: []*pb.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*pb.ServerReplicaResources{
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
			statuses: []*pb.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*pb.ServerReplicaResources{
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
			statuses: []*pb.ServerStatusResponse{
				{
					ExpectedReplicas: 1,
					Resources: []*pb.ServerReplicaResources{
						{
							TotalMemoryBytes: 1,
						},
					},
				},
				{
					ExpectedReplicas: 1,
					Resources: []*pb.ServerReplicaResources{
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
