/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"context"
	"sync"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

func TestStorageInMemory_Get(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		setup       func() *StorageInMemory[*db.Model]
		id          string
		expectError error
		expectName  string
	}{
		{
			name: "get existing model",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				return storage
			},
			id:          "model1",
			expectError: nil,
			expectName:  "model1",
		},
		{
			name: "get non-existent model",
			setup: func() *StorageInMemory[*db.Model] {
				return NewInMemoryStorage[*db.Model]()
			},
			id:          "non-existent",
			expectError: ErrNotFound,
			expectName:  "",
		},
		{
			name: "get from storage with multiple models",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model2"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model3"})
				return storage
			},
			id:          "model2",
			expectError: nil,
			expectName:  "model2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := tt.setup()
			result, err := storage.Get(context.Background(), tt.id)

			if tt.expectError != nil {
				g.Expect(err).To(Equal(tt.expectError))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(result.GetName()).To(Equal(tt.expectName))
			}
		})
	}
}

func TestStorageInMemory_Insert(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		setup       func() *StorageInMemory[*db.Model]
		record      *db.Model
		expectError error
	}{
		{
			name: "insert into empty storage",
			setup: func() *StorageInMemory[*db.Model] {
				return NewInMemoryStorage[*db.Model]()
			},
			record:      &db.Model{Name: "model1"},
			expectError: nil,
		},
		{
			name: "insert duplicate model",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				return storage
			},
			record:      &db.Model{Name: "model1"},
			expectError: ErrAlreadyExists,
		},
		{
			name: "insert multiple different models",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				return storage
			},
			record:      &db.Model{Name: "model2"},
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := tt.setup()
			err := storage.Insert(context.Background(), tt.record)

			if tt.expectError != nil {
				g.Expect(err).To(Equal(tt.expectError))
			} else {
				g.Expect(err).To(BeNil())
				// Verify it was inserted
				retrieved, getErr := storage.Get(context.Background(), tt.record.GetName())
				g.Expect(getErr).To(BeNil())
				g.Expect(retrieved.GetName()).To(Equal(tt.record.GetName()))
			}
		})
	}
}

func TestStorageInMemory_List(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name          string
		setup         func() *StorageInMemory[*db.Model]
		expectedLen   int
		expectedNames []string
	}{
		{
			name: "list empty storage",
			setup: func() *StorageInMemory[*db.Model] {
				return NewInMemoryStorage[*db.Model]()
			},
			expectedLen:   0,
			expectedNames: []string{},
		},
		{
			name: "list single model",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				return storage
			},
			expectedLen:   1,
			expectedNames: []string{"model1"},
		},
		{
			name: "list multiple models",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model2"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model3"})
				return storage
			},
			expectedLen:   3,
			expectedNames: []string{"model1", "model2", "model3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := tt.setup()
			results, err := storage.List(context.Background())

			g.Expect(err).To(BeNil())
			g.Expect(results).To(HaveLen(tt.expectedLen))

			if tt.expectedLen > 0 {
				names := make([]string, len(results))
				for i, r := range results {
					names[i] = r.GetName()
				}
				g.Expect(names).To(ConsistOf(tt.expectedNames))
			}
		})
	}
}

func TestStorageInMemory_Update(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		setup       func() *StorageInMemory[*db.Model]
		record      *db.Model
		expectError error
	}{
		{
			name: "update existing model",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{
					Name:     "model1",
					Versions: []*db.ModelVersion{{Version: 1}},
				})
				return storage
			},
			record: &db.Model{
				Name:     "model1",
				Versions: []*db.ModelVersion{{Version: 2}},
			},
			expectError: nil,
		},
		{
			name: "update non-existent model",
			setup: func() *StorageInMemory[*db.Model] {
				return NewInMemoryStorage[*db.Model]()
			},
			record:      &db.Model{Name: "non-existent"},
			expectError: ErrNotFound,
		},
		{
			name: "update one of multiple models",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model2"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model3"})
				return storage
			},
			record:      &db.Model{Name: "model2", Deleted: true},
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := tt.setup()
			err := storage.Update(context.Background(), tt.record)

			if tt.expectError != nil {
				g.Expect(err).To(Equal(tt.expectError))
			} else {
				g.Expect(err).To(BeNil())
				// Verify the update
				retrieved, getErr := storage.Get(context.Background(), tt.record.GetName())
				g.Expect(getErr).To(BeNil())
				g.Expect(retrieved.GetName()).To(Equal(tt.record.GetName()))
				if len(tt.record.Versions) > 0 {
					g.Expect(retrieved.Versions).To(HaveLen(len(tt.record.Versions)))
					g.Expect(retrieved.Versions[0].Version).To(Equal(tt.record.Versions[0].Version))
				}
			}
		})
	}
}

func TestStorageInMemory_Delete(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name           string
		setup          func() *StorageInMemory[*db.Model]
		id             string
		expectError    error
		remainingCount int
	}{
		{
			name: "delete existing model",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				return storage
			},
			id:             "model1",
			expectError:    nil,
			remainingCount: 0,
		},
		{
			name: "delete non-existent model",
			setup: func() *StorageInMemory[*db.Model] {
				return NewInMemoryStorage[*db.Model]()
			},
			id:             "non-existent",
			expectError:    ErrNotFound,
			remainingCount: 0,
		},
		{
			name: "delete one of multiple models",
			setup: func() *StorageInMemory[*db.Model] {
				storage := NewInMemoryStorage[*db.Model]()
				_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model2"})
				_ = storage.Insert(context.Background(), &db.Model{Name: "model3"})
				return storage
			},
			id:             "model2",
			expectError:    nil,
			remainingCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := tt.setup()
			err := storage.Delete(context.Background(), tt.id)

			if tt.expectError != nil {
				g.Expect(err).To(Equal(tt.expectError))
			} else {
				g.Expect(err).To(BeNil())
				// Verify deletion
				_, getErr := storage.Get(context.Background(), tt.id)
				g.Expect(getErr).To(Equal(ErrNotFound))

				// Check remaining count
				list, _ := storage.List(context.Background())
				g.Expect(list).To(HaveLen(tt.remainingCount))
			}
		})
	}
}

func TestStorageInMemory_ServerType(t *testing.T) {
	g := NewWithT(t)

	t.Run("operations with Server type", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Server]()

		// Insert
		server1 := &db.Server{Name: "server1", Shared: true}
		err := storage.Insert(context.Background(), server1)
		g.Expect(err).To(BeNil())

		// Get
		retrieved, err := storage.Get(context.Background(), "server1")
		g.Expect(err).To(BeNil())
		g.Expect(retrieved.GetName()).To(Equal("server1"))
		g.Expect(retrieved.Shared).To(Equal(true))

		// Update
		server1.Shared = false
		err = storage.Update(context.Background(), server1)
		g.Expect(err).To(BeNil())

		retrieved, err = storage.Get(context.Background(), "server1")
		g.Expect(err).To(BeNil())
		g.Expect(retrieved.Shared).To(Equal(false))

		// Delete
		err = storage.Delete(context.Background(), "server1")
		g.Expect(err).To(BeNil())

		_, err = storage.Get(context.Background(), "server1")
		g.Expect(err).To(Equal(ErrNotFound))
	})
}

func TestStorageInMemory_Cloning(t *testing.T) {
	g := NewWithT(t)

	t.Run("modifications to returned record don't affect storage", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Model]()
		original := &db.Model{
			Name:     "model1",
			Versions: []*db.ModelVersion{{Version: 1}},
		}
		err := storage.Insert(context.Background(), original)
		g.Expect(err).To(BeNil())

		// Get the model
		retrieved, err := storage.Get(context.Background(), "model1")
		g.Expect(err).To(BeNil())

		// Modify the retrieved model
		retrieved.Deleted = true
		retrieved.Versions = append(retrieved.Versions, &db.ModelVersion{Version: 2})

		// Get again and verify storage wasn't affected
		retrievedAgain, err := storage.Get(context.Background(), "model1")
		g.Expect(err).To(BeNil())
		g.Expect(retrievedAgain.Deleted).To(BeFalse())
		g.Expect(retrievedAgain.Versions).To(HaveLen(1))
		g.Expect(retrievedAgain.Versions[0].Version).To(Equal(uint32(1)))
	})

	t.Run("modifications to inserted record don't affect storage", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Model]()
		model := &db.Model{
			Name:     "model1",
			Versions: []*db.ModelVersion{{Version: 1}},
		}
		err := storage.Insert(context.Background(), model)
		g.Expect(err).To(BeNil())

		// Modify the original model
		model.Deleted = true
		model.Versions = append(model.Versions, &db.ModelVersion{Version: 2})

		// Get from storage and verify it wasn't affected
		retrieved, err := storage.Get(context.Background(), "model1")
		g.Expect(err).To(BeNil())
		g.Expect(retrieved.Deleted).To(BeFalse())
		g.Expect(retrieved.Versions).To(HaveLen(1))
		g.Expect(retrieved.Versions[0].Version).To(Equal(uint32(1)))
	})
}

func TestStorageInMemory_ConcurrentOperations(t *testing.T) {
	g := NewWithT(t)

	t.Run("concurrent inserts", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Model]()
		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				model := &db.Model{Name: "model" + string(rune('0'+idx))}
				_ = storage.Insert(context.Background(), model)
			}(i)
		}

		wg.Wait()

		list, err := storage.List(context.Background())
		g.Expect(err).To(BeNil())
		g.Expect(len(list)).To(BeNumerically("<=", numGoroutines))
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Model]()
		_ = storage.Insert(context.Background(), &db.Model{Name: "model1"})

		var wg sync.WaitGroup
		numReaders := 5
		numWriters := 5

		// Concurrent readers
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = storage.Get(context.Background(), "model1")
			}()
		}

		// Concurrent writers
		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = storage.Update(context.Background(), &db.Model{Name: "model1"})
			}()
		}

		wg.Wait()

		// Verify storage is still consistent
		retrieved, err := storage.Get(context.Background(), "model1")
		g.Expect(err).To(BeNil())
		g.Expect(retrieved.GetName()).To(Equal("model1"))
	})

	t.Run("concurrent list operations", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Model]()
		for i := 0; i < 5; i++ {
			_ = storage.Insert(context.Background(), &db.Model{Name: "model" + string(rune('0'+i))})
		}

		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				list, err := storage.List(context.Background())
				g.Expect(err).To(BeNil())
				g.Expect(len(list)).To(BeNumerically(">=", 0))
			}()
		}

		wg.Wait()
	})
}

func TestStorageInMemory_Integration(t *testing.T) {
	g := NewWithT(t)

	t.Run("full lifecycle", func(t *testing.T) {
		storage := NewInMemoryStorage[*db.Model]()

		// Insert multiple models
		for i := 1; i <= 3; i++ {
			model := &db.Model{
				Name:     "model" + string(rune('0'+i)),
				Versions: []*db.ModelVersion{{Version: uint32(i)}},
			}
			err := storage.Insert(context.Background(), model)
			g.Expect(err).To(BeNil())
		}

		// List all
		list, err := storage.List(context.Background())
		g.Expect(err).To(BeNil())
		g.Expect(list).To(HaveLen(3))

		// Update one
		model2, err := storage.Get(context.Background(), "model2")
		g.Expect(err).To(BeNil())
		model2.Deleted = true
		err = storage.Update(context.Background(), model2)
		g.Expect(err).To(BeNil())

		// Verify update
		updated, err := storage.Get(context.Background(), "model2")
		g.Expect(err).To(BeNil())
		g.Expect(updated.Deleted).To(BeTrue())

		// Delete one
		err = storage.Delete(context.Background(), "model1")
		g.Expect(err).To(BeNil())

		// Verify deletion
		list, err = storage.List(context.Background())
		g.Expect(err).To(BeNil())
		g.Expect(list).To(HaveLen(2))

		// Try to get deleted model
		_, err = storage.Get(context.Background(), "model1")
		g.Expect(err).To(Equal(ErrNotFound))
	})
}
