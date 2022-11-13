/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package xdscache

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sirupsen/logrus"
)

// Prevent compiler from optimising away benchmarks
var results []types.Resource

func benchmarkRouteContents(b *testing.B, numResources uint) {
	x := NewSeldonXDSCache(logrus.New(), nil)

	for n := 0; n < int(numResources); n++ {
		x.AddPipelineRoute(strconv.Itoa(n), strconv.Itoa(n), 100, false)
		x.AddRouteClusterTraffic(
			fmt.Sprintf("model-%d", n),
			fmt.Sprintf("model-%d", n),
			1,
			100,
			"http",
			"grpc",
			false,
			false,
		)
	}

	// Prevent compiler optimising away function calls
	var r []types.Resource
	for i := 0; i < b.N; i++ {
		r = x.RouteContents()
	}
	results = r
}

func BenchmarkRouteContents100(b *testing.B) { benchmarkRouteContents(b, 100) }
func BenchmarkRouteContents1K(b *testing.B)  { benchmarkRouteContents(b, 1_000) }
func BenchmarkRouteContents10K(b *testing.B) { benchmarkRouteContents(b, 10_000) }
