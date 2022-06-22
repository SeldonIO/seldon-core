package hodometer

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"

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
		numNodes      int
		shouldError   bool
		expected      *kubernetesMetrics
	}

	tests := []test{
		// Unable to easily test case of server version being unavailable,
		// due to how the FakeDiscovery object in the fake client is set up.
		{
			name:          "Only able to retrieve server version",
			hasKubeConfig: true,
			numNodes:      0,
			shouldError:   false,
			expected:      &kubernetesMetrics{version: "v1.23.4"},
		},
		{
			name:          "Fewer nodes than max per call",
			hasKubeConfig: true,
			numNodes:      maxNodesPerCall - 1,
			shouldError:   false,
			expected:      &kubernetesMetrics{version: "v1.23.4", nodeCount: maxNodesPerCall - 1},
		},
		{
			name:          "More nodes than max per call",
			hasKubeConfig: true,
			numNodes:      maxNodesPerCall + 1,
			shouldError:   false,
			expected:      &kubernetesMetrics{version: "v1.23.4", nodeCount: maxNodesPerCall + 1},
		},
		{
			name:          "Nodes should be zero on error",
			hasKubeConfig: true,
			numNodes:      1,
			shouldError:   true,
			expected:      &kubernetesMetrics{version: "v1.23.4", nodeCount: 0},
		},
		{
			name:          "should not return values when no kube config available",
			hasKubeConfig: false,
			numNodes:      0,
			shouldError:   false,
			expected:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.GreaterOrEqual(t, tt.numNodes, 0)

			// Given
			logger := logrus.New()
			logger.Out = ioutil.Discard
			ctx := context.Background()
			var client *fake.Clientset

			// When
			if tt.hasKubeConfig {
				client = fake.NewSimpleClientset()

				d := client.Discovery().(*fakeDiscovery.FakeDiscovery)
				d.FakedServerVersion = &version.Info{GitVersion: "v1.23.4"}

				switch {
				case tt.shouldError:
					addErrorOnListNodes(client, "unable to fetch nodes")
				case tt.numNodes > 0:
					addListNodes(client, tt.numNodes)
				}
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

func addErrorOnListNodes(client *fake.Clientset, msg string) {
	client.PrependReactor(
		"list",
		"nodes",
		func(action k8sTesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New(msg)
		},
	)
}

func addListNodes(client *fake.Clientset, numNodes int) {
	returnedNodes := 0
	client.PrependReactor(
		"list",
		"nodes",
		func(action k8sTesting.Action) (bool, runtime.Object, error) {
			nodes := &coreV1.NodeList{}

			for i := 0; i < maxNodesPerCall && returnedNodes < numNodes; i++ {
				returnedNodes++
				nodes.Items = append(nodes.Items, coreV1.Node{})
			}

			if returnedNodes < numNodes {
				nodes.Continue = "ongoing"
			}

			return true, nodes, nil
		},
	)
}
