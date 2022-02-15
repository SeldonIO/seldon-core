package xdscache

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

// Checks a cluster remains until all routes are removed
func TestAddRemoveHttpAndGrpcRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	c := NewSeldonXDSCache(logger)

	addVersionedRoute := func(c *SeldonXDSCache, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, modelName, false)
		c.AddCluster(grpcCluster, modelName, true)
		c.AddRouteClusterTraffic(modelName, version, traffic, httpCluster, grpcCluster, true)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	httpCluster := "http1"
	grpcCluster := "grpc1"
	model1 := "m1"
	model2 := "m2"
	addVersionedRoute(c, model1, httpCluster, grpcCluster, 100, 1)
	addVersionedRoute(c, model2, httpCluster, grpcCluster, 100, 1)

	err := c.RemoveRoute(model1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpCluster]
	g.Expect(ok).To(BeTrue()) // http Cluster remains as r2 still connected
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeTrue()) // grpc Cluster remains as r2 still connected
	err = c.RemoveRoute(model2)
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

	addVersionedRoute := func(c *SeldonXDSCache, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, modelName, false)
		c.AddCluster(grpcCluster, modelName, true)
		c.AddRouteClusterTraffic(modelName, version, traffic, httpCluster, grpcCluster, true)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger)

	httpCluster := "http1"
	grpcCluster := "grpc1"
	model1 := "m1"
	model2 := "m2"
	addVersionedRoute(c, model1, httpCluster, grpcCluster, 40, 1)
	addVersionedRoute(c, model1, httpCluster, grpcCluster, 60, 2)

	// check what we have added
	g.Expect(len(c.Routes[model1].Clusters)).To(Equal(2))
	clusters := c.Routes[model1].Clusters
	g.Expect(clusters[0].TrafficPercent).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficPercent).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[model1]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[model1]).To(BeTrue())

	addVersionedRoute(c, model2, httpCluster, grpcCluster, 100, 1)

	// check what we added
	g.Expect(len(c.Routes[model2].Clusters)).To(Equal(1))
	clusters = c.Routes[model2].Clusters
	g.Expect(clusters[0].TrafficPercent).To(Equal(uint32(100)))
	g.Expect(c.Clusters[httpCluster].Routes[model2]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[model2]).To(BeTrue())
	g.Expect(len(c.Clusters[httpCluster].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster].Endpoints)).To(Equal(1))

	err := c.RemoveRoute(model1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpCluster]
	g.Expect(ok).To(BeTrue()) // http Cluster remains as r2 still connected
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeTrue()) // grpc Cluster remains as r2 still connected
	err = c.RemoveRoute(model2)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
}
