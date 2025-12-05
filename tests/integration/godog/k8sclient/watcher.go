package k8sclient

import (
	"context"
	"fmt"
	"sync"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WatcherStorage interface {
	Put(runtime.Object)
	Get(runtime.Object) (runtime.Object, bool)
	WaitFor(ctx context.Context, obj runtime.Object, cond ConditionFunc) error
	Clear()
	Start()
	Stop()
}

type WatcherStore struct {
	namespace string
	label     map[string]string

	modelWatcher watch.Interface

	mu      sync.RWMutex
	store   map[string]runtime.Object // key: "namespace/name"
	waiters []*waiter

	doneChan chan struct{}
}

type waiter struct {
	key    string
	cond   ConditionFunc
	result chan error
}

type ConditionFunc func(obj runtime.Object) (done bool, err error)

// NewWatcherStore receives events that match on a particular object list and creates a database store to query crd state
func NewWatcherStore(namespace string, label map[string]string, w client.WithWatch) (*WatcherStore, error) {
	modelWatcher, err := w.Watch(
		context.Background(),
		&mlopsv1alpha1.ModelList{},
		client.InNamespace(namespace),
		client.MatchingLabels(label),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create model watcher: %w", err)
	}

	return &WatcherStore{
		namespace:    namespace,
		label:        label,
		modelWatcher: modelWatcher,
		store:        make(map[string]runtime.Object),
		doneChan:     make(chan struct{}),
	}, nil
}

// Start watches on events for models
func (s *WatcherStore) Start() {
	go func() {
		for {
			select {
			case event, ok := <-s.modelWatcher.ResultChan():
				if !ok {
					// channel closed: watcher terminated
					return
				}

				fmt.Printf("model watch event: %v\n", event)

				if event.Object == nil {
					continue
				}

				switch event.Type {
				case watch.Added, watch.Modified:
					s.Put(event.Object)
				case watch.Deleted:
					s.delete(event.Object)
				case watch.Error:
					fmt.Printf("model watch error: %v\n", event.Object)
				}

			case <-s.doneChan:
				// Stop underlying watcher and exit
				s.modelWatcher.Stop()
				return
			}
		}
	}()
}

// Stop terminates the watcher loop.
func (s *WatcherStore) Stop() {
	select {
	case <-s.doneChan:
		// already closed
	default:
		close(s.doneChan)
	}
}

func (s *WatcherStore) keyFor(obj runtime.Object) (string, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata accessor: %w", err)
	}

	ns := accessor.GetNamespace()
	if ns == "" {
		// fall back to store namespace if the object is cluster-scoped or unset
		ns = s.namespace
	}

	return fmt.Sprintf("%s/%s", ns, accessor.GetName()), nil
}

func (s *WatcherStore) Put(obj runtime.Object) {
	if obj == nil {
		return
	}

	key, err := s.keyFor(obj)
	if err != nil {
		// log if you have a logger
		return
	}

	s.mu.Lock()
	s.store[key] = obj
	s.mu.Unlock()
	s.notifyWaiters(key, obj)
}

func (s *WatcherStore) Get(obj runtime.Object) (runtime.Object, bool) {
	if obj == nil {
		return nil, false
	}

	key, err := s.keyFor(obj)
	if err != nil {
		return nil, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.store[key]
	return v, ok
}

func (s *WatcherStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = make(map[string]runtime.Object)
}

// internal helper for delete events
func (s *WatcherStore) delete(obj runtime.Object) {
	if obj == nil {
		return
	}

	key, err := s.keyFor(obj)
	if err != nil {
		return
	}

	s.mu.Lock()
	delete(s.store, key)
	s.mu.Unlock()

	// Optional: you may want to notify waiters that the object is gone.
	s.notifyWaiters(key, nil)
}

func (s *WatcherStore) WaitFor(ctx context.Context, obj runtime.Object, cond ConditionFunc) error {
	key, err := s.keyFor(obj)
	if err != nil {
		return err
	}

	// Fast path: check current state
	s.mu.RLock()
	existing, ok := s.store[key]
	s.mu.RUnlock()

	if ok {
		done, err := cond(existing)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	// Slow path: register a waiter
	w := &waiter{
		key:    key,
		cond:   cond,
		result: make(chan error, 1), // buffered so we don't block notifier
	}

	s.mu.Lock()
	s.waiters = append(s.waiters, w)
	s.mu.Unlock()

	// Wait for either condition satisfied or context cancelled
	select {
	case <-ctx.Done():
		s.removeWaiter(w)
		return ctx.Err()
	case err := <-w.result:
		return err
	}
}

func (s *WatcherStore) removeWaiter(target *waiter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, w := range s.waiters {
		if w == target {
			s.waiters = append(s.waiters[:i], s.waiters[i+1:]...)
			return
		}
	}
}

func (s *WatcherStore) notifyWaiters(key string, obj runtime.Object) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// We’ll rebuild the slice with only remaining waiters
	remaining := s.waiters[:0]

	for _, w := range s.waiters {
		if w.key != key {
			remaining = append(remaining, w)
			continue
		}

		done, err := w.cond(obj)
		if !done && err == nil {
			// keep waiting
			remaining = append(remaining, w)
			continue
		}

		// Condition satisfied or error: signal and drop waiter
		w.result <- err
		close(w.result)
	}

	s.waiters = remaining
}
