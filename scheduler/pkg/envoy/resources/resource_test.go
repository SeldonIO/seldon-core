/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package resources

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMakeRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                  string
		modelRoutes           map[string]Route
		pipelineRoutes        map[string]PipelineRoute
		expectedDefaultRoutes int
		expectedMirrorRoutes  int
	}

	tests := []test{
		{
			name: "one model",
			modelRoutes: map[string]Route{"r1": Route{
				RouteName: "r1",
				Clusters: []TrafficSplit{
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
			pipelineRoutes: map[string]PipelineRoute{"r1": PipelineRoute{
				RouteName: "r1",
				Clusters: []PipelineTrafficSplit{
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
			pipelineRoutes: map[string]PipelineRoute{"r1": PipelineRoute{
				RouteName: "r1",
				Clusters: []PipelineTrafficSplit{
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
			pipelineRoutes: map[string]PipelineRoute{"r1": PipelineRoute{
				RouteName: "r1",
				Clusters: []PipelineTrafficSplit{
					{
						PipelineName:  "p1",
						TrafficWeight: 50,
					},
					{
						PipelineName:  "p2",
						TrafficWeight: 50,
					},
				},
				Mirror: &PipelineTrafficSplit{
					PipelineName:  "p3",
					TrafficWeight: 100,
				},
			},
			},
			expectedDefaultRoutes: 6,
			expectedMirrorRoutes:  2,
		},
		{
			name: "model experiment",
			modelRoutes: map[string]Route{"r1": Route{
				RouteName: "r1",
				Clusters: []TrafficSplit{
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
			modelRoutes: map[string]Route{"r1": Route{
				RouteName: "r1",
				Clusters: []TrafficSplit{
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
				Mirror: &TrafficSplit{
					ModelName:     "m3",
					ModelVersion:  1,
					TrafficWeight: 100,
					HttpCluster:   "h1",
					GrpcCluster:   "g1",
				},
			},
			},
			expectedDefaultRoutes: 6,
			expectedMirrorRoutes:  2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rcDef, rcMirror := MakeRoutes(test.modelRoutes, test.pipelineRoutes)
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
