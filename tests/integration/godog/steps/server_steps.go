/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/cucumber/godog"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/yaml"
)

type server struct {
	label           map[string]string
	namespace       string
	seldonK8sClient versioned.Interface
	k8sClient       *k8sclient.K8sClient
	currentServer   *mlopsv1alpha1.Server
	log             logrus.FieldLogger
}

func newServer(label map[string]string, namespace string, seldonK8sClient versioned.Interface, log logrus.FieldLogger, k8sClient *k8sclient.K8sClient) *server {
	return &server{
		label:           label,
		namespace:       namespace,
		seldonK8sClient: seldonK8sClient,
		k8sClient:       k8sClient,
		log:             log,
	}
}

func LoadServerSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I deploy server spec with timeout "([^"]+)":$`, func(timeout string, spec *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.server.deployServerSpec(ctx, spec)
		})
	})
	scenario.Step(`^the server should eventually become Ready with timeout "([^"]+)"$`, func(timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.server.requiresCurrentServer(func() error {
				return w.server.waitForServerReady(ctx)
			})
		})
	})
	scenario.Step(`^ensure only "([^"]+)" pod\(s\) are deployed for server name "([^"]+)" and they are Ready$`, func(replicaCount int32, serverName string) error {
		return withTimeoutCtx("10s", func(ctx context.Context) error {
			return w.server.checkPodsAreReadyForServer(ctx, replicaCount, serverName)
		})
	})
	scenario.Step(`^ensure only "([^"]+)" pod\(s\) are deployed for server and they are Ready$`, func(replicaCount int32) error {
		return withTimeoutCtx("10s", func(ctx context.Context) error {
			return w.server.requiresCurrentServer(func() error {
				return w.server.checkPodsAreReady(ctx, replicaCount)
			})
		})
	})
	scenario.Step(`^remove any other server deployments$`, func() error {
		return withTimeoutCtx("10s", func(ctx context.Context) error {
			return w.server.requiresCurrentServer(func() error {
				return w.server.removeOtherServers(ctx)
			})
		})
	})
	scenario.Step(`^I remove any other server deployments which are not "([^"]+)"$`, func(servers string) error {
		return withTimeoutCtx("10s", func(ctx context.Context) error {
			return w.server.removeServers(ctx, servers)
		})
	})
	scenario.Step(`^I delete server "([^"]+)" with timeout "([^"]+)"$`, func(server, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.server.deleteServer(ctx, server)
		})
	})
}

func (s *server) requiresCurrentServer(callback func() error) error {
	if s.currentServer == nil {
		return errors.New("current server not set")
	}
	return callback()
}

func (s *server) deployServerSpec(ctx context.Context, spec *godog.DocString) error {
	serverSpec := &mlopsv1alpha1.Server{}
	if err := yaml.Unmarshal([]byte(spec.Content), &serverSpec); err != nil {
		return fmt.Errorf("failed unmarshalling server spec: %w", err)
	}
	serverSpec.Namespace = s.namespace
	s.currentServer = serverSpec
	s.applyScenarioLabel()

	s.log.Debugf("Attempts to create server %s", serverSpec.Name)

	if _, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).Create(ctx, s.currentServer, metav1.CreateOptions{}); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			s.log.Debugf("server %s already exists, checking if equal", s.currentServer.Name)
			deployedServer, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).Get(ctx, s.currentServer.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed getting server: %w", err)
			}
			if equality.Semantic.DeepEqual(serverSpec.Spec, deployedServer.Spec) {
				s.log.Debugf("server %s deployed spec equals desired spec", s.currentServer.Name)
				return nil
			}
			s.log.Debugf("server %s deployed spec needs updating to desired spec", s.currentServer.Name)
			deployedServer.Spec = s.currentServer.Spec
			if _, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).Update(ctx, deployedServer, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("failed updating server: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed creating server: %w", err)
	}
	return nil
}

func (s *server) applyScenarioLabel() {
	if s.currentServer.Labels == nil {
		s.currentServer.Labels = s.label
	} else {
		maps.Copy(s.currentServer.Labels, s.label)
	}

	// todo: change this approach
	for k, v := range k8sclient.DefaultCRDTestSuiteLabelMap {
		s.currentServer.Labels[k] = v
	}
}

func (s *server) removeServers(ctx context.Context, keepServers string) error {
	gotServers, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed listing servers: %w", err)
	}

	for _, server := range gotServers.Items {
		if strings.Contains(keepServers, server.Name) {
			continue
		}
		s.log.Debugf("Removing server %s not in keep list %s", server.Name, keepServers)
		if err := s.deleteServer(ctx, server.Name); err != nil {
			return fmt.Errorf("failed deleting server: %w", err)
		}
		s.log.Infof("removed server %q", server)
	}

	return nil
}

func (s *server) removeOtherServers(ctx context.Context) error {
	servers, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed listing servers: %w", err)
	}
	for _, server := range servers.Items {
		if server.Name == s.currentServer.Name {
			continue
		}
		if err := s.deleteServer(ctx, server.Name); err != nil {
			return fmt.Errorf("failed deleting server: %w", err)
		}
		s.log.Infof("removed server %q", server.Name)
	}

	return nil
}

func (s *server) deleteServer(ctx context.Context, server string) error {
	return s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).Delete(ctx, server, metav1.DeleteOptions{})
}

func (s *server) checkPodsAreReady(ctx context.Context, replicaCount int32) error {
	return s.checkPodsAreReadyForServer(ctx, replicaCount, s.currentServer.Name)
}

func (s *server) checkPodsAreReadyForServer(ctx context.Context, replicaCount int32, serverName string) error {
	statefulSet := &v1.StatefulSet{}
	if err := s.k8sClient.KubeClient.Get(ctx, types.NamespacedName{
		Namespace: s.namespace,
		Name:      serverName,
	}, statefulSet); err != nil {
		return fmt.Errorf("failed getting statefulSet: %w", err)
	}

	if *statefulSet.Spec.Replicas != replicaCount {
		return fmt.Errorf("expected %d replicas but got %d on statefulset spec", replicaCount, *statefulSet.Spec.Replicas)
	}

	if statefulSet.Status.ReadyReplicas == replicaCount {
		return nil
	}

	return fmt.Errorf("ready replicas %d does not match %d", statefulSet.Status.ReadyReplicas, replicaCount)
}

func (s *server) waitForServerReady(ctx context.Context) error {
	foundServer, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).Get(ctx, s.currentServer.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed getting server: %w", err)
	}

	if foundServer.Status.IsReady() {
		return nil
	}

	watcher, err := s.seldonK8sClient.MlopsV1alpha1().Servers(s.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", s.currentServer.Name),
		ResourceVersion: foundServer.ResourceVersion,
		Watch:           true,
	})
	if err != nil {
		return fmt.Errorf("failed subscribed to watch server: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}

			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			if event.Type == watch.Added || event.Type == watch.Modified {
				srv := event.Object.(*mlopsv1alpha1.Server)
				if srv.Status.IsReady() {
					return nil
				}
				s.log.Debugf("got watch event: server %s is not ready, still waiting", s.currentServer.Name)
				continue
			}

			if event.Type == watch.Deleted {
				return fmt.Errorf("resource was deleted")
			}
		}
	}
}
