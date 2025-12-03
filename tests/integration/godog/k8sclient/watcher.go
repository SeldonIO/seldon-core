package k8sclient

import (
	"fmt"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WatcherStorage interface {
	Put(runtime.Object)
	Get(runtime.Object) (runtime.Object, bool)
	Clear()
	Start()
}

type WatcherStore struct {
	namespace      string
	label          map[string]string
	watchInterface watch.Interface
	store          string //todo finish underlying store
	doneChan       chan struct{}
}

// NewWatcherStore receives events that match on a particular object list and creates a database store to query crd state
func NewWatcherStore(namespace string, label map[string]string, watch client.WithWatch) (*WatcherStore, error) {
	// have watchers on particular object lists

	watchInter, err := watch.Watch(nil, nil, nil)

	// process events from the resulting channel of the watcher and build a key value database for asserting crd states

	return &WatcherStore{}
}

func (s *WatcherStore) Start() {
	for {
		select {
		case eve := <-s.watchInterface.ResultChan():
		// todo: process event and put in key value store

		case <-s.doneChan:
			return
		}

	}
}
