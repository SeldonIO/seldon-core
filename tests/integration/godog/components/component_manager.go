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

func (e *EnvManager) Runtime() *SeldonRuntimeComponent {
	for _, comp := range e.components {
		if c, ok := comp.(*SeldonRuntimeComponent); ok {
			return c
		}
	}
	return nil
}
