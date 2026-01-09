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
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
)

type Storage[T proto.Message] interface {
	Get(ctx context.Context, name string) (T, error)
	Insert(ctx context.Context, record T) error
	List(ctx context.Context) ([]T, error)
	Update(ctx context.Context, record T) error
	Delete(ctx context.Context, name string) error
}
