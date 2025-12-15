/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package k8sclient

import (
	"context"
	"fmt"
	"sync"

	mlopsscheme "github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned/scheme"
	"github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned/typed/mlops/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type WatcherStorage interface {
	WaitForObjectCondition(ctx context.Context, obj runtime.Object, cond ConditionFunc) error
	WaitForModelCondition(ctx context.Context, modelName string, cond ConditionFunc) error
	WaitForPipelineCondition(ctx context.Context, modelName string, cond ConditionFunc) error
	Clear()
	Start()
	Stop()
}

type objectKind string

const (
	model    objectKind = "Model"
	pipeline objectKind = "Pipeline"
)

type WatcherStore struct {
	namespace       string
	label           string
	mlopsClient     v1alpha1.MlopsV1alpha1Interface
	modelWatcher    watch.Interface
	pipelineWatcher watch.Interface
	logger          log.FieldLogger
	scheme          *runtime.Scheme

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
func NewWatcherStore(namespace string, label string, mlopsClient v1alpha1.MlopsV1alpha1Interface, logger *log.Logger) (*WatcherStore, error) {
	if logger == nil {
		logger = log.New()
	}

	modelWatcher, err := mlopsClient.Models(namespace).Watch(context.Background(), v1.ListOptions{LabelSelector: DefaultCRDTestSuiteLabel})
	if err != nil {
		return nil, fmt.Errorf("failed to create model watcher: %w", err)
	}

	pipelineWatcher, err := mlopsClient.Pipelines(namespace).Watch(context.Background(), v1.ListOptions{LabelSelector: DefaultCRDTestSuiteLabel})
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline watcher: %w", err)
	}

	// Base scheme + register your CRDs
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)      // core k8s types (optional but fine)
	_ = mlopsscheme.AddToScheme(s) // <-- this is the key line for your CRDs

	return &WatcherStore{
		namespace:       namespace,
		label:           label,
		mlopsClient:     mlopsClient,
		modelWatcher:    modelWatcher,
		pipelineWatcher: pipelineWatcher,
		logger:          logger.WithField("client", "watcher_store"),
		store:           make(map[string]runtime.Object),
		doneChan:        make(chan struct{}),
		scheme:          s,
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

				accessor, err := meta.Accessor(event.Object)
				if err != nil {
					s.logger.WithError(err).Error("failed to access model watcher")
				} else {
					s.logger.WithField("event", event).Tracef("new model watch event with name: %s on namespace: %s", accessor.GetName(), accessor.GetNamespace())
				}

				if event.Object == nil {
					continue
				}

				switch event.Type {
				case watch.Added, watch.Modified:
					s.put(event.Object)
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
	go func() {
		for {
			select {
			case event, ok := <-s.pipelineWatcher.ResultChan():
				if !ok {
					// channel closed: watcher terminated
					return
				}

				accessor, err := meta.Accessor(event.Object)
				if err != nil {
					s.logger.WithError(err).Error("failed to access pipeline watcher")
				} else {
					s.logger.WithField("event", event).Tracef("new pipeline watch event with name: %s on namespace: %s", accessor.GetName(), accessor.GetNamespace())
				}

				if event.Object == nil {
					continue
				}

				switch event.Type {
				case watch.Added, watch.Modified:
					s.put(event.Object)
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
		ns = s.namespace // or "_cluster" if you prefer
	}

	// Prefer scheme-based GVK for typed objects
	gvks, _, err := s.scheme.ObjectKinds(obj)
	if err != nil || len(gvks) == 0 {
		// fallback: TypeMeta if present
		if ta, taErr := meta.TypeAccessor(obj); taErr == nil && ta.GetKind() != "" {
			return fmt.Sprintf("%s/%s/%s", ns, ta.GetKind(), accessor.GetName()), nil
		}
		return "", fmt.Errorf("failed to determine kind for %T: %w", obj, err)
	}

	kind := gvks[0].Kind
	return fmt.Sprintf("%s/%s/%s", ns, kind, accessor.GetName()), nil
}

func (s *WatcherStore) put(obj runtime.Object) {
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

func (s *WatcherStore) get(obj runtime.Object) (runtime.Object, bool) {
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

func (s *WatcherStore) WaitForObjectCondition(ctx context.Context, obj runtime.Object, cond ConditionFunc) error {
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
func (s *WatcherStore) WaitForModelCondition(ctx context.Context, modelName string, cond ConditionFunc) error {
	key := fmt.Sprintf("%s/%s/%s", s.namespace, model, modelName)
	return s.waitForKey(ctx, key, cond)
}

func (s *WatcherStore) WaitForPipelineCondition(ctx context.Context, pipelineName string, cond ConditionFunc) error {
	key := fmt.Sprintf("%s/%s/%s", s.namespace, pipeline, pipelineName)
	return s.waitForKey(ctx, key, cond)
}

func (s *WatcherStore) waitForKey(ctx context.Context, key string, cond ConditionFunc) error {

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
