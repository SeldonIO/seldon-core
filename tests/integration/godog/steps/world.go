package steps

import (
	"context"
	"fmt"

	"github.com/seldonio/seldon-core/godog/k8sclient"
	"github.com/seldonio/seldon-core/godog/k8sclient/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type World struct {
	namespace            string
	kubeClient           k8sclient.Client
	StartingClusterState string //todo: this will be a combination of starting state awareness of core 2 such as the
	//todo:  server config,seldon config and seldon runtime to be able to reconcile to starting state should we change
	//todo: the state such as reducing replicas to 0 of scheduler to test unavailability
	CurrentModel *Model
	Models       map[string]*Model
	k8sClient    *k8s.SeldonK8sAPI
}

func NewWorld(namespace string) *World {
	k8sClient, err := k8s.NewSeldonK8sAPI()
	if err != nil {
		panic(err)
	}

	return &World{
		namespace: namespace,
		k8sClient: k8sClient,
	}
}

func (w *World) deployModelSpec(spec string) error {
	return w.k8sClient.Create(spec)
}

func (w *World) waitForModelState(state, model, timeout string) error {
	watcher, err := w.k8sClient.Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil, fmt.Errorf("watch channel closed")
			}

			if event.Type == watch.Error {
				return nil, fmt.Errorf("watch error: %v", event.Object)
			}

			if event.Type == watch.Added || event.Type == watch.Modified {
				resource := event.Object.(*v1.MyResource)
				if isReady(resource) {
					return resource, nil
				}
			}

			if event.Type == watch.Deleted {
				return nil, fmt.Errorf("resource was deleted")
			}
		}
	}
}
