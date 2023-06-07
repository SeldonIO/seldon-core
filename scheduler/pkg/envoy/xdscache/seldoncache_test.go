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
	"testing"

	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
)

// Checks a cluster remains until all routes are removed
func TestAddRemoveHttpAndGrpcRoute(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	c := NewSeldonXDSCache(logger, &PipelineGatewayDetails{})

	addVersionedRoute := func(c *SeldonXDSCache, routeName string, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, routeName, modelName, version, false)
		c.AddCluster(grpcCluster, routeName, modelName, version, true)
		c.AddRouteClusterTraffic(routeName, modelName, version, traffic, httpCluster, grpcCluster, true, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	httpCluster := "http1"
	grpcCluster := "grpc1"
	model1 := "m1"
	model2 := "m2"
	addVersionedRoute(c, model1, model1, httpCluster, grpcCluster, 100, 1)
	addVersionedRoute(c, model2, model2, httpCluster, grpcCluster, 100, 1)

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

	addVersionedRoute := func(c *SeldonXDSCache, routeName string, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, routeName, modelName, version, false)
		c.AddCluster(grpcCluster, routeName, modelName, version, true)
		c.AddRouteClusterTraffic(routeName, modelName, version, traffic, httpCluster, grpcCluster, true, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger, &PipelineGatewayDetails{})

	httpCluster := "http1"
	grpcCluster := "grpc1"
	model1 := "m1"
	model2 := "m2"
	addVersionedRoute(c, model1, model1, httpCluster, grpcCluster, 40, 1)
	addVersionedRoute(c, model1, model1, httpCluster, grpcCluster, 60, 2)

	// check what we have added
	g.Expect(len(c.Routes[model1].Clusters)).To(Equal(2))
	clusters := c.Routes[model1].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model1, Version: 2}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model1, Version: 2}]).To(BeTrue())

	addVersionedRoute(c, model2, model2, httpCluster, grpcCluster, 100, 1)

	// check what we added
	g.Expect(len(c.Routes[model2].Clusters)).To(Equal(1))
	clusters = c.Routes[model2].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(100)))
	g.Expect(c.Clusters[httpCluster].Routes[resources.RouteVersionKey{RouteName: model2, ModelName: model2, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.RouteVersionKey{RouteName: model2, ModelName: model2, Version: 1}]).To(BeTrue())
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

	addVersionedRoute := func(c *SeldonXDSCache, routeName string, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, routeName, modelName, version, false)
		c.AddCluster(grpcCluster, routeName, modelName, version, true)
		c.AddRouteClusterTraffic(routeName, modelName, version, traffic, httpCluster, grpcCluster, true, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger, &PipelineGatewayDetails{})

	routeName := "r1"
	httpCluster := "http1"
	grpcCluster := "grpc1"
	model1 := "m1"
	addVersionedRoute(c, routeName, model1, httpCluster, grpcCluster, 40, 1)
	addVersionedRoute(c, routeName, model1, httpCluster, grpcCluster, 60, 2)

	// check what we have added
	g.Expect(len(c.Routes[routeName].Clusters)).To(Equal(2))
	clusters := c.Routes[routeName].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(40)))
	g.Expect(clusters[1].TrafficWeight).To(Equal(uint32(60)))
	g.Expect(len(c.Clusters[httpCluster].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcCluster].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpCluster].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcCluster].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpCluster].Routes[resources.RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 2}]).To(BeTrue())
	g.Expect(c.Clusters[grpcCluster].Routes[resources.RouteVersionKey{RouteName: routeName, ModelName: model1, Version: 2}]).To(BeTrue())

	err := c.RemoveRoute(routeName)
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
		c.AddCluster(httpCluster, modelRouteName, modelName, version, false)
		c.AddCluster(grpcCluster, modelRouteName, modelName, version, true)
		c.AddRouteClusterTraffic(modelRouteName, modelName, version, traffic, httpCluster, grpcCluster, true, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger, &PipelineGatewayDetails{})

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
	g.Expect(c.Clusters[httpClusterModel1].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel2].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model2, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel2].Routes[resources.RouteVersionKey{RouteName: model1, ModelName: model2, Version: 1}]).To(BeTrue())

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

func TestAddRemoveHttpAndGrpcRouteVersionsForDifferentRoutesSameModel(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()

	addVersionedRoute := func(c *SeldonXDSCache, modelRouteName string, modelName string, httpCluster string, grpcCluster string, traffic uint32, version uint32) {
		c.AddCluster(httpCluster, modelRouteName, modelName, version, false)
		c.AddCluster(grpcCluster, modelRouteName, modelName, version, true)
		c.AddRouteClusterTraffic(modelRouteName, modelName, version, traffic, httpCluster, grpcCluster, true, false)
		c.AddEndpoint(httpCluster, "0.0.0.0", 9000)
		c.AddEndpoint(grpcCluster, "0.0.0.0", 9001)
	}

	c := NewSeldonXDSCache(logger, &PipelineGatewayDetails{})

	route1 := "r1"
	route2 := "r2"
	httpClusterModel1 := "model1_http1"
	grpcClusterModel1 := "model1_grpc1"
	model1 := "m1"
	addVersionedRoute(c, route1, model1, httpClusterModel1, grpcClusterModel1, 100, 1)
	addVersionedRoute(c, route2, model1, httpClusterModel1, grpcClusterModel1, 100, 1)

	// check what we have added
	g.Expect(len(c.Routes[route1].Clusters)).To(Equal(1))
	g.Expect(len(c.Routes[route2].Clusters)).To(Equal(1))
	clusters := c.Routes[route1].Clusters
	g.Expect(clusters[0].TrafficWeight).To(Equal(uint32(100)))
	g.Expect(len(c.Clusters[httpClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(len(c.Clusters[grpcClusterModel1].Endpoints)).To(Equal(1))
	g.Expect(c.Clusters[httpClusterModel1].Grpc).To(BeFalse())
	g.Expect(c.Clusters[grpcClusterModel1].Grpc).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel1].Routes[resources.RouteVersionKey{RouteName: route1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[resources.RouteVersionKey{RouteName: route1, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[httpClusterModel1].Routes[resources.RouteVersionKey{RouteName: route2, ModelName: model1, Version: 1}]).To(BeTrue())
	g.Expect(c.Clusters[grpcClusterModel1].Routes[resources.RouteVersionKey{RouteName: route2, ModelName: model1, Version: 1}]).To(BeTrue())

	err := c.RemoveRoute(route1)
	g.Expect(err).To(BeNil())
	_, ok := c.Clusters[httpClusterModel1]
	g.Expect(ok).To(BeTrue()) // http Cluster not removed
	_, ok = c.Clusters[grpcClusterModel1]
	g.Expect(ok).To(BeTrue()) // grpc Cluster not removed
	err = c.RemoveRoute(route2)
	g.Expect(err).To(BeNil())
	_, ok = c.Clusters[httpClusterModel1]
	g.Expect(ok).To(BeFalse()) // http Cluster removed
	_, ok = c.Clusters[grpcClusterModel1]
	g.Expect(ok).To(BeFalse()) // grpc Cluster removed
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
			c := NewSeldonXDSCache(logger, &PipelineGatewayDetails{})
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
			err := c.SetupTLS()
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
