package k8sclient

import (
	"context"
	"sync"
)

type ComponentManager interface {
	Name() ComponentName
	Snapshot(ctx context.Context) error
	Restore(ctx context.Context) error
}

type EnvManager struct {
	mu         sync.Mutex
	components map[ComponentName]ComponentManager
}

type ComponentName string

func NewEnvManager(components ...ComponentManager) *EnvManager {
	comps := make(map[ComponentName]ComponentManager, len(components))
	for _, component := range components {
		comps[component.Name()] = component
	}
	return &EnvManager{components: comps}
}

func (e *EnvManager) SnapshotAll(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, c := range e.components {
		if err := c.Snapshot(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (e *EnvManager) RestoreAll(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, c := range e.components {
		if err := c.Restore(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (e *EnvManager) Kafka() *KafkaComponent {
	for _, component := range e.components {
		if k, ok := component.(*KafkaComponent); ok {
			return k
		}
	}
	return nil
}
