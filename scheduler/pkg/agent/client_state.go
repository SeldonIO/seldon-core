package agent

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
)

type ClientState struct {
	mu                   sync.RWMutex
	replicaConfig        *agent.ReplicaConfig
	loadedModels         map[string]*ModelVersions
	availableMemoryBytes uint64
}

func NewClientState(replicaConfig *agent.ReplicaConfig) *ClientState {
	return &ClientState{
		replicaConfig:        replicaConfig,
		loadedModels:         make(map[string]*ModelVersions),
		availableMemoryBytes: replicaConfig.MemoryBytes,
	}
}

func (c *ClientState) addModelVersion(req *agent.ModelVersion) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	modelName := req.GetModel().GetMeta().GetName()
	mv := c.loadedModels[modelName]
	if mv == nil {
		mv = NewModelVersions()
		c.loadedModels[modelName] = mv
	}
	if mv.getModelVersion(req) != nil {
		return fmt.Errorf("Model version already exists for %s:%d", modelName, req.GetVersion())
	}
	c.availableMemoryBytes += mv.totalMemoryBytes // remove existing before recalculating later
	mv.addModelVersion(req)
	var err error
	if mv.totalMemoryBytes > c.availableMemoryBytes {
		mv.removeModelVersion(req)
		err = fmt.Errorf("Model version exceeds available memory %s:%d this version requested %d bytes which added to extant versions is %d bytes but we only have available on this replica %d bytes",
			modelName, req.GetVersion(), req.GetModel().GetModelSpec().GetMemoryBytes(), mv.totalMemoryBytes, c.availableMemoryBytes)
	}
	c.availableMemoryBytes -= mv.totalMemoryBytes // reset available memory
	return err
}

func (c *ClientState) removeModelVersion(req *agent.ModelVersion) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.removeModelVersionImpl(req)
}

// Remove model version and return true if no versions left (in which case we remove from map)
func (c *ClientState) removeModelVersionImpl(req *agent.ModelVersion) bool {

	modelName := req.GetModel().GetMeta().GetName()
	mv := c.loadedModels[modelName]
	if mv == nil {
		return true
	}
	c.availableMemoryBytes += mv.totalMemoryBytes // remove existing before recalculcating later
	mv.removeModelVersion(req)
	c.availableMemoryBytes -= mv.totalMemoryBytes // reset available memory
	if mv.numVersions() == 0 {
		delete(c.loadedModels, modelName)
		return true
	}
	return false
}

type ModelVersions struct {
	versions         map[uint32]*agent.ModelVersion
	totalMemoryBytes uint64
}

func NewModelVersions() *ModelVersions {
	return &ModelVersions{
		versions: make(map[uint32]*agent.ModelVersion),
	}
}

func (m *ModelVersions) numVersions() int {
	return len(m.versions)
}

func (m *ModelVersions) getTotalMemory() uint64 {
	var total uint64
	for _, mv := range m.versions {
		total += mv.GetModel().GetModelSpec().GetMemoryBytes()
	}
	return total
}

func (m *ModelVersions) getModelVersion(md *agent.ModelVersion) *agent.ModelVersion {
	return m.versions[md.GetVersion()]
}

func (m *ModelVersions) addModelVersion(md *agent.ModelVersion) {
	m.versions[md.Version] = md
	m.totalMemoryBytes = m.getTotalMemory()
}

func (m *ModelVersions) removeModelVersion(md *agent.ModelVersion) {
	if _, ok := m.versions[md.Version]; !ok {
		return
	}
	delete(m.versions, md.Version)
	m.totalMemoryBytes = m.getTotalMemory()
}
