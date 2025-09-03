package store

import (
	"errors"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	log "github.com/sirupsen/logrus"
)

type TestMemoryStore struct {
	*MemoryStore
}

type ModelID struct {
	Name    string
	Version uint32
}

// NewTestMemory DO NOT USE for non-test code
func NewTestMemory(
	logger log.FieldLogger,
	store *LocalSchedulerStore,
	eventHub *coordinator.EventHub) *TestMemoryStore {
	m := NewMemoryStore(logger, store, eventHub)
	return &TestMemoryStore{m}
}

func (t *TestMemoryStore) HACKDirectlyUpdateModelStatus(model ModelID, state ModelStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	found, ok := t.store.models[model.Name]
	if !ok {
		return errors.New("model not found")
	}

	version := found.GetVersion(model.Version)
	if version == nil {
		return errors.New("version not found")
	}

	version.state = state
	return nil
}
