package resources

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMakeRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                  string
		modelRoutes           []*Route
		pipelineRoutes        []*PipelineRoute
		expectedDefaultRoutes int
		expectedMirrorRoutes  int
	}

	tests := []test{
		{
			name: "one model",
			modelRoutes: []*Route{
				{
					RouteName: "r1",
					Clusters: []TrafficSplits{
						{
							ModelName:     "m1",
							ModelVersion:  1,
							TrafficWeight: 100,
							HttpCluster:   "h1",
							GrpcCluster:   "g1",
						},
					},
				},
			},
			expectedDefaultRoutes: 2,
			expectedMirrorRoutes:  0,
		},
		{
			name: "one pipeline",
			pipelineRoutes: []*PipelineRoute{
				{
					RouteName: "r1",
					Clusters: []PipelineTrafficSplits{
						{
							PipelineName:  "p1",
							TrafficWeight: 100,
						},
					},
				},
			},
			expectedDefaultRoutes: 2,
			expectedMirrorRoutes:  0,
		},
		{
			name: "pipeline experiment",
			pipelineRoutes: []*PipelineRoute{
				{
					RouteName: "r1",
					Clusters: []PipelineTrafficSplits{
						{
							PipelineName:  "p1",
							TrafficWeight: 50,
						},
						{
							PipelineName:  "p2",
							TrafficWeight: 50,
						},
					},
				},
			},
			expectedDefaultRoutes: 6,
			expectedMirrorRoutes:  0,
		},
		{
			name: "pipeline experiment with mirror",
			pipelineRoutes: []*PipelineRoute{
				{
					RouteName: "r1",
					Clusters: []PipelineTrafficSplits{
						{
							PipelineName:  "p1",
							TrafficWeight: 50,
						},
						{
							PipelineName:  "p2",
							TrafficWeight: 50,
						},
					},
					Mirrors: []PipelineTrafficSplits{
						{
							PipelineName:  "p3",
							TrafficWeight: 100,
						},
					},
				},
			},
			expectedDefaultRoutes: 6,
			expectedMirrorRoutes:  2,
		},
		{
			name: "model experiment",
			modelRoutes: []*Route{
				{
					RouteName: "r1",
					Clusters: []TrafficSplits{
						{
							ModelName:     "m1",
							ModelVersion:  1,
							TrafficWeight: 50,
							HttpCluster:   "h1",
							GrpcCluster:   "g1",
						},
						{
							ModelName:     "m2",
							ModelVersion:  1,
							TrafficWeight: 50,
							HttpCluster:   "h1",
							GrpcCluster:   "g1",
						},
					},
				},
			},
			expectedDefaultRoutes: 6,
			expectedMirrorRoutes:  0,
		},
		{
			name: "experiment with model mirror",
			modelRoutes: []*Route{
				{
					RouteName: "r1",
					Clusters: []TrafficSplits{
						{
							ModelName:     "m1",
							ModelVersion:  1,
							TrafficWeight: 50,
							HttpCluster:   "h1",
							GrpcCluster:   "g1",
						},
						{
							ModelName:     "m2",
							ModelVersion:  1,
							TrafficWeight: 50,
							HttpCluster:   "h1",
							GrpcCluster:   "g1",
						},
					},
					Mirrors: []TrafficSplits{
						{
							ModelName:     "m3",
							ModelVersion:  1,
							TrafficWeight: 100,
							HttpCluster:   "h1",
							GrpcCluster:   "g1",
						},
					},
				},
			},
			expectedDefaultRoutes: 6,
			expectedMirrorRoutes:  2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rcDef, rcMirror := MakeRoute(test.modelRoutes, test.pipelineRoutes)
			g.Expect(len(rcDef.VirtualHosts[0].Routes)).To(Equal(test.expectedDefaultRoutes))
			g.Expect(len(rcMirror.VirtualHosts[0].Routes)).To(Equal(test.expectedMirrorRoutes))
		})
	}
}

func TestGetRouteName(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name       string
		routeName  string
		isPipeline bool
		isGrpc     bool
		isMirror   bool
		expected   string
	}

	tests := []test{
		{
			name:       "http model",
			routeName:  "foo",
			isPipeline: false,
			isGrpc:     false,
			isMirror:   false,
			expected:   "foo_http",
		},
		{
			name:       "grpc model",
			routeName:  "foo",
			isPipeline: false,
			isGrpc:     true,
			isMirror:   false,
			expected:   "foo_grpc",
		},
		{
			name:       "http model mirror",
			routeName:  "foo",
			isPipeline: false,
			isGrpc:     false,
			isMirror:   true,
			expected:   "foo_http_mirror",
		},
		{
			name:       "grpc model mirror",
			routeName:  "foo",
			isPipeline: false,
			isGrpc:     true,
			isMirror:   true,
			expected:   "foo_grpc_mirror",
		},
		{
			name:       "http pipeline",
			routeName:  "foo",
			isPipeline: true,
			isGrpc:     false,
			isMirror:   false,
			expected:   "foo_pipeline_http",
		},
		{
			name:       "grpc pipeline",
			routeName:  "foo",
			isPipeline: true,
			isGrpc:     true,
			isMirror:   false,
			expected:   "foo_pipeline_grpc",
		},
		{
			name:       "http pipeline mirror",
			routeName:  "foo",
			isPipeline: true,
			isGrpc:     false,
			isMirror:   true,
			expected:   "foo_pipeline_http_mirror",
		},
		{
			name:       "grpc pipeline mirror",
			routeName:  "foo",
			isPipeline: true,
			isGrpc:     true,
			isMirror:   true,
			expected:   "foo_pipeline_grpc_mirror",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routeName := getRouteName(test.routeName, test.isPipeline, test.isGrpc, test.isMirror)
			g.Expect(routeName).To(Equal(test.expected))
		})
	}
}
