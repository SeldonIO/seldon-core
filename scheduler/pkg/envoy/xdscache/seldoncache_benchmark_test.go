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
		x.AddPipelineRoute(strconv.Itoa(n))
		x.AddRouteClusterTraffic(
			fmt.Sprintf("model-%d", n),
			fmt.Sprintf("model-%d", n),
			1,
			100,
			"http",
			"grpc",
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
