package in_memory

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"google.golang.org/protobuf/proto"
)

type Storage struct {
	mu      sync.RWMutex
	servers map[string]*db.Server
	models  map[string]*db.Model
}

var _ store.Storage = &Storage{}

func NewStorage() *Storage {
	return &Storage{
		servers: make(map[string]*db.Server),
		models:  make(map[string]*db.Model),
	}
}

// GetModel retrieves a model by name, returning a clone to prevent external modification
func (s *Storage) GetModel(name string) (*db.Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	model, ok := s.models[name]
	if !ok {
		return nil, store.ErrNotFound
	}

	return proto.Clone(model).(*db.Model), nil
}

// AddModel adds a new model, storing a clone to prevent external modification
func (s *Storage) AddModel(model *db.Model) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.models[model.Name]; ok {
		return fmt.Errorf("model with name %s already exists", model.Name)
	}

	s.models[model.Name] = proto.Clone(model).(*db.Model)
	return nil
}

// ListModels returns all models, cloning each to prevent external modification
func (s *Storage) ListModels() ([]*db.Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	models := make([]*db.Model, 0, len(s.models))
	for _, model := range s.models {
		models = append(models, proto.Clone(model).(*db.Model))
	}

	return models, nil
}

// UpdateModel updates an existing model, storing a clone to prevent external modification
func (s *Storage) UpdateModel(model *db.Model) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.models[model.Name]; !ok {
		return store.ErrNotFound
	}

	s.models[model.Name] = proto.Clone(model).(*db.Model)
	return nil
}

// DeleteModel removes a model by name
func (s *Storage) DeleteModel(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.models[name]; !ok {
		return store.ErrNotFound
	}

	delete(s.models, name)
	return nil
}

// GetServer retrieves a server by name, returning a clone to prevent external modification
func (s *Storage) GetServer(name string) (*db.Server, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, ok := s.servers[name]
	if !ok {
		return nil, store.ErrNotFound
	}

	return proto.Clone(server).(*db.Server), nil
}

// AddServer adds a new server, storing a clone to prevent external modification
func (s *Storage) AddServer(server *db.Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.servers[server.Name]; ok {
		return fmt.Errorf("server with name %s already exists", server.Name)
	}

	s.servers[server.Name] = proto.Clone(server).(*db.Server)
	return nil
}

// ListServers returns all servers, cloning each to prevent external modification
func (s *Storage) ListServers() ([]*db.Server, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	servers := make([]*db.Server, 0, len(s.servers))
	for _, server := range s.servers {
		servers = append(servers, proto.Clone(server).(*db.Server))
	}

	return servers, nil
}

// UpdateServer updates an existing server, storing a clone to prevent external modification
func (s *Storage) UpdateServer(server *db.Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.servers[server.Name]; !ok {
		return store.ErrNotFound
	}

	s.servers[server.Name] = proto.Clone(server).(*db.Server)
	return nil
}

// DeleteServer removes a server by name
func (s *Storage) DeleteServer(serverName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.servers[serverName]; !ok {
		return store.ErrNotFound
	}

	delete(s.servers, serverName)
	return nil
}
