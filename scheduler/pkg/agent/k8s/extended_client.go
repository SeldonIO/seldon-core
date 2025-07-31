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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

//go:generate go tool mockgen -source=extended_client.go -destination=./mocks/mock_extended_client.go -package=mocks ExtendedClient
//go:generate go tool mockgen -destination=./mocks/mock_core_v1.go -package=mocks k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,EndpointsInterface
//go:generate go tool mockgen -destination=./mocks/mock_k8s_client.go -package=mocks k8s.io/client-go/kubernetes Interface
//go:generate go tool mockgen -destination=./mocks/watch/mock_watch.go -package=mocks k8s.io/apimachinery/pkg/watch Interface
type ExtendedClient interface {
	kubernetes.Interface
	// HasPublishedIP will first perform GET request from endpoints resource to see if there's an IP assigned to
	// name (corresponding to object with metadata.name == name). If no IP is assigned, it is blocking on a watch
	// channel, waiting for an IP to be assigned. A nil error indicates the named resource has an IP.
	HasPublishedIP(ctx context.Context, name string) error
}

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

func (e *extendedClient) HasPublishedIP(ctx context.Context, name string) error {
	endpoint, err := e.CoreV1().Endpoints(e.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get endpoint for %s: %v", name, err)
	}

	if len(endpoint.Subsets) > 0 && len(endpoint.Subsets[0].Addresses) > 0 {
		return nil
	}

	watcher, err := e.CoreV1().Endpoints(e.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", name),
		ResourceVersion: endpoint.ResourceVersion,
		Watch:           true,
	})
	if err != nil {
		return fmt.Errorf("failed to watch endpoints for %s: %v", name, err)
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
				endpoint, ok := event.Object.(*corev1.Endpoints)
				if !ok {
					return fmt.Errorf("failed to cast object to Endpoints: %v", event.Object)
				}

				if endpoint.ObjectMeta.Name != name {
					// watch channel is sending us events we didn't request, safest to quit the channel
					// and attempt re-connect
					return fmt.Errorf("event from channel <%s> does not match requested <%s>", endpoint.ObjectMeta.Name, name)
				}

				if len(endpoint.Subsets) > 0 && len(endpoint.Subsets[0].Addresses) > 0 {
					return nil
				}
			}
		}
	}
}
