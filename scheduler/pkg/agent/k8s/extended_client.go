package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch2 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type ExtendedClient interface {
	kubernetes.Interface
	// HasPublishedIP TODO
	HasPublishedIP(ctx context.Context, name string) error
}

type extendedClient struct {
	kubernetes.Interface
	namespace string
}

func NewExtendedClient(namespace string, client kubernetes.Interface) ExtendedClient {
	return &extendedClient{
		client,
		namespace,
	}
}

func (e *extendedClient) HasPublishedIP(ctx context.Context, name string) error {
	endpoint, err := e.CoreV1().Endpoints(e.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to list endpoints for %s: %v", name, err)
	}

	if len(endpoint.Subsets) > 0 && len(endpoint.Subsets[0].Addresses) > 0 {
		return nil
	}

	watch, err := e.CoreV1().Endpoints(e.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", name),
		ResourceVersion: endpoint.ResourceVersion,
		Watch:           true,
	})
	if err != nil {
		return fmt.Errorf("failed to watch endpoints for %s: %v", name, err)
	}

	defer watch.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, closed := <-watch.ResultChan():
			if closed {
				return fmt.Errorf("watch channel unexpectedly closed")
			}
			if event.Type == watch2.Added || event.Type == watch2.Modified {
				endpoint, ok := event.Object.(*v1.EndpointSlice)
				if !ok {
					return fmt.Errorf("failed to cast object to EndpointSlice: %v", event.Object)
				}
				if len(endpoint.Endpoints) > 0 && len(endpoint.Endpoints[0].Addresses) > 0 {
					return nil
				}
			}
		}
	}
}
