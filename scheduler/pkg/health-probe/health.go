package health_probe

import (
	"fmt"
	"slices"
)

type ProbeType int

const (
	ProbeStartUp ProbeType = 1 << iota
	ProbeReadiness
	ProbeLiveness
)

//go:generate go tool mockgen -source=health.go -destination=./mocks/mock_health.go -package=mocks Manager
type Manager interface {
	AddCheck(cb ProbeCallback, probes ...ProbeType)
	HasCallbacks(probe ProbeType) bool
	CheckReadiness() error
	CheckLiveness() error
	CheckStartup() error
}

type ProbeCallback func() error

func (p ProbeType) Valid() bool {
	return slices.Contains([]ProbeType{ProbeStartUp, ProbeReadiness, ProbeLiveness}, p)
}

type manager struct {
	svcs map[ProbeType][]ProbeCallback
}

func NewManager() Manager {
	return &manager{
		svcs: make(map[ProbeType][]ProbeCallback, 0),
	}
}

func (m *manager) HasCallbacks(probe ProbeType) bool {
	_, ok := m.svcs[probe]
	return ok
}

func (m *manager) AddCheck(cb ProbeCallback, probes ...ProbeType) {
	if cb == nil {
		panic("nil callback")
	}

	for _, probe := range probes {
		if !probe.Valid() {
			panic(fmt.Sprintf("Invalid probe type %v", probe))
		}
		if _, ok := m.svcs[probe]; !ok {
			m.svcs[probe] = make([]ProbeCallback, 0)
		}
		m.svcs[probe] = append(m.svcs[probe], cb)
	}
}

func (m *manager) CheckReadiness() error {
	return m.runCheck(ProbeReadiness)
}

func (m *manager) CheckStartup() error {
	return m.runCheck(ProbeStartUp)
}

func (m *manager) CheckLiveness() error {
	return m.runCheck(ProbeLiveness)
}

func (m *manager) runCheck(probe ProbeType) error {
	for _, cb := range m.svcs[probe] {
		if err := cb(); err != nil {
			return fmt.Errorf("failed probe: %w", err)
		}
	}
	return nil
}
