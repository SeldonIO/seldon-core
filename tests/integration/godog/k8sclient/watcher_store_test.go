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
	"testing"
	"time"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	mlopsscheme "github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned/scheme"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// helper to build a minimal WatcherStore for unit tests (no real watchers)
func newTestWatcherStore(t *testing.T) *WatcherStore {
	t.Helper()

	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = mlopsscheme.AddToScheme(s)

	return &WatcherStore{
		namespace: "test-ns",
		logger:    log.New().WithField("test", t.Name()),
		store:     make(map[string]runtime.Object),
		scheme:    s,
		// watchers and doneChan are unused in these tests
	}
}

func newPipeline(name, ns string) *v1alpha1.Pipeline {
	return &v1alpha1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

// 1) WaitForPipelineCondition returns immediately if condition already satisfied
func TestWaitForPipelineConditionImmediate(t *testing.T) {
	ws := newTestWatcherStore(t)
	p := newPipeline("p1", "test-ns")

	// pre-populate store
	ws.put(p)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	called := false
	cond := func(obj runtime.Object) (bool, error) {
		called = true
		if obj == nil {
			return false, nil
		}
		// we only care that it's non-nil for this test
		return true, nil
	}

	if err := ws.WaitForPipelineCondition(ctx, "p1", cond); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatalf("condition func was not called in fast path")
	}
}

// 2) WaitForPipelineCondition blocks until a matching put() arrives
func TestWaitForPipelineConditionOnAddEvent(t *testing.T) {
	ws := newTestWatcherStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		cond := func(obj runtime.Object) (bool, error) {
			return obj != nil, nil
		}
		done <- ws.WaitForPipelineCondition(ctx, "p2", cond)
	}()

	// give the waiter a moment to register
	time.Sleep(100 * time.Millisecond)

	// now simulate the watch ADD event
	p := newPipeline("p2", "test-ns")
	ws.put(p)

	err := <-done
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// 3) WaitForPipelineCondition with a "deleted" condition unblocks on delete
//
// This assumes you have a PipelineDeleted condition that treats obj == nil as deleted.
func TestWaitForPipelineDeletedOnDeleteEvent(t *testing.T) {
	ws := newTestWatcherStore(t)
	p := newPipeline("p3", "test-ns")

	// object exists initially
	ws.put(p)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		// Using a deleted-style condition: obj == nil => done
		cond := func(obj runtime.Object) (bool, error) {
			// pipeline considered deleted when obj is nil
			return obj == nil, nil
		}
		done <- ws.WaitForPipelineCondition(ctx, "p3", cond)
	}()

	time.Sleep(100 * time.Millisecond)

	// simulate watch DELETED event
	ws.delete(p)

	err := <-done
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// 4) WaitForPipelineCondition returns context error if condition not met in time
func TestWaitForPipelineConditionTimeout(t *testing.T) {
	ws := newTestWatcherStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cond := func(obj runtime.Object) (bool, error) {
		// condition that can never be satisfied in this test
		return false, nil
	}

	err := ws.WaitForPipelineCondition(ctx, "non-existent", cond)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if ctx.Err() == nil {
		t.Fatalf("expected context to be cancelled")
	}
}

// 5) notifyWaiters removes waiter once condition satisfied
func TestNotifyWaitersRemovesSatisfiedWaiter(t *testing.T) {
	ws := newTestWatcherStore(t)

	// We'll wait on a specific key
	key := "test-ns/Pipeline/p4"

	resultCh := make(chan error, 1)
	w := &waiter{
		key: key,
		cond: func(obj runtime.Object) (bool, error) {
			return true, nil // immediately satisfied
		},
		result: resultCh,
	}

	ws.mu.Lock()
	ws.waiters = append(ws.waiters, w)
	ws.mu.Unlock()

	// notify with some obj (could be nil or not, condition returns true anyway)
	ws.notifyWaiters(key, nil)

	// waiter should have been signalled and removed
	select {
	case err := <-resultCh:
		if err != nil {
			t.Fatalf("expected nil error from waiter, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected waiter to be signalled")
	}

	ws.mu.RLock()
	defer ws.mu.RUnlock()
	if len(ws.waiters) != 0 {
		t.Fatalf("expected no remaining waiters, got %d", len(ws.waiters))
	}
}
