package xdscache

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

// Checks bad use with only http cluster added
func TestHttpRouteOnlyAdded(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	c := NewSeldonXDSCache(logger)

	httpCluster := "http1"
	r1 := "r1"
	r2 := "r2"
	c.AddRoute(r1, "m1", "http1", "grpc1", false, 100, 1)
	c.AddCluster(httpCluster, r1, false)
	c.AddEndpoint("http1", "0.0.0.0", 9000)
	c.AddRoute(r2, "m2", "http1", "grpc1", false, 100, 1)
	c.AddCluster("http1", r2, false)
	c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
	err := c.RemoveRoutes(r1)
	g.Expect(err).ToNot(BeNil())
}

// Checks a cluster remains until all routes are removed
func TestAddRemoveHttpAndGrpcRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	c := NewSeldonXDSCache(logger)

	addVersionedRoute := func(c *SeldonXDSCache, httpCluster, grpcCluster, route string, traffic uint32, version uint32) {
		c.AddRoute(route, "m1", httpCluster, grpcCluster, false, traffic, version)
		c.AddCluster(httpCluster, route, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddCluster(grpcCluster, route, true)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	httpCluster := "http1"
	grpcCluster := "grpc1"
	r1 := "r1"
	r2 := "r2"
	addVersionedRoute(c, httpCluster, grpcCluster, r1, 50, 1)
	addVersionedRoute(c, httpCluster, grpcCluster, r2, 100, 1)

	err := c.RemoveRoutes(r1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpCluster]
	g.Expect(ok).To(BeTrue()) // http Cluster remains as r2 still connected
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeTrue()) // grpc Cluster remains as r2 still connected
	err = c.RemoveRoutes(r2)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
}

// Checks a cluster remains until all routes and versions are removed
func TestAddRemoveHttpAndGrpcRouteVersions(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	addVersionedRoute := func(c *SeldonXDSCache, httpCluster, grpcCluster, route string, traffic uint32, version uint32) {
		c.AddRoute(route, "m1", httpCluster, grpcCluster, false, traffic, version)
		c.AddCluster(httpCluster, route, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddCluster(grpcCluster, route, true)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger)

	httpCluster := "http1"
	grpcCluster := "grpc1"
	r1 := "r1"
	r2 := "r2"
	addVersionedRoute(c, httpCluster, grpcCluster, r1, 50, 1)
	addVersionedRoute(c, httpCluster, grpcCluster, r1, 100, 2)
	addVersionedRoute(c, httpCluster, grpcCluster, r2, 100, 1)
	err := c.RemoveRoutes(r1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpCluster]
	g.Expect(ok).To(BeTrue()) // http Cluster remains as r2 still connected
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeTrue()) // grpc Cluster remains as r2 still connected
	err = c.RemoveRoutes(r2)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
}
