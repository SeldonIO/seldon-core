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
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

var (
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	GetModel(name string) (*db.Model, error)
	AddModel(model *db.Model) error
	ListModels() ([]*db.Model, error)
	UpdateModel(model *db.Model) error
	DeleteModel(name string) error
	GetServer(name string) (*db.Server, error)
	AddServer(server *db.Server) error
	ListServers() ([]*db.Server, error)
	UpdateServer(server *db.Server) error
	DeleteServer(serverName string) error
}
