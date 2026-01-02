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
	"fmt"
	"testing"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
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
	t *testing.T,
	logger log.FieldLogger,
	modelStore Storage[*db.Model],
	serverStore Storage[*db.Server],
	eventHub *coordinator.EventHub) *TestMemoryStore {
	if t == nil {
		panic("testing.T is required, must only be run via tests")
	}
	m := NewMemoryStore(logger, modelStore, serverStore, eventHub)
	return &TestMemoryStore{m}
}

func (t *TestMemoryStore) DirectlyUpdateModelStatus(model ModelID, state *db.ModelStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	found, err := t.store.models.Get(model.Name)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	version := found.GetVersion(model.Version)
	if version == nil {
		return errors.New("version not found")
	}

	version.State = state
	return t.store.models.Update(found)
}
