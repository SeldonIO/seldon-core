/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

// Checks a cluster remains until all routes are removed
func TestAddRemoveHttpAndGrpcRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	c, err := NewSeldonXDSCache(logger, &PipelineGatewayDetails{}, nil)
	g.Expect(err).To(BeNil())
	httpCluster := "m1_1_http"
	grpcCluster := "m1_1_grpc"
	model1 := "m1"
	route1 := "r1"
	route2 := "r2"

	addVersionedRoute(c, route1, model1, httpCluster, grpcCluster, 100, 1)
	addVersionedRoute(c, route2, model1, httpCluster, grpcCluster, 100, 1)

	_, ok := c.clustersToAdd[httpCluster]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcCluster]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds

	err = c.RemoveRoute(route1)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster]
	g.Expect(ok).To(BeTrue()) // http Cluster remains as r2 still connected
	_, ok = c.Clusters[grpcCluster]
	g.Expect(ok).To(BeTrue()) // grpc Cluster remains as r2 still connected
	err = c.RemoveRoute(route2)
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

	c, err := NewSeldonXDSCache(logger, &PipelineGatewayDetails{}, nil)
	g.Expect(err).To(BeNil())

	httpCluster1 := "m1_1_http"
	grpcCluster1 := "m1_1_grpc"
	httpCluster2 := "m1_2_http"
	grpcCluster2 := "m1_2_grpc"
	model1 := "m1"
	route1 := "r1"
	route2 := "r2"

	addVersionedRoute(c, route1, model1, httpCluster1, grpcCluster1, 40, 1)
	addVersionedRoute(c, route1, model1, httpCluster2, grpcCluster2, 60, 2)

	_, ok := c.clustersToAdd[httpCluster1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcCluster1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds
	_, ok = c.clustersToAdd[httpCluster2]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcCluster2]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds

	// check what we have added
	g.Expect(len(c.Routes[route1].Clusters)).To(Equal(2))
	clusters := c.Routes[route1].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster1].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster1].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster1].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster1].Routes[RouteVersionKey{RouteName: route1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster1].Routes[RouteVersionKey{RouteName: route1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpCluster2].Routes[RouteVersionKey{RouteName: route1, ModelName: model1, Version: 2}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster2].Routes[RouteVersionKey{RouteName: route1, ModelName: model1, Version: 2}]).To(BeTrue())

	addVersionedRoute(c, route2, model1, httpCluster1, grpcCluster1, 100, 1)

	// check what we added
	g.Expect(len(c.Routes[route2].Clusters)).To(Equal(1))
	clusters = c.Routes[route2].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(100)))
	g.Expect(c.Clusters[httpCluster1].Routes[RouteVersionKey{RouteName: route2, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster1].Routes[RouteVersionKey{RouteName: route2, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(len(c.Clusters[httpCluster1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster1].Endpoints)).To(Equal(1))

	err = c.RemoveRoute(route1)
	g.Expect(err).To(BeNil())
	err = c.RemoveClusters() // remove clusters to clear clustersToRemove
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster1]
	g.Expect(ok).To(BeTrue()) // http Cluster remains as r2 still connected
	_, ok = c.Clusters[grpcCluster1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster remains as r2 still connected
	_, ok = c.clustersToRemove[httpCluster1]
	g.Expect(ok).To(BeFalse()) // http Cluster not to be removed from cds
	_, ok = c.clustersToRemove[grpcCluster1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster not to be removed from cds
	ok = c.shouldRemoveCluster(httpCluster1)
	g.Expect(ok).To(BeFalse()) // http Cluster not to be removed from cds
	ok = c.shouldRemoveCluster(grpcCluster1)
	g.Expect(ok).To(BeFalse()) // grpc Cluster not to be removed from cds
	err = c.RemoveRoute(route2)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster1]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcCluster1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
	_, ok = c.clustersToRemove[httpCluster1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	_, ok = c.clustersToRemove[grpcCluster1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
	ok = c.shouldRemoveCluster(httpCluster1)
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	ok = c.shouldRemoveCluster(grpcCluster1)
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
}

// Checks a cluster with multiple versions is created correctly
func TestAddRemoveHttpAndGrpcRouteVersionsForSameModel(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	c, err := NewSeldonXDSCache(logger, &PipelineGatewayDetails{}, nil)
	g.Expect(err).To(BeNil())

	routeName := "r1"
	httpCluster1 := "m1_1_http"
	grpcCluster1 := "m1_1_grpc"
	httpCluster2 := "m1_2_http"
	grpcCluster2 := "m1_2_grpc"
	model1 := "m1"

	addVersionedRoute(c, routeName, model1, httpCluster1, grpcCluster1, 40, 1)
	addVersionedRoute(c, routeName, model1, httpCluster2, grpcCluster2, 60, 2)

	_, ok := c.clustersToAdd[httpCluster1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcCluster1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds
	_, ok = c.clustersToAdd[httpCluster2]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcCluster2]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds

	// check what we have added
	g.Expect(len(c.Routes[routeName].Clusters)).To(Equal(2))
	clusters := c.Routes[routeName].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster1].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster1].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster1].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster1].Routes[RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster1].Routes[RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpCluster2].Routes[RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 2}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster2].Routes[RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 2}]).To(BeTrue())

	err = c.RemoveRoute(routeName)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpCluster1]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcCluster1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
	_, ok = c.clustersToRemove[httpCluster1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	_, ok = c.clustersToRemove[grpcCluster1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
	ok = c.shouldRemoveCluster(httpCluster1)
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	ok = c.shouldRemoveCluster(grpcCluster1)
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
}

// Checks a cluster with multiple versions is created correctly
func TestAddRemoveHttpAndGrpcRouteVersionsForDifferentModels(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	c, err := NewSeldonXDSCache(logger, &PipelineGatewayDetails{}, nil)
	g.Expect(err).To(BeNil())

	httpClusterModel1 := "m1_1_http"
	grpcClusterModel1 := "m1_1_grpc"
	httpClusterModel2 := "m2_1_http"
	grpcClusterModel2 := "m2_1_grpc"
	model1 := "m1"
	model2 := "m2"
	routeName := "r1"

	addVersionedRoute(c, routeName, model1, httpClusterModel1, grpcClusterModel1, 40, 1)
	addVersionedRoute(c, routeName, model2, httpClusterModel2, grpcClusterModel2, 60, 1)

	_, ok := c.clustersToAdd[httpClusterModel1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcClusterModel1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds
	_, ok = c.clustersToAdd[httpClusterModel2]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcClusterModel2]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds

	// check what we have added
	g.Expect(len(c.Routes[routeName].Clusters)).To(Equal(2))
	clusters := c.Routes[routeName].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpClusterModel1].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcClusterModel1].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel1].Routes[RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel2].Routes[RouteVersionKey{RouteName: routeName, ModelName: model2, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel2].Routes[RouteVersionKey{RouteName: routeName, ModelName: model2, Version: 1}]).To(BeTrue())

	err = c.RemoveRoute(routeName)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpClusterModel1]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcClusterModel1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
	_, ok = c.Clusters[httpClusterModel2]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcClusterModel2]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
	_, ok = c.clustersToRemove[httpClusterModel1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	_, ok = c.clustersToRemove[grpcClusterModel1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
	ok = c.shouldRemoveCluster(httpClusterModel1)
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	ok = c.shouldRemoveCluster(grpcClusterModel1)
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
	_, ok = c.clustersToRemove[httpClusterModel2]
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	_, ok = c.clustersToRemove[grpcClusterModel2]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
	ok = c.shouldRemoveCluster(httpClusterModel2)
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	ok = c.shouldRemoveCluster(grpcClusterModel2)
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
}

func TestAddRemoveHttpAndGrpcRouteVersionsForDifferentRoutesSameModel(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	c, err := NewSeldonXDSCache(logger, &PipelineGatewayDetails{}, nil)
	g.Expect(err).To(BeNil())

	route1 := "r1"
	route2 := "r2"
	httpClusterModel1 := "m1_1_http"
	grpcClusterModel1 := "m1_1_grpc"
	model1 := "m1"

	addVersionedRoute(c, route1, model1, httpClusterModel1, grpcClusterModel1, 100, 1)
	addVersionedRoute(c, route2, model1, httpClusterModel1, grpcClusterModel1, 100, 1)

	_, ok := c.clustersToAdd[httpClusterModel1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be added to cds
	_, ok = c.clustersToAdd[grpcClusterModel1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be added to cds

	// check what we have added
	g.Expect(len(c.Routes[route1].Clusters)).To(Equal(1))
	g.Expect(len(c.Routes[route2].Clusters)).To(Equal(1))
	clusters := c.Routes[route1].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(100)))
	g.Expect(len(c.Clusters[httpClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpClusterModel1].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcClusterModel1].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel1].Routes[RouteVersionKey{RouteName: route1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[RouteVersionKey{RouteName: route1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel1].Routes[RouteVersionKey{RouteName: route2, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[RouteVersionKey{RouteName: route2, ModelName: model1, Version: 1}]).To(BeTrue())

	err = c.RemoveRoute(route1)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpClusterModel1]
	g.Expect(ok).To(BeTrue()) // http Cluster not removed
	_, ok = c.Clusters[grpcClusterModel1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster not removed
	ok = c.shouldRemoveCluster(httpClusterModel1)
	g.Expect(ok).To(BeFalse()) // http Cluster not to be removed from cds
	ok = c.shouldRemoveCluster(grpcClusterModel1)
	g.Expect(ok).To(BeFalse()) // grpc Cluster not to be removed from cds
	err = c.RemoveRoute(route2)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpClusterModel1]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcClusterModel1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
	_, ok = c.clustersToRemove[httpClusterModel1]
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	_, ok = c.clustersToRemove[grpcClusterModel1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
	ok = c.shouldRemoveCluster(httpClusterModel1)
	g.Expect(ok).To(BeTrue()) // http Cluster to be removed from cds
	ok = c.shouldRemoveCluster(grpcClusterModel1)
	g.Expect(ok).To(BeTrue()) // grpc Cluster to be removed from cds
}

func TestSetupTLS(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	type test struct {
		name         string
		setTLS       bool
		envsPrefixes []string
		err          bool
	}

	tests := []test{
		{
			name:         "no tls",
			setTLS:       false,
			envsPrefixes: []string{},
			err:          false,
		},
		{
			name:         "only downstream server",
			setTLS:       true,
			envsPrefixes: []string{seldontls.EnvSecurityPrefixEnvoyDownstreamServer},
			err:          true,
		},
		{
			name:         "tls ok",
			setTLS:       true,
			envsPrefixes: []string{seldontls.EnvSecurityPrefixEnvoyUpstreamServer, seldontls.EnvSecurityPrefixEnvoyDownstreamServer, seldontls.EnvSecurityPrefixEnvoyUpstreamClient},
			err:          false,
		},
		{
			name:         "tls ok with downstream mtls",
			setTLS:       true,
			envsPrefixes: []string{seldontls.EnvSecurityPrefixEnvoyUpstreamServer, seldontls.EnvSecurityPrefixEnvoyDownstreamServer, seldontls.EnvSecurityPrefixEnvoyUpstreamClient, seldontls.EnvSecurityPrefixEnvoyDownstreamClient},
			err:          false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := NewSeldonXDSCache(logger, &PipelineGatewayDetails{}, nil)
			g.Expect(err).To(BeNil())
			if test.setTLS {
				t.Setenv(fmt.Sprintf("%s%s", seldontls.EnvSecurityPrefixEnvoy, seldontls.EnvSecurityProtocolSuffix), seldontls.SecurityProtocolSSL)
			}
			// Setup envs for cert for each prefix
			for _, envPrefix := range test.envsPrefixes {
				tmpFolder := t.TempDir()
				err := copy.Copy("testdata", tmpFolder)
				g.Expect(err).To(BeNil())
				t.Setenv(fmt.Sprintf("%s%s", envPrefix, seldontls.EnvCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", tmpFolder))
				t.Setenv(fmt.Sprintf("%s%s", envPrefix, seldontls.EnvKeyLocationSuffix), fmt.Sprintf("%s/tls.key", tmpFolder))
				t.Setenv(fmt.Sprintf("%s%s", envPrefix, seldontls.EnvCaLocationSuffix), fmt.Sprintf("%s/ca.crt", tmpFolder))
			}
			err = c.setupTLS()
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func addVersionedRoute(c *SeldonXDSCache, modelRouteName string, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
	modelVersion := store.NewModelVersion(
		&scheduler.Model{
			Meta:           &scheduler.MetaData{Name: modelName},
			DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
		},
		version,
		"server",
		map[int]store.ReplicaStatus{
			1: {State: store.Loaded},
		},
		false,
		store.ModelAvailable,
	)

	server := &store.ServerSnapshot{
		Name: "server",
		Replicas: map[int]*store.ServerReplica{
			1: store.NewServerReplica("0.0.0.0", 9000, 9001, 1, store.NewServer("server", false), nil, 100, 100, 0, nil, 100),
		},
	}
	c.AddClustersForRoute(modelRouteName, modelName, httpCluster, grpcCluster, modelVersion.GetVersion(), []int{1}, server)
	c.AddRouteClusterTraffic(modelRouteName, modelName, httpCluster, grpcCluster, modelVersion.GetVersion(), traffic, false, false)
}
