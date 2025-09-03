package store

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

type TestMemoryStore struct {
	*MemoryStore
}

type ModelID struct {
	Name    string
	Version uint32
}

// NewTestMemory DO NOT USE for non-test code. This is purely meant for using in tests where an integration test is
// wanted where the real memory store is needed, but the test needs the ability to directly manipulate the model
// statuses, which can't be achieved with MemoryStore. TestMemoryStore embeds MemoryStore and adds DirectlyUpdateModelStatus
// to modify the statuses.
func NewTestMemory(
	logger log.FieldLogger,
	store *LocalSchedulerStore,
	eventHub *coordinator.EventHub) *TestMemoryStore {
	m := NewMemoryStore(logger, store, eventHub)
	return &TestMemoryStore{m}
}

func (t *TestMemoryStore) DirectlyUpdateModelStatus(model ModelID, state ModelStatus) error {
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
