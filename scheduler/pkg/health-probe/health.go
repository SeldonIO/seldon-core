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

type Manager interface {
	RegisterSvc(id string, cb ProbeCallback, probes ...ProbeType)
	CheckReadiness() error
	CheckLiveness() error
	CheckStartup() error
}

type ProbeCallback func() error

func (p ProbeType) Valid() bool {
	return slices.Contains([]ProbeType{ProbeStartUp, ProbeReadiness, ProbeLiveness}, p)
}

type manager struct {
	svcs map[string]service
}

type service struct {
	probes ProbeType
	cb     ProbeCallback
}

func NewManager() Manager {
	return &manager{
		svcs: make(map[string]service),
	}
}

func (m *manager) RegisterSvc(id string, cb ProbeCallback, probes ...ProbeType) {
	if _, ok := m.svcs[id]; ok {
		panic(fmt.Sprintf("Service %s already added", id))
	}

	var probesSum ProbeType
	for _, probe := range probes {
		if !probe.Valid() {
			panic(fmt.Sprintf("Invalid probe type %v", probe))
		}
		probesSum |= probe
	}

	m.svcs[id] = service{
		probes: probesSum,
		cb:     cb,
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
	for id, svc := range m.svcs {
		if svc.probes&probe == 0 {
			continue
		}

		if svc.cb == nil {
			return fmt.Errorf("service %s callback not set", id)
		}
		if err := svc.cb(); err != nil {
			return fmt.Errorf("service %s is not ready: %w", id, err)
		}
	}
	return nil
}
