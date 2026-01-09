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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"google.golang.org/protobuf/proto"
)

type StorageInMemory[T interface {
	proto.Message
	GetName() string
}] struct {
	mu      sync.RWMutex
	records map[string]T
}

var _ Storage[*db.Model] = &StorageInMemory[*db.Model]{}
var _ Storage[*db.Server] = &StorageInMemory[*db.Server]{}

func NewInMemoryStorage[T interface {
	proto.Message
	GetName() string
}]() *StorageInMemory[T] {
	return &StorageInMemory[T]{
		records: make(map[string]T),
	}
}

func (s *StorageInMemory[T]) Get(_ context.Context, id string) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[id]
	if !ok {
		return *new(T), ErrNotFound
	}

	return proto.Clone(record).(T), nil
}

func (s *StorageInMemory[T]) Insert(_ context.Context, record T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[record.GetName()]; ok {
		return ErrAlreadyExists
	}

	s.records[record.GetName()] = proto.Clone(record).(T)
	return nil
}

func (s *StorageInMemory[T]) List(_ context.Context) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]T, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, proto.Clone(record).(T))
	}

	return records, nil
}

func (s *StorageInMemory[T]) Update(_ context.Context, record T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[record.GetName()]; !ok {
		return ErrNotFound
	}

	s.records[record.GetName()] = proto.Clone(record).(T)
	return nil
}

func (s *StorageInMemory[T]) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[id]; !ok {
		return ErrNotFound
	}

	delete(s.records, id)
	return nil
}
