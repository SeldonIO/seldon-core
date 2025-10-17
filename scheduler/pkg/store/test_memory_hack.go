/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

type TestMemoryStore struct {
	*ModelServerService
}

type ModelID struct {
	Name    string
	Version uint32
}

// NewTestMemory DO NOT USE for non-test code. This is purely meant for using in tests where an integration test is
// wanted where the real memory cache is needed, but the test needs the ability to directly manipulate the model
// statuses, which can't be achieved with ModelServerService. TestMemoryStore embeds ModelServerService and adds DirectlyUpdateModelStatus
// to modify the statuses.
func NewTestMemory(
	t *testing.T,
	logger log.FieldLogger,
	store *LocalSchedulerStore,
	eventHub *coordinator.EventHub) *TestMemoryStore {
	if t == nil {
		panic("testing.T is required, must only be run via tests")
	}
	m := NewModelServerService(logger, store, eventHub)
	return &TestMemoryStore{m}
}

func (t *TestMemoryStore) DirectlyUpdateModelStatus(model ModelID, state ModelStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	found, ok := t.cache.models[model.Name]
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
