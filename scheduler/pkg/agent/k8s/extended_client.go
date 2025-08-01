/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package k8s

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

//go:generate go tool mockgen -source=extended_client.go -destination=./mocks/mock_extended_client.go -package=mocks ExtendedClient
//go:generate go tool mockgen -destination=./mocks/mock_discovery_v1.go -package=mocks k8s.io/client-go/kubernetes/typed/discovery/v1 DiscoveryV1Interface,EndpointSliceInterface
//go:generate go tool mockgen -destination=./mocks/mock_k8s_client.go -package=mocks k8s.io/client-go/kubernetes Interface
//go:generate go tool mockgen -destination=./mocks/watch/mock_watch.go -package=mocks k8s.io/apimachinery/pkg/watch Interface
type ExtendedClient interface {
	kubernetes.Interface
	// HasPublishedIP will first perform GET request from endpoints resource to see if there's an IP assigned to
	// id (corresponding to object with metadata.labels.kubernetes.io/service-name == name).
	// If no IP is assigned, it is blocking on a watch channel, waiting for an IP to be assigned. A nil error indicates
	// the named resource has an IP.
	HasPublishedIP(ctx context.Context, name, fieldSelector string) error
}

const defaultLabelSelector = "kubernetes.io/service-name"

type extendedClient struct {
	kubernetes.Interface
	namespace string
	log       *log.Logger
}

func NewExtendedClient(namespace string, client kubernetes.Interface, logger *log.Logger) ExtendedClient {
	return &extendedClient{
		client,
		namespace,
		logger,
	}
}

func (e *extendedClient) HasPublishedIP(ctx context.Context, id, optionalLabelSelector string) error {
	selector, err := e.labelSelector(optionalLabelSelector, id)
	if err != nil {
		return fmt.Errorf("invalid selector: %v", err)
	}

	endpoint, err := e.DiscoveryV1().EndpointSlices(e.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
		Limit:         1,
	})
	if err != nil {
		return fmt.Errorf("failed to list endpointslice for %s: %v", id, err)
	}

	if len(endpoint.Items) > 0 && e.ipIsReady(endpoint.Items[0]) {
		return nil
	}

	watcher, err := e.DiscoveryV1().EndpointSlices(e.namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector:   selector.String(),
		ResourceVersion: endpoint.ResourceVersion,
		Watch:           true,
	})
	if err != nil {
		return fmt.Errorf("failed to watch endpoints for %s: %v", id, err)
	}

	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel unexpectedly closed")
			}

			e.log.Infof("Watch event received from channel: %v", event)

			if event.Type == watch.Added || event.Type == watch.Modified {
				endpoint, ok := event.Object.(*v1.EndpointSlice)
				if !ok {
					return fmt.Errorf("failed to cast object to EndpointSlice: %v", event.Object)
				}

				if !selector.Matches(labels.Set(endpoint.ObjectMeta.Labels)) {
					// watch channel is sending us events for resources we didn't request, safest to quit the channel
					// and attempt re-connect
					return fmt.Errorf("event labels from channel <%s> does not match requested <%s>",
						endpoint.ObjectMeta.Labels, selector.String())
				}

				if e.ipIsReady(*endpoint) {
					return nil
				}
			}
		}
	}
}

func (e *extendedClient) labelSelector(selector, id string) (labels.Selector, error) {
	if selector == "" {
		selector = defaultLabelSelector
	}
	return labels.Parse(fmt.Sprintf("%s=%s", selector, id))
}

func (e *extendedClient) ipIsReady(ep v1.EndpointSlice) bool {
	return len(ep.Endpoints) > 0 &&
		len(ep.Endpoints[0].Addresses) > 0 &&
		(ep.Endpoints[0].Conditions.Ready == nil ||
			*ep.Endpoints[0].Conditions.Ready)
}
