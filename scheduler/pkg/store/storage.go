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
