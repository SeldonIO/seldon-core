package xdscache

import (
	"testing"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

// Checks a cluster remains until all routes are removed
func TestAddRemoveHttpAndGrpcRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	c := NewSeldonXDSCache(logger)

	addVersionedRoute := func(c *SeldonXDSCache, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, modelName, version, false)
		c.AddCluster(grpcCluster, modelName, version, true)
		c.AddRouteClusterTraffic(modelName, modelName, version, traffic, httpCluster, grpcCluster, true)
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
		c.AddCluster(httpCluster, modelName, version, false)
		c.AddCluster(grpcCluster, modelName, version, true)
		c.AddRouteClusterTraffic(modelName, modelName, version, traffic, httpCluster, grpcCluster, true)
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
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 2}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 2}]).To(BeTrue())

	addVersionedRoute(c, model2, httpCluster, grpcCluster, 100, 1)

	// check what we added
	g.Expect(len(c.Routes[model2].Clusters)).To(Equal(1))
	clusters = c.Routes[model2].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(100)))
	g.Expect(c.Clusters[httpCluster].Routes[resources.ModelVersionKey{Name: model2, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.ModelVersionKey{Name: model2, Version: 1}]).To(BeTrue())
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

// Checks a cluster with multiple versions is created correctly
func TestAddRemoveHttpAndGrpcRouteVersionsForSameModel(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	addVersionedRoute := func(c *SeldonXDSCache, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, modelName, version, false)
		c.AddCluster(grpcCluster, modelName, version, true)
		c.AddRouteClusterTraffic(modelName, modelName, version, traffic, httpCluster, grpcCluster, true)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger)

	httpCluster := "http1"
	grpcCluster := "grpc1"
	model1 := "m1"
	addVersionedRoute(c, model1, httpCluster, grpcCluster, 40, 1)
	addVersionedRoute(c, model1, httpCluster, grpcCluster, 60, 2)

	// check what we have added
	g.Expect(len(c.Routes[model1].Clusters)).To(Equal(2))
	clusters := c.Routes[model1].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 2}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.ModelVersionKey{Name: model1, Version: 2}]).To(BeTrue())

	err := c.RemoveRoute(model1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpCluster]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
}

// Checks a cluster with multiple versions is created correctly
func TestAddRemoveHttpAndGrpcRouteVersionsForDifferentModels(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	addVersionedRoute := func(c *SeldonXDSCache, modelRouteName string, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, modelName, version, false)
		c.AddCluster(grpcCluster, modelName, version, true)
		c.AddRouteClusterTraffic(modelRouteName, modelName, version, traffic, httpCluster, grpcCluster, true)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger)

	httpClusterModel1 := "model1_http1"
	grpcClusterModel1 := "model1_grpc1"
	httpClusterModel2 := "model2_http1"
	grpcClusterModel2 := "model2_grpc1"
	model1 := "m1"
	model2 := "m2"
	addVersionedRoute(c, model1, model1, httpClusterModel1, grpcClusterModel1, 40, 1)
	addVersionedRoute(c, model1, model2, httpClusterModel2, grpcClusterModel2, 60, 1)

	// check what we have added
	g.Expect(len(c.Routes[model1].Clusters)).To(Equal(2))
	clusters := c.Routes[model1].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpClusterModel1].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcClusterModel1].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel1].Routes[resources.ModelVersionKey{Name: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[resources.ModelVersionKey{Name: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel2].Routes[resources.ModelVersionKey{Name: model2, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel2].Routes[resources.ModelVersionKey{Name: model2, Version: 1}]).To(BeTrue())

	err := c.RemoveRoute(model1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpClusterModel1]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcClusterModel1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
	_, ok = c.Clusters[httpClusterModel2]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcClusterModel2]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
}
