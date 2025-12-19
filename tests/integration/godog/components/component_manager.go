package components

import (
	"context"
	"fmt"
	"sync"
)

type Component interface {
	Name() ComponentName
	Snapshot(ctx context.Context) error
	Restore(ctx context.Context) error
	MakeUnavailable(ctx context.Context) error
	MakeAvailable(ctx context.Context) error
	Scale(ctx context.Context, replicas int32) error // optional capability
}

type ComponentName string

type EnvManager struct {
	mu         sync.Mutex
	components map[ComponentName]Component
	order      []ComponentName // deterministic snapshot/restore ordering
}

func NewEnvManager(cs ...Component) (*EnvManager, error) {
	m := make(map[ComponentName]Component, len(cs))
	order := make([]ComponentName, 0, len(cs))

	for _, c := range cs {
		if c == nil {
			return nil, fmt.Errorf("nil component provided")
		}
		name := c.Name()
		if name == "" {
			return nil, fmt.Errorf("component has empty name: %T", c)
		}
		if _, exists := m[name]; exists {
			return nil, fmt.Errorf("duplicate component name: %q", name)
		}
		m[name] = c
		order = append(order, name)
	}

	return &EnvManager{
		components: m,
		order:      order,
	}, nil
}

func (e *EnvManager) Component(name ComponentName) (Component, error) {
	c, ok := e.components[name]
	if !ok {
		return nil, fmt.Errorf("unknown component: %q", name)
	}
	return c, nil
}

//// Capability getters (no assertions in steps)
//func (e *EnvManager) Scalable(name ComponentName) (Scalable, error) {
//	c, err := e.Component(name)
//	if err != nil {
//		return nil, err
//	}
//	s, ok := c.(Scalable)
//	if !ok {
//		return nil, fmt.Errorf("component %q does not support scaling", name)
//	}
//	return s, nil
//}
//
//func (e *EnvManager) Restartable(name ComponentName) (Restartable, error) {
//	c, err := e.Component(name)
//	if err != nil {
//		return nil, err
//	}
//	r, ok := c.(Restartable)
//	if !ok {
//		return nil, fmt.Errorf("component %q does not support restart", name)
//	}
//	return r, nil
//}

func (e *EnvManager) SnapshotAll(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, name := range e.order {
		if err := e.components[name].Snapshot(ctx); err != nil {
			return fmt.Errorf("snapshot %s: %w", name, err)
		}
	}
	return nil
}

func (e *EnvManager) RestoreAll(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, name := range e.order {
		if err := e.components[name].Restore(ctx); err != nil {
			return fmt.Errorf("restore %s: %w", name, err)
		}
	}
	return nil
}

// Typed retrieval without scattering type assertions in steps:
func (e *EnvManager) Kafka() (*KafkaComponent, error) {
	c, err := e.Component(Kafka)
	if err != nil {
		return nil, err
	}
	k, ok := c.(*KafkaComponent)
	if !ok {
		return nil, fmt.Errorf("component %q is not *KafkaComponent", Kafka)
	}
	return k, nil
}
