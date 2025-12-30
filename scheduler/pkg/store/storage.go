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
	"google.golang.org/protobuf/proto"
)

var (
	ErrNotFound = errors.New("record not found")
)

type Storage[T proto.Message] interface {
	Get(name string) (T, error)
	Insert(record T) error
	List() ([]T, error)
	Update(record T) error
	Delete(name string) error
}
