/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package in_memory

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"google.golang.org/protobuf/proto"
)

type Storage[T interface {
	proto.Message
	GetName() string
}] struct {
	mu      sync.RWMutex
	records map[string]T
}

var _ store.Storage[*db.Model] = &Storage[*db.Model]{}
var _ store.Storage[*db.Server] = &Storage[*db.Server]{}

func NewStorage[T interface {
	proto.Message
	GetName() string
}]() *Storage[T] {
	return &Storage[T]{
		records: make(map[string]T),
	}
}

func (s *Storage[T]) Get(id string) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[id]
	if !ok {
		return *new(T), store.ErrNotFound
	}

	return proto.Clone(record).(T), nil
}

func (s *Storage[T]) Insert(record T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[record.GetName()]; ok {
		return fmt.Errorf("record with name %s already exists", record.GetName())
	}

	s.records[record.GetName()] = proto.Clone(record).(T)
	return nil
}

func (s *Storage[T]) List() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]T, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, proto.Clone(record).(T))
	}

	return records, nil
}

func (s *Storage[T]) Update(record T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[record.GetName()]; !ok {
		return store.ErrNotFound
	}

	s.records[record.GetName()] = proto.Clone(record).(T)
	return nil
}

func (s *Storage[T]) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[id]; !ok {
		return store.ErrNotFound
	}

	delete(s.records, id)
	return nil
}
