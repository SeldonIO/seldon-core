/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	GetModel(name string) (*Model, error)
	AddModel(model *Model) error
	ListModels() ([]*Model, error)
	UpdateModel(model *Model) error
	DeleteModel(name string) error
	GetServer(name string) (*Server, error)
	AddServer(server *Server) error
	ListServers() ([]*Server, error)
	UpdateServer(server *Server) error
	DeleteServer(serverName string) error
}
